import fs from 'node:fs';
import path from 'node:path';

export interface SeriesSource {
  title: string;
  url: string;
  image?: string;   // Path relative to public/ (e.g., "images/series/ddia.jpg")
}

export interface Series {
  id: string;
  name: string;
  description: string;
  body?: string;
  image?: string;
  sources?: SeriesSource[];
  order?: string[];
  dag?: Record<string, string[]>;
  transitions?: Record<string, string>;
  color?: string;
  shape?: string;
}

export interface ResolvedDag {
  parentOf: Record<string, string | null>;  // child → parent (null = root)
  childrenOf: Record<string, string[]>;     // parent → children
  roots: string[];
  allNodes: string[];                        // topological order (BFS from roots)
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
  const meta = getSeriesMeta(seriesId);
  if (meta?.order?.length) return meta.order;
  // Fall back to DAG topological order
  const dag = resolveSeriesDag(seriesId);
  return dag?.allNodes ?? [];
}

/**
 * Resolve a series' DAG structure.
 * - If explicit `dag` exists, use it directly
 * - Else if `order` has 2+ entries, derive a linear chain
 * - Else return null (radial layout / no DAG)
 */
export function resolveSeriesDag(seriesId: string): ResolvedDag | null {
  const meta = getSeriesMeta(seriesId);
  if (!meta) return null;

  let childrenOf: Record<string, string[]>;

  if (meta.dag && Object.keys(meta.dag).length > 0) {
    childrenOf = { ...meta.dag };
  } else if (meta.order && meta.order.length >= 2) {
    // Auto-derive linear chain: order[0]→order[1]→...
    childrenOf = {};
    for (let i = 0; i < meta.order.length - 1; i++) {
      childrenOf[meta.order[i]] = [meta.order[i + 1]];
    }
  } else {
    return null;
  }

  // Build parentOf by inverting childrenOf
  const parentOf: Record<string, string | null> = {};
  const allChildSet = new Set<string>();

  for (const [parent, children] of Object.entries(childrenOf)) {
    if (!(parent in parentOf)) parentOf[parent] = null; // default to root
    for (const child of children) {
      parentOf[child] = parent;
      allChildSet.add(child);
      if (!(child in childrenOf)) childrenOf[child] = [];
    }
  }

  // Roots: keys in childrenOf that never appear as a child
  const roots = Object.keys(childrenOf).filter(k => !allChildSet.has(k));

  // BFS from roots for topological order (with visited set to guard against cycles)
  const allNodes: string[] = [];
  const visited = new Set<string>();
  const queue = [...roots];
  while (queue.length > 0) {
    const node = queue.shift()!;
    if (visited.has(node)) continue;
    visited.add(node);
    allNodes.push(node);
    for (const child of (childrenOf[node] ?? [])) {
      if (!visited.has(child)) queue.push(child);
    }
  }

  return { parentOf, childrenOf, roots, allNodes };
}

/**
 * Walk parentOf from node to root, returning [node, parent, grandparent, ..., root].
 */
export function getAncestorChain(dag: ResolvedDag, baseName: string): string[] {
  const chain: string[] = [];
  let current: string | null = baseName;
  const visited = new Set<string>();
  while (current != null && !visited.has(current)) {
    visited.add(current);
    chain.push(current);
    current = dag.parentOf[current] ?? null;
  }
  return chain;
}
