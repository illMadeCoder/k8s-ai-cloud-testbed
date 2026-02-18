import { describe, expect, it } from 'vitest';
import { formatDuration, formatBytes, formatDate, formatCost, formatValue } from './format';

describe('formatDuration', () => {
  it('formats zero', () => {
    expect(formatDuration(0)).toBe('0s');
  });

  it('formats sub-minute values', () => {
    expect(formatDuration(0.5)).toBe('1s');
    expect(formatDuration(30)).toBe('30s');
    expect(formatDuration(59)).toBe('59s');
  });

  it('formats minute boundary', () => {
    expect(formatDuration(60)).toBe('1m');
  });

  it('formats sub-hour values', () => {
    expect(formatDuration(120)).toBe('2m');
    expect(formatDuration(3599)).toBe('60m');
  });

  it('formats hour boundary', () => {
    expect(formatDuration(3600)).toBe('1h');
  });

  it('formats hours and minutes', () => {
    expect(formatDuration(7261)).toBe('2h 1m');
    expect(formatDuration(7200)).toBe('2h');
  });
});

describe('formatBytes', () => {
  it('formats zero', () => {
    expect(formatBytes(0)).toBe('0 B');
  });

  it('formats bytes', () => {
    expect(formatBytes(1)).toBe('1 B');
    expect(formatBytes(1023)).toBe('1023 B');
  });

  it('formats KiB boundary', () => {
    expect(formatBytes(1024)).toBe('1.0 KiB');
  });

  it('formats KiB values', () => {
    expect(formatBytes(1536)).toBe('1.5 KiB');
  });

  it('formats MiB boundary', () => {
    expect(formatBytes(1048576)).toBe('1.0 MiB');
  });

  it('formats GiB boundary', () => {
    expect(formatBytes(1073741824)).toBe('1.00 GiB');
  });
});

describe('formatDate', () => {
  it('formats ISO date string', () => {
    expect(formatDate('2026-01-15T12:00:00Z')).toBe('Jan 15, 2026');
  });

  it('formats another date', () => {
    expect(formatDate('2025-12-31T12:00:00Z')).toBe('Dec 31, 2025');
  });
});

describe('formatCost', () => {
  it('formats zero', () => {
    expect(formatCost(0)).toBe('$0.00');
  });

  it('rounds small values', () => {
    expect(formatCost(0.001)).toBe('$0.00');
  });

  it('formats typical values', () => {
    expect(formatCost(1.5)).toBe('$1.50');
    expect(formatCost(99.99)).toBe('$99.99');
  });

  it('formats large values', () => {
    expect(formatCost(1234.56)).toBe('$1234.56');
  });
});

describe('formatValue', () => {
  describe('seconds unit', () => {
    it('converts sub-10ms values to milliseconds with 2 decimal places', () => {
      expect(formatValue(0.001, 'seconds')).toBe('1.00ms');
      expect(formatValue(0.005, 'seconds')).toBe('5.00ms');
      expect(formatValue(0.003, 'seconds')).toBe('3.00ms');
      expect(formatValue(0.0099, 'seconds')).toBe('9.90ms');
    });

    it('converts sub-second values to milliseconds with 1 decimal place', () => {
      expect(formatValue(0.01, 'seconds')).toBe('10.0ms');
      expect(formatValue(0.1, 'seconds')).toBe('100.0ms');
      expect(formatValue(0.5, 'seconds')).toBe('500.0ms');
      expect(formatValue(0.99, 'seconds')).toBe('990.0ms');
    });

    it('formats values >= 1s as seconds', () => {
      expect(formatValue(1.0, 'seconds')).toBe('1.00s');
      expect(formatValue(60.5, 'seconds')).toBe('60.50s');
    });

    it('never displays non-zero values as 0.00s', () => {
      // This is the key regression test for the bug fixed in commit 6269623
      const smallValues = [0.001, 0.002, 0.003, 0.004, 0.005, 0.006, 0.007, 0.008, 0.009];
      for (const v of smallValues) {
        const formatted = formatValue(v, 'seconds');
        expect(formatted).not.toBe('0.00s');
        expect(formatted).toMatch(/ms$/);
      }
    });
  });

  describe('bytes unit', () => {
    it('delegates to formatBytes', () => {
      expect(formatValue(0, 'bytes')).toBe('0 B');
      expect(formatValue(1024, 'bytes')).toBe('1.0 KiB');
      expect(formatValue(1048576, 'bytes')).toBe('1.0 MiB');
    });
  });

  describe('percentage unit', () => {
    it('formats percentage values', () => {
      expect(formatValue(0, '%')).toBe('0.0%');
      expect(formatValue(50, '%')).toBe('50.0%');
      expect(formatValue(99.9, '%')).toBe('99.9%');
      expect(formatValue(100, '%')).toBe('100.0%');
    });
  });

  describe('cores unit', () => {
    it('formats CPU cores', () => {
      expect(formatValue(0.5, 'cores')).toBe('0.500 cores');
      expect(formatValue(2, 'cores')).toBe('2.000 cores');
    });
  });

  describe('req/s unit', () => {
    it('formats request rate', () => {
      expect(formatValue(150, 'req/s')).toBe('150 req/s');
      expect(formatValue(0, 'req/s')).toBe('0 req/s');
    });
  });

  describe('ms unit', () => {
    it('formats milliseconds', () => {
      expect(formatValue(42.56, 'ms')).toBe('42.6 ms');
      expect(formatValue(0, 'ms')).toBe('0.0 ms');
    });
  });

  describe('no unit', () => {
    it('formats as plain number with 2 decimals', () => {
      expect(formatValue(42, undefined)).toBe('42.00');
      expect(formatValue(0.123, undefined)).toBe('0.12');
    });
  });

  describe('unknown unit', () => {
    it('formats with unit suffix', () => {
      expect(formatValue(42, 'widgets')).toBe('42.00 widgets');
    });
  });
});
