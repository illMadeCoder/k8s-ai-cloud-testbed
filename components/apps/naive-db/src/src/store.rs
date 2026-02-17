use std::fs::{self, File, OpenOptions};
use std::io::{self, Seek, SeekFrom, Write};
use std::os::unix::fs::FileExt;
use std::path::Path;
use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::Mutex;
use std::time::Instant;

pub struct WriteResult {
    pub row_id: u64,
    pub fsync_secs: f64,
}

pub struct FileStore {
    writer: Mutex<File>,
    reader: File,
    row_count: AtomicU64,
}

impl FileStore {
    pub fn open_or_create(data_dir: &str) -> io::Result<Self> {
        fs::create_dir_all(data_dir)?;
        let path = Path::new(data_dir).join("naive.db");

        let mut writer = OpenOptions::new()
            .create(true)
            .read(true)
            .write(true)
            .open(&path)?;

        // Truncate partial writes (each row is 4 bytes)
        let len = writer.metadata()?.len();
        let valid_len = len - (len % 4);
        if valid_len != len {
            writer.set_len(valid_len)?;
            writer.seek(SeekFrom::End(0))?;
        }

        let row_count = valid_len / 4;
        let reader = File::open(&path)?;

        Ok(Self {
            writer: Mutex::new(writer),
            reader,
            row_count: AtomicU64::new(row_count),
        })
    }

    pub fn write(&self, value: i32) -> io::Result<WriteResult> {
        let mut file = self.writer.lock().unwrap();
        file.seek(SeekFrom::End(0))?;
        file.write_all(&value.to_le_bytes())?;

        let fsync_start = Instant::now();
        file.sync_all()?;
        let fsync_secs = fsync_start.elapsed().as_secs_f64();

        let row_id = self.row_count.fetch_add(1, Ordering::SeqCst);
        Ok(WriteResult { row_id, fsync_secs })
    }

    pub fn read(&self, row_id: u64) -> Option<i32> {
        if row_id >= self.row_count.load(Ordering::SeqCst) {
            return None;
        }
        let mut buf = [0u8; 4];
        self.reader.read_at(&mut buf, row_id * 4).ok()?;
        Some(i32::from_le_bytes(buf))
    }

    pub fn row_count(&self) -> u64 {
        self.row_count.load(Ordering::SeqCst)
    }
}
