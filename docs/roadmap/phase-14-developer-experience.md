## Phase 14: Developer Experience & Internal Platform

*Build an Internal Developer Platform (IDP) that ties together all the platform pieces.*

### 14.1 Backstage Developer Portal

**Goal:** Deploy Backstage as a unified developer portal

**Learning objectives:**
- Understand Internal Developer Platform (IDP) concepts
- Configure Backstage catalog and plugins
- Integrate with existing platform components

**Tasks:**
- [ ] Create `experiments/scenarios/backstage-tutorial/`
- [ ] Deploy Backstage:
  - [ ] Helm chart or ArgoCD Application
  - [ ] PostgreSQL backend (via CloudNativePG from Phase 10)
  - [ ] Authentication (integrate with Auth0 from Phase 3.7)
- [ ] Software Catalog:
  - [ ] Define catalog-info.yaml for services
  - [ ] Component types (service, website, library)
  - [ ] System and domain groupings
  - [ ] API definitions (OpenAPI, AsyncAPI, gRPC)
- [ ] Integrations:
  - [ ] Kubernetes plugin (show deployments, pods)
  - [ ] ArgoCD plugin (deployment status)
  - [ ] GitHub/GitLab integration (repo info, CI status)
  - [ ] Prometheus/Grafana plugin (metrics links)
  - [ ] PagerDuty or Opsgenie plugin (on-call info)
- [ ] Software Templates:
  - [ ] Create scaffolder template for new services
  - [ ] Include CI/CD pipeline, Dockerfile, Helm chart
  - [ ] Integrate with Crossplane for infrastructure
- [ ] TechDocs:
  - [ ] Enable TechDocs plugin
  - [ ] Generate docs from markdown in repos
- [ ] Document Backstage patterns
- [ ] **ADR:** Document IDP strategy (Backstage vs Port vs Cortex)

---

### 14.2 Self-Service Infrastructure

**Goal:** Enable developer self-service through the platform

**Learning objectives:**
- Design golden paths for common workflows
- Balance flexibility with guardrails
- Measure developer productivity

**Tasks:**
- [ ] Create `experiments/scenarios/self-service-infra/`
- [ ] Golden paths:
  - [ ] New service creation (Backstage template → repo → CI/CD → deployed)
  - [ ] Database provisioning (Backstage → Crossplane claim → ready)
  - [ ] Environment creation (dev/staging/prod namespaces)
- [ ] Guardrails integration:
  - [ ] Policies from Phase 3.5 enforced automatically
  - [ ] Cost controls from Phase 9.6 applied
  - [ ] Security scanning from Phase 2 in templates
- [ ] Developer metrics:
  - [ ] Lead time for changes
  - [ ] Deployment frequency
  - [ ] Time to onboard new service
- [ ] Document self-service patterns

---

