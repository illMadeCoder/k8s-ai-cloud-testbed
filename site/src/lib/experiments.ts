import fs from 'node:fs';
import path from 'node:path';
import type { ExperimentSummary } from '../types';

const dataDir = path.resolve(process.cwd(), 'data');

export interface ExperimentGroup {
  baseName: string;
  runs: ExperimentSummary[];
  latest: ExperimentSummary;
  tags: string[];
}

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

/**
 * Strip the random K8s generateName suffix.
 * Experiment names follow `{prefix}-{random}` where random is 6 lowercase alphanumeric chars.
 * E.g. "gateway-comparison-g7h8i9" → "gateway-comparison"
 */
export function getBaseName(name: string): string {
  const match = name.match(/^(.+)-[a-z0-9]{6}$/);
  return match ? match[1] : name;
}

/**
 * Group experiments by base name, sorted by most recent run.
 */
export function groupExperiments(experiments: ExperimentSummary[]): ExperimentGroup[] {
  const map = new Map<string, ExperimentSummary[]>();

  for (const exp of experiments) {
    const base = getBaseName(exp.name);
    const runs = map.get(base) ?? [];
    runs.push(exp);
    map.set(base, runs);
  }

  const groups: ExperimentGroup[] = [];
  for (const [baseName, runs] of map) {
    // Sort runs newest-first
    runs.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());

    // Collect union of all tags
    const tagSet = new Set<string>();
    for (const run of runs) {
      for (const tag of run.tags ?? []) {
        tagSet.add(tag);
      }
    }

    groups.push({
      baseName,
      runs,
      latest: runs[0],
      tags: [...tagSet].sort(),
    });
  }

  // Sort groups by latest run date (newest first)
  groups.sort((a, b) => new Date(b.latest.createdAt).getTime() - new Date(a.latest.createdAt).getTime());

  return groups;
}

/**
 * Get all unique tags across all experiments.
 */
export function getAllTags(experiments: ExperimentSummary[]): string[] {
  const tags = new Set<string>();
  for (const exp of experiments) {
    for (const tag of exp.tags ?? []) {
      tags.add(tag);
    }
  }
  return [...tags].sort();
}
