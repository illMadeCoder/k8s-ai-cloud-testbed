/** Mirrors Go structs in operators/experiment-operator/internal/metrics/collector.go */

export interface ExperimentSummary {
  name: string;
  namespace: string;
  description: string;
  createdAt: string;
  completedAt: string;
  durationSeconds: number;
  phase: string;
  tags?: string[];
  targets: TargetSummary[];
  workflow: WorkflowSummary;
  metrics?: MetricsResult;
  costEstimate?: CostEstimate;
}

export interface TargetSummary {
  name: string;
  clusterName?: string;
  clusterType: string;
  machineType?: string;
  nodeCount?: number;
  components?: string[];
}

export interface WorkflowSummary {
  name: string;
  template: string;
  phase: string;
  startedAt?: string;
  finishedAt?: string;
}

export interface MetricsResult {
  collectedAt: string;
  timeRange: TimeRange;
  queries: Record<string, QueryResult>;
}

export interface TimeRange {
  start: string;
  end: string;
  duration: string;
  stepSeconds: number;
}

export interface QueryResult {
  query: string;
  type: 'instant' | 'range';
  unit?: string;
  description?: string;
  error?: string;
  data?: DataPoint[];
}

export interface DataPoint {
  labels?: Record<string, string>;
  timestamp: string;
  value: number;
}

export interface CostEstimate {
  totalUSD: number;
  durationHours: number;
  perTarget?: Record<string, number>;
  note: string;
}
