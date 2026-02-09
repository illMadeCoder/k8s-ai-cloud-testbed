export function formatDuration(seconds: number): string {
  if (seconds < 60) return `${Math.round(seconds)}s`;
  if (seconds < 3600) return `${Math.round(seconds / 60)}m`;
  const h = Math.floor(seconds / 3600);
  const m = Math.round((seconds % 3600) / 60);
  return m > 0 ? `${h}h ${m}m` : `${h}h`;
}

export function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KiB`;
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MiB`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GiB`;
}

export function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

export function formatCost(usd: number): string {
  return `$${usd.toFixed(2)}`;
}

export function formatValue(value: number, unit?: string): string {
  if (!unit) return value.toFixed(2);
  switch (unit) {
    case 'bytes':
      return formatBytes(value);
    case 'cores':
      return `${value.toFixed(3)} cores`;
    case 'req/s':
      return `${value.toFixed(0)} req/s`;
    case 'ms':
      return `${value.toFixed(1)} ms`;
    case 'seconds':
      return `${value.toFixed(2)}s`;
    case '%':
      return `${value.toFixed(1)}%`;
    default:
      return `${value.toFixed(2)} ${unit}`;
  }
}
