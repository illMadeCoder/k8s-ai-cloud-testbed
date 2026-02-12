import type { QueryResult, DataPoint } from '../types';

interface VegaLiteSpec {
  $schema: string;
  width: string;
  height: number;
  title?: string;
  data: { values: VegaDataPoint[] };
  mark: unknown;
  encoding: unknown;
}

interface VegaDataPoint {
  series: string;
  timestamp?: string;
  value: number;
}

const HARNESS_POD_PREFIXES = ['alloy-', 'ts-vm-hub-', 'ts-vm-'];

function isHarnessPod(label: string): boolean {
  return HARNESS_POD_PREFIXES.some((p) => label.startsWith(p));
}

function extractSeriesLabel(dp: DataPoint): string {
  if (!dp.labels || Object.keys(dp.labels).length === 0) return 'Total';
  return dp.labels['target'] ?? dp.labels['pod'] ?? dp.labels['instance']
    ?? dp.labels['scope'] ?? Object.values(dp.labels)[0] ?? 'Total';
}

function transformValue(value: number, unit?: string): number {
  switch (unit) {
    case 'bytes':
      return value / (1024 * 1024); // → MiB
    case 'seconds':
      return value * 1000; // → ms
    default:
      return value;
  }
}

function displayUnit(unit?: string): string {
  switch (unit) {
    case 'bytes':
      return 'MiB';
    case 'seconds':
      return 'ms';
    default:
      return unit ?? '';
  }
}

export function buildBarSpec(name: string, qr: QueryResult): VegaLiteSpec {
  const values: VegaDataPoint[] = (qr.data ?? [])
    .filter((dp) => !isHarnessPod(extractSeriesLabel(dp)))
    .map((dp) => ({
      series: extractSeriesLabel(dp),
      value: transformValue(dp.value, qr.unit),
    }));

  const yUnit = displayUnit(qr.unit);
  const yTitle = yUnit ? `${qr.description ?? name} (${yUnit})` : (qr.description ?? name);

  const seriesCount = values.length;
  const hasLongLabels = values.some((v) => v.series.length > 15);
  const needsAngle = seriesCount > 3 || hasLongLabels;
  const height = Math.min(350, Math.max(250, 200 + seriesCount * 40));

  return {
    $schema: 'https://vega.github.io/schema/vega-lite/v5.json',
    width: 'container',
    height,
    data: { values },
    mark: { type: 'bar', tooltip: true, cornerRadiusEnd: 4 },
    encoding: {
      x: {
        field: 'series', type: 'nominal', title: null,
        sort: { field: 'value', order: 'descending' as const },
        axis: { labelAngle: needsAngle ? -45 : 0, labelLimit: 200 },
      },
      y: { field: 'value', type: 'quantitative', title: yTitle },
    },
  };
}

export function buildLineSpec(name: string, qr: QueryResult): VegaLiteSpec {
  const values: VegaDataPoint[] = (qr.data ?? [])
    .filter((dp) => !isHarnessPod(extractSeriesLabel(dp)))
    .map((dp) => ({
      series: extractSeriesLabel(dp),
      timestamp: dp.timestamp,
      value: transformValue(dp.value, qr.unit),
    }));

  const yUnit = displayUnit(qr.unit);
  const yTitle = yUnit ? `${qr.description ?? name} (${yUnit})` : (qr.description ?? name);

  return {
    $schema: 'https://vega.github.io/schema/vega-lite/v5.json',
    width: 'container',
    height: 300,
    data: { values },
    mark: { type: 'line', tooltip: true, point: true },
    encoding: {
      x: { field: 'timestamp', type: 'temporal', title: 'Time' },
      y: { field: 'value', type: 'quantitative', title: yTitle },
      color: { field: 'series', type: 'nominal', title: 'Target' },
    },
  };
}

export function buildSpec(name: string, qr: QueryResult): VegaLiteSpec {
  return qr.type === 'range' ? buildLineSpec(name, qr) : buildBarSpec(name, qr);
}
