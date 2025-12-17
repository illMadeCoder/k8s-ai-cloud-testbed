## Phase 6: Network Security & Edge Protection

*Secure the network perimeter and protect against external threats. Covers firewalls, DDoS mitigation, WAF, and network security layers.*

### 6.1 Kubernetes Network Policies Deep Dive

**Goal:** Master native Kubernetes network segmentation

**Learning objectives:**
- Understand NetworkPolicy semantics and limitations
- Implement zero-trust network architecture
- Design policy hierarchies for multi-tenant clusters

**Tasks:**
- [ ] Create `experiments/scenarios/network-policies/`
- [ ] Default deny policies:
  - [ ] Implement default-deny ingress and egress
  - [ ] Namespace isolation patterns
  - [ ] Pod-to-pod communication controls
- [ ] Policy patterns:
  - [ ] Allow DNS egress (required for most workloads)
  - [ ] Allow specific namespace communication
  - [ ] Label-based pod selection
  - [ ] CIDR-based external access
- [ ] CNI-specific features:
  - [ ] Cilium NetworkPolicy extensions (L7, DNS)
  - [ ] Calico GlobalNetworkPolicy
  - [ ] Host network policy
- [ ] Testing and validation:
  - [ ] Policy testing with netshoot
  - [ ] Connectivity visualization
  - [ ] Policy audit and drift detection
- [ ] Document network policy patterns
- [ ] **ADR:** Document CNI selection for network policy needs

---

### 6.2 Web Application Firewall (WAF)

**Goal:** Protect applications from OWASP Top 10 and application-layer attacks

**Learning objectives:**
- Understand WAF architectures (reverse proxy vs sidecar)
- Configure rule sets and custom rules
- Balance security with false positive rates

**Tasks:**
- [ ] Create `experiments/scenarios/waf/`
- [ ] ModSecurity with Ingress:
  - [ ] Deploy NGINX Ingress with ModSecurity
  - [ ] OWASP Core Rule Set (CRS) configuration
  - [ ] Custom rules for application-specific threats
  - [ ] Anomaly scoring vs traditional mode
- [ ] Cloud WAF integration:
  - [ ] Azure Front Door WAF policies
  - [ ] AWS WAF with ALB Ingress Controller
  - [ ] Cloudflare WAF (tunnel integration)
- [ ] WAF operations:
  - [ ] Rule tuning and false positive reduction
  - [ ] Logging and alerting integration
  - [ ] Rate limiting rules
  - [ ] Geo-blocking policies
- [ ] Testing:
  - [ ] OWASP ZAP scanning
  - [ ] SQL injection and XSS test cases
  - [ ] WAF bypass testing
- [ ] Document WAF deployment patterns
- [ ] **ADR:** Document WAF placement (edge vs cluster)

---

### 6.3 DDoS Protection

**Goal:** Implement multi-layer DDoS mitigation strategies

**Learning objectives:**
- Understand DDoS attack types (volumetric, protocol, application)
- Implement defense-in-depth for DDoS
- Configure rate limiting and traffic shaping

**Tasks:**
- [ ] Create `experiments/scenarios/ddos-protection/`
- [ ] Cloud-native DDoS protection:
  - [ ] Azure DDoS Protection Standard
  - [ ] AWS Shield Standard and Advanced
  - [ ] Cloudflare DDoS protection (via tunnel)
- [ ] Cluster-level protection:
  - [ ] Ingress rate limiting (NGINX, Contour)
  - [ ] Connection limits per IP
  - [ ] Request rate limits per path/method
  - [ ] Slowloris and slow POST protection
- [ ] Application-level protection:
  - [ ] API rate limiting (per user, per API key)
  - [ ] CAPTCHA integration for suspicious traffic
  - [ ] Bot detection and mitigation
- [ ] Cilium-based protection:
  - [ ] eBPF-based packet filtering
  - [ ] XDP (eXpress Data Path) for early drop
  - [ ] Connection tracking limits
- [ ] Testing and simulation:
  - [ ] Load testing to find limits
  - [ ] Controlled DDoS simulation (vegeta, wrk)
  - [ ] Alerting threshold tuning
- [ ] Document DDoS mitigation architecture
- [ ] **ADR:** Document DDoS protection strategy by layer

---

### 6.4 Firewall & Network Segmentation

**Goal:** Implement defense-in-depth network architecture

**Learning objectives:**
- Design network zones and segmentation
- Integrate cloud firewalls with Kubernetes
- Implement egress controls

**Tasks:**
- [ ] Create `experiments/scenarios/network-segmentation/`
- [ ] Cloud firewall integration:
  - [ ] Azure Firewall with AKS
  - [ ] AWS Network Firewall with EKS
  - [ ] VPC/VNet network rules
  - [ ] Private endpoints for cloud services
- [ ] Egress control:
  - [ ] Egress gateway pattern
  - [ ] FQDN-based egress policies (Cilium)
  - [ ] Proxy-based egress (Squid, Envoy)
  - [ ] Audit and log all egress traffic
- [ ] Network zones:
  - [ ] DMZ for internet-facing workloads
  - [ ] Internal zone for backend services
  - [ ] Data zone for databases/storage
  - [ ] Management zone for cluster operations
