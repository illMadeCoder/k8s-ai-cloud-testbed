## Phase 5: Deployment Strategies

*Progressive complexity: rolling → blue-green → canary → GitOps patterns → feature flags.*

### 7.1 Rolling Updates Optimization

**Goal:** Master Kubernetes native rolling deployments

**Learning objectives:**
- Understand rolling update parameters
- Optimize for zero-downtime deployments
- Handle graceful shutdown correctly

**Tasks:**
- [ ] Create `experiments/scenarios/rolling-update-tutorial/`
- [ ] Build app with slow startup and graceful shutdown
- [ ] Test parameter combinations:
  - [ ] maxSurge/maxUnavailable variations
  - [ ] minReadySeconds impact
  - [ ] progressDeadlineSeconds
- [ ] Implement graceful shutdown:
  - [ ] preStop hooks
  - [ ] terminationGracePeriodSeconds
  - [ ] Connection draining
- [ ] Readiness probe tuning:
  - [ ] initialDelaySeconds
  - [ ] periodSeconds
  - [ ] failureThreshold
- [ ] Load test during rollout (measure errors)
- [ ] Document recommended configurations

---

### 7.2 Blue-Green Deployments

**Goal:** Implement instant cutover deployments

**Learning objectives:**
- Understand blue-green pattern
- Implement with different tools
- Handle rollback scenarios

**Tasks:**
- [ ] Create `experiments/scenarios/blue-green-tutorial/`
- [ ] Implement blue-green with:
  - [ ] Kubernetes Services (label selector swap)
  - [ ] Gateway API traffic switching
  - [ ] Argo Rollouts BlueGreen strategy
- [ ] Test scenarios:
  - [ ] Successful deployment
  - [ ] Failed health check (no switch)
  - [ ] Rollback after deployment
- [ ] Measure:
  - [ ] Cutover time
  - [ ] Request failures during switch
  - [ ] Resource overhead (2x replicas)
- [ ] Handle stateful considerations:
  - [ ] Database compatibility
  - [ ] Session handling
- [ ] Document blue-green patterns

---

### 7.3 Canary Deployments with Argo Rollouts

**Goal:** Implement gradual traffic shifting with automated analysis

*Requires: Phase 4.2 (SLOs) for analysis metrics, Phase 5.1 (Gateway API) for traffic splitting*

**Learning objectives:**
- Understand canary deployment pattern
- Configure Argo Rollouts
- Implement metric-based promotion/rollback

**Tasks:**
- [ ] Create `experiments/scenarios/canary-tutorial/`
- [ ] Install Argo Rollouts
- [ ] Configure Rollout resource:
  - [ ] Traffic splitting steps (5% → 25% → 50% → 100%)
  - [ ] Pause durations
  - [ ] Manual gates
- [ ] Implement AnalysisTemplate:
  - [ ] Success rate query (Prometheus)
  - [ ] Latency threshold query
  - [ ] Custom business metrics
- [ ] Create "bad" versions to test:
  - [ ] High error rate version
  - [ ] High latency version
- [ ] Test automated rollback on failure
- [ ] Integrate with:
  - [ ] Gateway API (traffic splitting)
  - [ ] Istio (if mesh deployed)
- [ ] Document canary analysis patterns

---

### 7.4 GitOps Patterns with ArgoCD

**Goal:** Master ArgoCD for GitOps deployments

**Learning objectives:**
- Understand ArgoCD sync strategies
- Implement progressive delivery via Git
- Use ApplicationSets for multi-cluster

**Tasks:**
- [ ] Create `experiments/scenarios/argocd-patterns/`
- [ ] Sync strategies:
  - [ ] Auto-sync vs manual
  - [ ] Self-heal behavior
  - [ ] Prune policies
- [ ] Sync waves and hooks:
  - [ ] Pre-sync hooks (DB migration job)
  - [ ] Sync waves (ordering)
  - [ ] Post-sync hooks (smoke tests)
  - [ ] SyncFail hooks (notifications)
- [ ] ApplicationSet patterns:
  - [ ] Git generator (directory/file)
  - [ ] Cluster generator (multi-cluster)
  - [ ] Matrix generator (combinations)
  - [ ] Progressive rollout across clusters
- [ ] App-of-apps pattern
- [ ] Document GitOps workflow patterns

---

### 7.5 Feature Flags & Progressive Delivery

**Goal:** Decouple deployment from release with feature flags

**Learning objectives:**
- Understand feature flag patterns
- Implement OpenFeature standard
- Combine with deployment strategies

**Tasks:**
- [ ] Create `experiments/scenarios/feature-flags-tutorial/`
- [ ] Deploy feature flag service:
  - [ ] Flagsmith (self-hosted) OR
  - [ ] OpenFeature with flagd
- [ ] Implement flag patterns:
  - [ ] Boolean flags (feature on/off)
  - [ ] Percentage rollout
  - [ ] User segment targeting
  - [ ] A/B testing variants
- [ ] Integrate with application:
  - [ ] OpenFeature SDK integration
  - [ ] Server-side evaluation
  - [ ] Client-side evaluation
- [ ] Operational patterns:
  - [ ] Flag lifecycle (create → test → release → cleanup)
  - [ ] Kill switches for incidents
  - [ ] Gradual rollout with monitoring
- [ ] Combine with canary:
  - [ ] Deploy code → enable flag → monitor → full release
- [ ] Document feature flag patterns
- [ ] **ADR:** Document deployment vs release strategy

---

### 7.6 SLO-Based Release Decisions

**Goal:** Use SLOs and error budgets to gate deployment promotion and rollback

*Requires: Phase 4.2 (SLOs & Error Budgets) - see also: Phase 12.4 (Chaos SLO Impact)*

**Learning objectives:**
- Connect SLIs to deployment analysis
- Implement error budget-aware promotions
- Automate rollback based on SLO violation

**Tasks:**
- [ ] Create `experiments/scenarios/slo-based-deployment/`
- [ ] SLO integration with Argo Rollouts:
  - [ ] AnalysisTemplate using SLI queries
  - [ ] Error rate SLO (e.g., 99.9% success rate)
  - [ ] Latency SLO (e.g., p99 < 200ms)
  - [ ] Multi-metric analysis (AND/OR logic)
- [ ] Error budget-aware deployment:
  - [ ] Query remaining error budget before deploy
  - [ ] Block deployment if budget exhausted
  - [ ] Gradual rollout based on budget remaining
- [ ] SLO burn rate during canary:
  - [ ] Measure burn rate increase with new version
  - [ ] Auto-rollback if burn rate exceeds threshold
  - [ ] Compare baseline vs canary SLI performance
- [ ] Deployment gates:
  - [ ] Pre-deployment: Check error budget > threshold
  - [ ] During canary: SLI degradation triggers rollback
  - [ ] Post-deployment: SLO validation period
- [ ] Alerting integration:
  - [ ] Deployment-aware alert suppression
  - [ ] Rollback notifications
  - [ ] SLO status in deployment dashboards
- [ ] Document SLO-based deployment patterns
- [ ] **ADR:** Document SLO thresholds for deployment decisions

---

