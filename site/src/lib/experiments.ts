import fs from 'node:fs';
import path from 'node:path';
import type { ExperimentSummary } from '../types';

const dataDir = path.resolve(process.cwd(), 'data');

export function loadAllExperiments(): ExperimentSummary[] {
  if (!fs.existsSync(dataDir)) return [];

  return fs
    .readdirSync(dataDir)
    .filter((f) => f.endsWith('.json'))
    .map((f) => {
      const raw = fs.readFileSync(path.join(dataDir, f), 'utf-8');
      return JSON.parse(raw) as ExperimentSummary;
    })
    .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
}

export function loadExperiment(slug: string): ExperimentSummary | undefined {
  const experiments = loadAllExperiments();
  return experiments.find((e) => e.name === slug);
}

export function getExperimentSlug(exp: ExperimentSummary): string {
  return exp.name;
}
