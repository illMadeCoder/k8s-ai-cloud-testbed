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
  analysis?: AnalysisResult;
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
  source?: string;
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

export interface AnalysisResult {
  // Backward-compatible fields
  summary: string;
  metricInsights: Record<string, string>;
  recommendations?: string[];
  generatedAt: string;
  model: string;

  // Structured analysis sections
  abstract?: string;
  targetAnalysis?: TargetAnalysis;
  performanceAnalysis?: PerformanceAnalysis;
  finopsAnalysis?: FinopsAnalysis;
  secopsAnalysis?: SecopsAnalysis;
  capabilitiesMatrix?: CapabilitiesMatrix;
  body?: AnalysisBody;
  feedback?: AnalysisFeedback;
}

export interface TargetAnalysis {
  overview: string;
  perTarget?: Record<string, string>;
  comparisonToBaseline?: string;
}

export interface PerformanceAnalysis {
  overview: string;
  findings?: string[];
  bottlenecks?: string[];
}

export interface FinopsAnalysis {
  overview: string;
  costDrivers?: string[];
  projection?: string;
  optimizations?: string[];
}

export interface SecopsAnalysis {
  overview: string;
  findings?: string[];
  supplyChain?: string;
}

export interface CapabilitiesMatrix {
  technologies: string[];
  categories: CapabilitiesCategory[];
}

export interface CapabilitiesCategory {
  name: string;
  capabilities: CapabilityEntry[];
}

export interface CapabilityEntry {
  name: string;
  values: Record<string, string>;
}

export interface AnalysisBody {
  methodology: string;
  results: string;
  discussion: string;
}

export interface AnalysisFeedback {
  recommendations?: string[];
  experimentDesign?: string[];
}
