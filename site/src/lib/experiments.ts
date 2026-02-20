import fs from 'node:fs';
import path from 'node:path';
import type { ExperimentSummary } from '../types';
import { getSeriesOrder } from './series';

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
    .filter((f) => f.endsWith('.json') && !f.startsWith('_'))
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
 * Experiment names follow `{prefix}-{random}` where random is 5 lowercase alphanumeric chars.
 * E.g. "gateway-comparison-qh4rc" → "gateway-comparison"
 */
export function getBaseName(name: string): string {
  const match = name.match(/^(.+)-[a-z0-9]{5}$/);
  return match ? match[1] : name;
}

/** Well-known acronyms to preserve when formatting display names. */
const acronyms = new Set(['tsdb', 'api', 'cpu', 'gpu', 'slo', 'sli', 'http', 'grpc', 'dns', 'tls', 'gke', 'k8s', 'vm', 'ci', 'cd']);

/**
 * Format a base name for human display.
 * E.g. "tsdb-comparison" → "TSDB Comparison", "gateway-comparison" → "Gateway Comparison"
 */
export function formatDisplayName(baseName: string): string {
  return baseName
    .split('-')
    .map((word) => acronyms.has(word) ? word.toUpperCase() : word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}

/**
 * Get a human-readable display title for an experiment group.
 * Prefers the explicit `title` field from any run (latest first),
 * falling back to formatting the base name.
 */
export function getDisplayTitle(group: ExperimentGroup): string {
  for (const run of group.runs) {
    if (run.title) return run.title;
  }
  return formatDisplayName(group.baseName);
}

/**
 * Get a display title for a single experiment run.
 * Prefers the explicit `title` field, falling back to formatting the base name.
 */
export function getRunDisplayTitle(exp: ExperimentSummary): string {
  if (exp.title) return exp.title;
  return formatDisplayName(getBaseName(exp.name));
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

const domainTagMap: Record<string, string> = {
  metrics: 'observability',
  logging: 'observability',
  tracing: 'observability',
  slos: 'observability',
  cost: 'observability',
  prometheus: 'observability',
  'victoria-metrics': 'observability',
  loki: 'observability',
  tempo: 'observability',
  grafana: 'observability',
  gateways: 'networking',
  ingress: 'networking',
  'service-mesh': 'networking',
  gateway: 'networking',
  envoy: 'networking',
  nginx: 'networking',
  traefik: 'networking',
  'object-storage': 'storage',
  database: 'storage',
  s3: 'storage',
  seaweedfs: 'storage',
  pipelines: 'cicd',
  ci: 'cicd',
  cd: 'cicd',
  'supply-chain': 'cicd',
};

const subdomainTagMap: Record<string, string> = {
  prometheus: 'metrics',
  'victoria-metrics': 'metrics',
  metrics: 'metrics',
  loki: 'logging',
  logging: 'logging',
  tempo: 'tracing',
  tracing: 'tracing',
  slos: 'slos',
  cost: 'cost',
  gateway: 'gateways',
  gateways: 'gateways',
  ingress: 'gateways',
  envoy: 'gateways',
  nginx: 'gateways',
  traefik: 'gateways',
  'object-storage': 'object-storage',
  s3: 'object-storage',
  seaweedfs: 'object-storage',
  database: 'databases',
  pipelines: 'pipelines',
  ci: 'pipelines',
  cd: 'pipelines',
};

const typeTagMap: Record<string, string> = {
  comparison: 'comparison',
  benchmark: 'comparison',
  tutorial: 'tutorial',
  demo: 'demo',
  baseline: 'baseline',
};

/**
 * Derive the domain (observability/networking/storage/cicd/general) from tags.
 */
export function deriveDomain(tags: string[]): string {
  for (const tag of tags) {
    const domain = domainTagMap[tag];
    if (domain) return domain;
  }
  return 'general';
}

/**
 * Derive the subdomain (metrics/logging/gateways/etc.) from tags.
 */
export function deriveSubdomain(tags: string[]): string | undefined {
  for (const tag of tags) {
    const sub = subdomainTagMap[tag];
    if (sub) return sub;
  }
  return undefined;
}

/**
 * Derive the experiment type (comparison/tutorial/demo/baseline) from tags.
 */
export function deriveExperimentType(tags: string[]): string {
  for (const tag of tags) {
    const type = typeTagMap[tag];
    if (type) return type;
  }
  return 'demo';
}

/**
 * Group experiments by derived domain.
 */
export function groupByDomain(experiments: ExperimentSummary[]): Record<string, ExperimentGroup[]> {
  const groups = groupExperiments(experiments);
  const result: Record<string, ExperimentGroup[]> = {};

  for (const group of groups) {
    const domain = deriveDomain(group.tags);
    if (!result[domain]) result[domain] = [];
    result[domain].push(group);
  }

  return result;
}

/**
 * Group experiments by series field.
 */
export function groupBySeries(experiments: ExperimentSummary[]): Record<string, ExperimentGroup[]> {
  const groups = groupExperiments(experiments);
  const result: Record<string, ExperimentGroup[]> = {};

  for (const group of groups) {
    const series = group.latest.series;
    if (!series) continue;
    if (!result[series]) result[series] = [];
    result[series].push(group);
  }

  return result;
}

/**
 * Get sibling experiment groups in the same series, ordered by the series order array.
 * Returns undefined if the experiment is not in a series.
 */
export function getSeriesSiblings(
  baseName: string,
  experiments: ExperimentSummary[],
): ExperimentGroup[] | undefined {
  const groups = groupExperiments(experiments);
  const thisGroup = groups.find((g) => g.baseName === baseName);
  if (!thisGroup) return undefined;

  const seriesId = thisGroup.latest.series;
  if (!seriesId) return undefined;

  const order = getSeriesOrder(seriesId);

  const siblings = groups.filter((g) => g.latest.series === seriesId);

  // Sort by series order array; unordered experiments come after ordered ones chronologically
  siblings.sort((a, b) => {
    const ai = order.indexOf(a.baseName);
    const bi = order.indexOf(b.baseName);
    if (ai !== -1 && bi !== -1) return ai - bi;
    if (ai !== -1) return -1;
    if (bi !== -1) return 1;
    return new Date(b.latest.createdAt).getTime() - new Date(a.latest.createdAt).getTime();
  });

  return siblings;
}

/**
 * Filter to comparison-type experiments only.
 */
export function getComparisons(experiments: ExperimentSummary[]): ExperimentGroup[] {
  const groups = groupExperiments(experiments);
  return groups.filter((g) => deriveExperimentType(g.tags) === 'comparison');
}

/**
 * Get aggregate stats across all experiments.
 */
export function getStats(experiments: ExperimentSummary[]): {
  experimentCount: number;
  domainCount: number;
  componentCount: number;
  totalRuns: number;
  totalCostUSD: number;
} {
  const groups = groupExperiments(experiments);
  const domains = new Set(groups.map((g) => deriveDomain(g.tags)));
  const components = new Set<string>();
  let totalCostUSD = 0;

  for (const exp of experiments) {
    for (const target of exp.targets) {
      for (const comp of target.components ?? []) {
        components.add(comp);
      }
    }
    if (exp.costEstimate?.totalUSD) {
      totalCostUSD += exp.costEstimate.totalUSD;
    }
  }

  return {
    experimentCount: groups.length,
    domainCount: domains.size,
    componentCount: components.size,
    totalRuns: experiments.length,
    totalCostUSD,
  };
}

/** Get hypothesis claim, reading new field with old fallback. */
export function getHypothesisClaim(exp: ExperimentSummary): string | undefined {
  return exp.hypothesis?.claim ?? exp.study?.hypothesis;
}

/** Get experiment questions, reading new field with old fallback. */
export function getQuestions(exp: ExperimentSummary): string[] {
  return exp.hypothesis?.questions ?? exp.study?.questions ?? [];
}

/** Get focus areas, reading new field with old fallback. */
export function getFocus(exp: ExperimentSummary): string[] {
  return exp.hypothesis?.focus ?? exp.study?.focus ?? [];
}

/** Get machine verdict from success criteria evaluation. */
export function getMachineVerdict(exp: ExperimentSummary): string | undefined {
  return exp.hypothesis?.machineVerdict;
}

