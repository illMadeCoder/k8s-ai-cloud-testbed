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
  project?: string;
  study?: StudyContext;
  analysisConfig?: { sections: string[] };
  hypothesis?: HypothesisContext;
  analyzerConfig?: { sections: string[] };
  targets: TargetSummary[];
  workflow: WorkflowSummary;
  metrics?: MetricsResult;
  costEstimate?: CostEstimate;
  analysis?: AnalysisResult;
}

export interface StudyContext {
  hypothesis?: string;
  questions?: string[];
  focus?: string[];
}

export interface HypothesisContext {
  claim?: string;
  questions?: string[];
  focus?: string[];
  successCriteria?: SuccessCriterionSummary[];
  machineVerdict?: 'validated' | 'invalidated' | 'insufficient';
}

export interface SuccessCriterionSummary {
  metric: string;
  operator: string;
  value: string;
  description?: string;
  passed?: boolean;
  actualValue?: string;
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
  generatedAt: string;
  model: string;

  // Hypothesis verdict for at-a-glance display
  hypothesisVerdict?: 'validated' | 'invalidated' | 'insufficient';

  // Structured analysis sections
  abstract?: string;
  capabilitiesMatrix?: CapabilitiesMatrix;
  body?: AnalysisBody;
  feedback?: AnalysisFeedback;
  architectureDiagram?: string;
  architectureDiagramFormat?: 'ascii' | 'mermaid';
}

export interface CapabilitiesMatrix {
  technologies: string[];
  categories: CapabilitiesCategory[];
  summary?: string;
}

export interface CapabilitiesCategory {
  name: string;
  capabilities: CapabilityEntry[];
}

export interface CapabilityEntry {
  name: string;
  values: Record<string, string>;
}

export type BodyBlock =
  | { type: 'text'; content: string }
  | { type: 'topic'; title: string; blocks: BodyBlock[] }
  | { type: 'metric'; key: string; insight?: string; size?: 'large' | 'small' }
  | { type: 'comparison'; items: Array<{ label: string; value: string; description?: string }> }
  | { type: 'capabilityRow'; capability: string; values: Record<string, string> }
  | { type: 'table'; headers: string[]; rows: string[][]; caption?: string }
  | { type: 'architecture'; diagram: string; caption?: string; format?: 'ascii' | 'mermaid' }
  | { type: 'callout'; variant: 'info' | 'warning' | 'success' | 'finding'; title: string; content: string }
  | { type: 'recommendation'; priority: 'p0' | 'p1' | 'p2' | 'p3'; title: string; description: string; effort?: 'low' | 'medium' | 'high' }
  | { type: 'row'; blocks: BodyBlock[] };

export interface AnalysisBody {
  blocks: BodyBlock[];
}

export interface AnalysisFeedback {
  recommendations?: string[];
  experimentDesign?: string[];
}
