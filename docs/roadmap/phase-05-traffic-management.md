## Phase 5: Traffic Management

*Control how traffic flows before learning deployment strategies that depend on it.*

### 5.1 Gateway API Deep Dive

**Goal:** Master Kubernetes Gateway API for ingress and traffic routing

**Learning objectives:**
- Understand Gateway API resources (Gateway, HTTPRoute, GRPCRoute)
- Implement advanced routing patterns
- Compare with legacy Ingress

**Tasks:**
- [ ] Create `experiments/scenarios/gateway-api-tutorial/`
- [ ] Deploy Gateway API implementation:
  - [ ] **Contour** (work requirement - Envoy-based)
  - [ ] Envoy Gateway (alternative)
  - [ ] Cilium Gateway (if using Cilium CNI)
- [ ] Configure Gateway resource
- [ ] Implement HTTPRoute patterns:
  - [ ] Path-based routing
  - [ ] Host-based routing (virtual hosts)
  - [ ] Header matching
  - [ ] Query parameter routing
  - [ ] Method matching (GET vs POST)
- [ ] Traffic manipulation:
  - [ ] Weight-based splitting (A/B)
  - [ ] Request mirroring
  - [ ] URL rewriting
  - [ ] Header modification (add/remove/set)
  - [ ] Redirects
- [ ] Advanced features:
  - [ ] Timeouts and retries
  - [ ] Rate limiting (via policy attachment)
  - [ ] CORS configuration
- [ ] TLS configuration:
  - [ ] TLS termination (with cert-manager certs)
  - [ ] TLS passthrough
  - [ ] mTLS with client certificates
- [ ] Multi-gateway setup:
  - [ ] Internal vs external gateways
  - [ ] Namespace isolation (ReferenceGrant)
- [ ] Document Gateway API vs Ingress migration
- [ ] **ADR:** Document Gateway API implementation choice

---

### 5.2 Ingress Controllers Comparison

**Goal:** Understand trade-offs between ingress implementations

**Learning objectives:**
- Compare nginx, Traefik, and Envoy-based controllers
- Understand feature/performance trade-offs
- Make informed controller selection

**Tasks:**
- [ ] Create `experiments/scenarios/ingress-comparison/`
- [ ] Deploy and configure:
  - [ ] **Contour** (work requirement - Envoy-based, Gateway API native)
  - [ ] Nginx Ingress Controller
  - [ ] Traefik
  - [ ] Envoy Gateway
- [ ] Implement equivalent routing on each
- [ ] Compare:
  - [ ] Configuration complexity
  - [ ] Feature availability
  - [ ] Resource consumption
  - [ ] Custom resource patterns
- [ ] Test advanced features:
  - [ ] Rate limiting implementation
  - [ ] Authentication integration
  - [ ] Custom error pages
- [ ] Document selection criteria

---

### 5.3 API Gateway Patterns

**Goal:** Implement API management patterns beyond basic routing

**Learning objectives:**
- Understand API gateway responsibilities
- Implement authentication, rate limiting, and API versioning
- Evaluate managed vs self-hosted options

**Tasks:**
- [ ] Create `experiments/scenarios/api-gateway-tutorial/`
- [ ] Deploy Kong or Ambassador (or use Envoy Gateway)
- [ ] Implement API management features:
  - [ ] API key authentication
  - [ ] JWT validation (integrate with **Auth0** from Phase 3.7)
  - [ ] OAuth2/OIDC integration
- [ ] Rate limiting patterns:
  - [ ] Per-client rate limits
  - [ ] Global rate limits
  - [ ] Quota management
- [ ] API versioning strategies:
  - [ ] Path-based (/v1/, /v2/)
  - [ ] Header-based (Accept-Version)
  - [ ] Request transformation between versions
- [ ] API analytics:
  - [ ] Request logging
  - [ ] Usage metrics per consumer
  - [ ] Error rate by endpoint
- [ ] Developer portal (optional):
  - [ ] API documentation (OpenAPI)
  - [ ] Self-service key provisioning
- [ ] Document API gateway patterns
- [ ] **ADR:** Document API versioning strategy

---