- [ ] Micro-segmentation:
  - [ ] Workload identity-based policies
  - [ ] Service-to-service firewall rules
  - [ ] East-west traffic inspection
- [ ] Document network architecture patterns
- [ ] **ADR:** Document egress control strategy

---

### 6.5 API Gateway Security

**Goal:** Secure API endpoints with authentication, authorization, and threat protection

**Learning objectives:**
- Understand API gateway security features
- Implement API authentication patterns
- Configure API-specific threat protection

**Tasks:**
- [ ] Create `experiments/scenarios/api-gateway-security/`
- [ ] API authentication:
  - [ ] JWT validation at gateway
  - [ ] OAuth 2.0 / OIDC integration
  - [ ] API key management
  - [ ] mTLS for service-to-service
- [ ] API authorization:
  - [ ] Scope-based access control
  - [ ] OPA policies for API access
  - [ ] Rate limiting by consumer tier
- [ ] API threat protection:
  - [ ] Request size limits
  - [ ] JSON/XML schema validation
  - [ ] SQL injection in query params
  - [ ] Path traversal prevention
- [ ] Gateway options:
  - [ ] Kong Gateway security plugins
  - [ ] Ambassador/Emissary auth filters
  - [ ] Envoy external authorization
  - [ ] Cloud API Management (Azure APIM, AWS API Gateway)
- [ ] Document API security patterns
- [ ] **ADR:** Document API gateway selection

---

### 6.6 DNS Security

**Goal:** Protect DNS infrastructure and implement DNS-based security controls

**Learning objectives:**
- Understand DNS attack vectors
- Implement DNS security extensions
- Use DNS for threat detection

**Tasks:**
- [ ] Create `experiments/scenarios/dns-security/`
- [ ] DNS infrastructure hardening:
  - [ ] DNSSEC validation
  - [ ] DNS over HTTPS (DoH) / DNS over TLS (DoT)
  - [ ] Private DNS zones
  - [ ] Split-horizon DNS
- [ ] DNS-based security controls:
  - [ ] DNS sinkholing for malware domains
  - [ ] DNS query logging and analysis
  - [ ] DNS allowlisting for egress control
  - [ ] Response policy zones (RPZ)
- [ ] Kubernetes DNS security:
  - [ ] CoreDNS security configuration
  - [ ] DNS NetworkPolicy (allow DNS egress only to CoreDNS)
  - [ ] External DNS provider security
- [ ] Threat detection:
  - [ ] DNS tunneling detection
  - [ ] DGA (Domain Generation Algorithm) detection
  - [ ] Anomalous query pattern alerting
- [ ] Document DNS security architecture
- [ ] **ADR:** Document DNS security strategy

---

### 6.7 Zero Trust Network Architecture

**Goal:** Implement zero trust principles for network access

**Learning objectives:**
- Understand zero trust network model
- Implement identity-based access
- Design for least-privilege network access

**Tasks:**
- [ ] Create `experiments/scenarios/zero-trust-network/`
- [ ] Identity-based networking:
  - [ ] SPIFFE/SPIRE for workload identity
  - [ ] mTLS everywhere (service mesh)
  - [ ] Identity-aware proxy (IAP)
- [ ] Micro-perimeters:
  - [ ] Per-workload network policies
  - [ ] Just-in-time network access
  - [ ] Session-based access controls
- [ ] Continuous verification:
  - [ ] Network traffic analysis
  - [ ] Anomaly detection for lateral movement
  - [ ] Real-time policy enforcement
- [ ] Cloud integration:
  - [ ] Azure Private Link
  - [ ] AWS PrivateLink
  - [ ] Google Cloud Private Service Connect
- [ ] Document zero trust network architecture
- [ ] **ADR:** Document zero trust implementation approach

---

### 6.8 Network Observability & Threat Detection

**Goal:** Gain visibility into network traffic for security monitoring

**Learning objectives:**
- Implement network flow collection
- Build network security dashboards
- Create alerts for suspicious network activity

**Tasks:**
- [ ] Create `experiments/scenarios/network-observability/`
- [ ] Flow collection:
  - [ ] Hubble (Cilium network observability)
  - [ ] VPC Flow Logs (cloud)
  - [ ] NetFlow/IPFIX collection
- [ ] Network metrics:
  - [ ] Connection rates by source/destination
  - [ ] Bandwidth utilization
  - [ ] Packet drops and errors
  - [ ] Protocol distribution
- [ ] Security dashboards:
  - [ ] Grafana network security dashboard
  - [ ] Top talkers visualization
  - [ ] Geographic traffic mapping
  - [ ] Policy violation tracking
- [ ] Threat detection:
  - [ ] Port scanning detection
  - [ ] Unusual outbound connections
  - [ ] Data exfiltration indicators
  - [ ] C2 beacon detection
- [ ] Integration:
  - [ ] SIEM integration (Splunk, Elastic)
  - [ ] Automated incident creation
  - [ ] Network forensics collection
- [ ] Document network security monitoring
- [ ] **ADR:** Document network observability tooling

---
