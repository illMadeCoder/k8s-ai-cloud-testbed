import fs from 'node:fs';
import path from 'node:path';

export interface Series {
  id: string;
  name: string;
  description: string;
  order?: string[];
}

interface SeriesData {
  series: Series[];
}

const seriesPath = path.resolve(process.cwd(), 'data', '_series.json');

let cached: SeriesData | null = null;

function loadSeriesData(): SeriesData {
  if (cached) return cached;
  const raw = fs.readFileSync(seriesPath, 'utf-8');
  cached = JSON.parse(raw) as SeriesData;
  return cached;
}

export function loadSeries(): Series[] {
  return loadSeriesData().series;
}

export function getSeriesMeta(seriesId: string): Series | undefined {
  return loadSeries().find((s) => s.id === seriesId);
}

export function getAllSeriesIds(): string[] {
  return loadSeries().map((s) => s.id);
}

export function getSeriesOrder(seriesId: string): string[] {
  return getSeriesMeta(seriesId)?.order ?? [];
}
