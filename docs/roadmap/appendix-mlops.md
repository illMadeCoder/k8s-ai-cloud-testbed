## Appendix: MLOps & AI Infrastructure

*Learn modern MLOps patterns for training and serving ML models on Kubernetes. This phase is specialized - do it when you need ML workloads.*

> **Note on AI-assisted experiment analysis:** We'll discover where AI helps as we run experiments manually. Pain points will reveal the right solutions - maybe an agent per workload per cluster, maybe something else. This will evolve organically and be captured in ADRs as patterns emerge.

### A.1 Kubeflow Pipelines & MLOps

**Goal:** Implement ML training and serving pipelines on Kubernetes

**Learning objectives:**
- Understand Kubeflow components and architecture
- Build end-to-end ML pipelines
- Implement model versioning and serving

**Tasks:**
- [ ] Create `experiments/scenarios/kubeflow-tutorial/`
- [ ] Deploy Kubeflow components:
  - [ ] Kubeflow Pipelines
  - [ ] Katib (hyperparameter tuning)
  - [ ] KServe (model serving)
- [ ] Build ML pipeline:
  - [ ] Data preprocessing step
  - [ ] Model training step
  - [ ] Model evaluation step
  - [ ] Model registration
- [ ] Implement MLOps patterns:
  - [ ] Experiment tracking (MLflow or Kubeflow native)
  - [ ] Model versioning and lineage
  - [ ] A/B model serving
  - [ ] Canary model deployments
- [ ] Integrate with platform:
  - [ ] Artifact storage (MinIO)
  - [ ] Metrics to Prometheus
  - [ ] Pipeline triggers from Argo Events
- [ ] Document MLOps architecture
- [ ] **ADR:** Document ML platform selection

---

### A.2 KServe Model Serving

**Goal:** Deploy and manage ML models in production

**Learning objectives:**
- Understand KServe architecture
- Implement inference autoscaling
- Configure model monitoring

**Tasks:**
- [ ] Create `experiments/scenarios/kserve-tutorial/`
- [ ] Deploy KServe:
  - [ ] Serverless inference
  - [ ] RawDeployment mode comparison
- [ ] Model serving patterns:
  - [ ] Single model deployment
  - [ ] Multi-model serving
  - [ ] Model transformers (pre/post processing)
- [ ] Autoscaling:
  - [ ] Scale-to-zero configuration
  - [ ] GPU autoscaling
  - [ ] Request-based scaling
- [ ] Traffic management:
  - [ ] Canary rollouts for models
  - [ ] A/B testing
  - [ ] Shadow deployments
- [ ] Monitoring:
  - [ ] Inference latency metrics
  - [ ] Model drift detection
  - [ ] Request logging
- [ ] Document model serving patterns

---

### A.3 Vector Databases & RAG Infrastructure

**Goal:** Deploy vector search infrastructure for AI applications

**Learning objectives:**
- Understand vector database architectures
- Implement RAG (Retrieval Augmented Generation) patterns
- Evaluate different vector DB options

**Tasks:**
- [ ] Create `experiments/scenarios/vector-db-tutorial/`
- [ ] Deploy vector databases:
  - [ ] Qdrant (Kubernetes-native)
  - [ ] Weaviate OR Milvus (comparison)
- [ ] Implement RAG pipeline:
  - [ ] Document ingestion and chunking
  - [ ] Embedding generation
  - [ ] Vector storage and indexing
  - [ ] Semantic search queries
  - [ ] LLM integration for generation
- [ ] Operational patterns:
  - [ ] Index management
  - [ ] Backup and restore
  - [ ] Horizontal scaling
- [ ] Build practical application:
  - [ ] Documentation search for this lab
  - [ ] Experiment results Q&A
- [ ] Compare vector DBs:
  - [ ] Query performance
  - [ ] Resource consumption
  - [ ] Ease of operation
- [ ] Document RAG architecture patterns
- [ ] **ADR:** Document vector database selection

---

### A.4 GPU Scheduling & Resource Management

**Goal:** Manage GPU resources for ML workloads

**Learning objectives:**
- Understand GPU scheduling in Kubernetes
- Configure GPU sharing and limits
- Optimize GPU utilization

**Tasks:**
- [ ] Create `experiments/scenarios/gpu-scheduling/`
- [ ] GPU fundamentals:
  - [ ] NVIDIA device plugin
  - [ ] GPU resource requests/limits
  - [ ] Node selectors and taints for GPU nodes
- [ ] GPU sharing:
  - [ ] Time-slicing
  - [ ] MIG (Multi-Instance GPU) if available
  - [ ] MPS (Multi-Process Service)
- [ ] Cloud GPU options:
  - [ ] AKS GPU node pools
  - [ ] EKS GPU instances
  - [ ] Spot/preemptible GPU instances for cost
- [ ] Monitoring:
  - [ ] DCGM exporter for GPU metrics
  - [ ] Grafana GPU dashboards
- [ ] Document GPU scheduling patterns
- [ ] **ADR:** Document GPU strategy (cloud vs local, sharing approach)

---
