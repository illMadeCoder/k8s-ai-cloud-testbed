## Appendix: PKI & Certificate Management

*Deep dive into Public Key Infrastructure, certificate lifecycle, and TLS configuration. Covers everything from PKI fundamentals to running your own Certificate Authority with step-ca and automating certificates with cert-manager.*

### C.1 PKI Fundamentals

**Goal:** Understand the foundations of Public Key Infrastructure

**Learning objectives:**
- Understand asymmetric cryptography basics
- Grasp the certificate chain of trust model
- Know the components of X.509 certificates

**Tasks:**
- [ ] Create `experiments/scenarios/pki-fundamentals/`
- [ ] Asymmetric cryptography review:
  - [ ] Public/private key pairs
  - [ ] RSA key generation and sizes (2048, 4096)
  - [ ] ECDSA curves (P-256, P-384, P-521)
  - [ ] Ed25519 (modern, fast, fixed size)
  - [ ] Key exchange vs signing vs encryption
- [ ] X.509 certificate structure:
  - [ ] Version, serial number
  - [ ] Subject and Issuer Distinguished Names
  - [ ] Validity period (notBefore, notAfter)
  - [ ] Public key and algorithm
  - [ ] Extensions (critical vs non-critical)
  - [ ] Signature
- [ ] Key extensions:
  - [ ] Subject Alternative Name (SAN) - DNS, IP, URI
  - [ ] Key Usage (digitalSignature, keyEncipherment, etc.)
  - [ ] Extended Key Usage (serverAuth, clientAuth, codeSigning)
  - [ ] Basic Constraints (CA:TRUE/FALSE, pathlen)
  - [ ] Authority/Subject Key Identifier
- [ ] Certificate types:
  - [ ] Root CA certificates
  - [ ] Intermediate CA certificates
  - [ ] End-entity certificates (leaf)
  - [ ] Self-signed certificates
- [ ] Chain of trust:
  - [ ] Root → Intermediate → Leaf hierarchy
  - [ ] Trust anchors and trust stores
  - [ ] Certificate path validation
  - [ ] Cross-certification
- [ ] Certificate encoding formats:
  - [ ] PEM (Base64, -----BEGIN CERTIFICATE-----)
  - [ ] DER (binary)
  - [ ] PKCS#12/PFX (certificate + private key bundle)
  - [ ] PKCS#7 (certificate chain)
- [ ] Hands-on with OpenSSL:
  - [ ] Generate key pairs
  - [ ] Create CSRs
  - [ ] Inspect certificates
  - [ ] Verify certificate chains
- [ ] **ADR:** Document certificate key algorithm selection (RSA vs ECDSA vs Ed25519)

---

### C.2 TLS Protocol Deep Dive

**Goal:** Understand TLS handshake, versions, and cipher suites

**Learning objectives:**
- Understand TLS 1.2 and 1.3 handshakes
- Configure cipher suites appropriately
- Troubleshoot TLS connection issues

**Tasks:**
- [ ] Create `experiments/scenarios/tls-deep-dive/`
- [ ] TLS 1.2 handshake:
  - [ ] ClientHello / ServerHello
  - [ ] Certificate exchange
  - [ ] Key exchange (RSA, DHE, ECDHE)
  - [ ] Finished messages
  - [ ] Session resumption (session IDs, session tickets)
- [ ] TLS 1.3 improvements:
  - [ ] Reduced round trips (1-RTT, 0-RTT)
  - [ ] Removed legacy algorithms
  - [ ] Encrypted handshake (after ServerHello)
  - [ ] Key derivation (HKDF)
  - [ ] 0-RTT replay risks
- [ ] Cipher suite anatomy:
  - [ ] Key exchange (ECDHE, DHE)
  - [ ] Authentication (RSA, ECDSA)
  - [ ] Bulk encryption (AES-GCM, ChaCha20-Poly1305)
  - [ ] PRF/HKDF
- [ ] Recommended cipher suites:
  - [ ] TLS 1.3 suites (limited, all strong)
  - [ ] TLS 1.2 recommended suites
  - [ ] Suites to disable (CBC, RC4, 3DES, export)
- [ ] Server Name Indication (SNI):
  - [ ] How SNI works
  - [ ] Virtual hosting with TLS
  - [ ] SNI privacy concerns (ECH)
- [ ] Certificate verification:
  - [ ] Chain building
  - [ ] Revocation checking (CRL, OCSP)
  - [ ] OCSP stapling
  - [ ] Certificate Transparency (CT) logs
- [ ] Common issues:
  - [ ] Certificate name mismatch
  - [ ] Expired certificates
  - [ ] Incomplete certificate chains
  - [ ] Self-signed certificate errors
- [ ] Troubleshooting tools:
  - [ ] openssl s_client
  - [ ] curl -v
  - [ ] testssl.sh
  - [ ] SSL Labs server test
- [ ] **ADR:** Document TLS version and cipher suite policy

---

### C.3 Private Certificate Authority with step-ca

**Goal:** Deploy and operate a private CA for internal services

**Learning objectives:**
- Deploy step-ca as an internal Certificate Authority
- Issue certificates via ACME protocol
- Integrate with Kubernetes workloads

**Tasks:**
- [ ] Create `experiments/scenarios/private-ca/`
- [ ] step-ca architecture:
  - [ ] Smallstep ecosystem (step CLI, step-ca)
  - [ ] Provisioners (ACME, JWK, OIDC, X5C)
  - [ ] Database backends (Badger, PostgreSQL, MySQL)
- [ ] Deployment on Kubernetes:
  - [ ] Helm chart installation
  - [ ] Root CA initialization
  - [ ] Intermediate CA setup
  - [ ] High availability considerations
- [ ] Root CA security:
  - [ ] Offline root CA pattern
  - [ ] Root key protection (HSM, KMS)
  - [ ] Root certificate distribution
- [ ] ACME provisioner:
  - [ ] ACME protocol overview (RFC 8555)
  - [ ] Challenge types (HTTP-01, DNS-01, TLS-ALPN-01)
  - [ ] ACME account management
  - [ ] Rate limiting
- [ ] Certificate issuance:
  - [ ] step CLI certificate requests
  - [ ] Automated issuance via ACME
  - [ ] Certificate templates
  - [ ] Short-lived certificates (hours, not years)
- [ ] OIDC provisioner:
  - [ ] Identity-based certificates
  - [ ] Integration with Keycloak/Dex
  - [ ] Workload identity certificates
- [ ] SSH certificates (bonus):
  - [ ] step-ca as SSH CA
  - [ ] Short-lived SSH certificates
  - [ ] Single sign-on for SSH
- [ ] Operations:
  - [ ] Backup and disaster recovery
  - [ ] Key rotation
  - [ ] Monitoring and alerting
  - [ ] Audit logging
- [ ] Trust distribution:
  - [ ] Distributing root CA to nodes
  - [ ] Trust bundle ConfigMaps
  - [ ] OS trust store updates
- [ ] **ADR:** Document private CA architecture decisions

---

### C.4 cert-manager for Kubernetes

**Goal:** Automate certificate lifecycle in Kubernetes

**Learning objectives:**
- Deploy and configure cert-manager
- Use different issuers (Let's Encrypt, step-ca, Vault)
- Automate certificate rotation

**Tasks:**
- [ ] Create `experiments/scenarios/cert-manager/`
- [ ] cert-manager architecture:
  - [ ] Custom Resource Definitions (Certificate, Issuer, etc.)
  - [ ] Controller architecture
  - [ ] Webhook for validation
- [ ] Installation:
  - [ ] Helm chart deployment
  - [ ] CRD installation
  - [ ] RBAC configuration
- [ ] Issuer types:
  - [ ] ClusterIssuer vs Issuer (namespace-scoped)
  - [ ] ACME issuer (Let's Encrypt)
  - [ ] CA issuer (self-signed, private CA)
  - [ ] Vault issuer
  - [ ] Venafi issuer (enterprise)
- [ ] Let's Encrypt integration:
  - [ ] Staging vs production
  - [ ] HTTP-01 solver (ingress)
  - [ ] DNS-01 solver (Cloudflare, Route53, etc.)
  - [ ] Rate limits and best practices
- [ ] Private CA integration:
  - [ ] step-ca ACME issuer
  - [ ] CA issuer with step-ca root
  - [ ] Vault PKI backend
- [ ] Certificate resources:
  - [ ] Certificate spec (dnsNames, duration, renewBefore)
  - [ ] Secret templates
  - [ ] Key algorithm selection
- [ ] Ingress integration:
  - [ ] Annotation-based certificates
  - [ ] ingress-shim
  - [ ] Gateway API support
- [ ] Certificate lifecycle:
  - [ ] Automatic renewal
  - [ ] Renewal timing (renewBefore)
  - [ ] Failed issuance handling
  - [ ] Manual renewal triggers
- [ ] Advanced patterns:
  - [ ] Certificate per namespace
  - [ ] Wildcard certificates
  - [ ] Cross-namespace secrets (trust-manager)
  - [ ] Certificate policies
- [ ] trust-manager:
  - [ ] Trust bundle distribution
  - [ ] CA bundle ConfigMaps
  - [ ] Automatic updates
- [ ] Monitoring:
  - [ ] Prometheus metrics
  - [ ] Certificate expiry alerts
  - [ ] Grafana dashboards
- [ ] **ADR:** Document cert-manager issuer strategy

---

### C.5 Mutual TLS (mTLS)

**Goal:** Implement client certificate authentication

**Learning objectives:**
- Understand mTLS authentication model
- Configure mTLS for service-to-service communication
- Integrate mTLS with service mesh and ingress

**Tasks:**
- [ ] Create `experiments/scenarios/mtls/`
- [ ] mTLS fundamentals:
  - [ ] Server authentication (standard TLS)
  - [ ] Client authentication (mTLS addition)
  - [ ] Certificate-based identity
- [ ] mTLS handshake:
  - [ ] CertificateRequest from server
  - [ ] Client certificate and verify messages
  - [ ] Client certificate chain validation
- [ ] Use cases:
  - [ ] Service-to-service authentication
  - [ ] Zero-trust networking
  - [ ] API client authentication
  - [ ] IoT device authentication
- [ ] Manual mTLS setup:
  - [ ] Generate client certificates
  - [ ] Configure server for client auth
  - [ ] Client certificate in requests (curl, code)
- [ ] Ingress mTLS:
  - [ ] NGINX ingress client auth
  - [ ] Traefik mTLS
  - [ ] Contour client validation
  - [ ] Client certificate headers to backend
- [ ] Service mesh mTLS:
  - [ ] Istio automatic mTLS
  - [ ] Linkerd identity
  - [ ] mTLS modes (strict, permissive)
  - [ ] PeerAuthentication policies
- [ ] SPIFFE/SPIRE:
  - [ ] SPIFFE ID format
  - [ ] SPIRE architecture (server, agent)
  - [ ] Workload attestation
  - [ ] SVIDs (SPIFFE Verifiable Identity Documents)
  - [ ] x509-SVID vs JWT-SVID
- [ ] Certificate identity mapping:
  - [ ] Extracting identity from certificates
  - [ ] Subject DN to authorization
  - [ ] SAN-based identity
- [ ] Client certificate management:
  - [ ] Issuance workflows
  - [ ] Revocation handling
  - [ ] Rotation strategies
- [ ] **ADR:** Document mTLS implementation approach

---

### C.6 Certificate Transparency & Revocation

**Goal:** Understand certificate validity beyond expiration

**Learning objectives:**
- Understand Certificate Transparency logs
- Implement revocation checking
- Monitor for certificate misuse

**Tasks:**
- [ ] Create `experiments/scenarios/cert-transparency/`
- [ ] Certificate Transparency (CT):
  - [ ] Why CT exists (DigiNotar, etc.)
  - [ ] CT log structure (Merkle tree)
  - [ ] SCTs (Signed Certificate Timestamps)
  - [ ] CT log operators
- [ ] CT for your certificates:
  - [ ] Let's Encrypt automatic CT logging
  - [ ] Private CA and CT (usually not logged)
  - [ ] Pre-certificate vs final certificate
- [ ] CT monitoring:
  - [ ] crt.sh searches
  - [ ] Certificate monitoring services
  - [ ] Detecting unauthorized certificates
  - [ ] Alerting on new certificates for your domains
- [ ] Certificate revocation:
  - [ ] Why revocation is needed
  - [ ] Revocation != expiration
- [ ] CRL (Certificate Revocation Lists):
  - [ ] CRL structure
  - [ ] CRL Distribution Points extension
  - [ ] CRL size and freshness problems
  - [ ] Delta CRLs
- [ ] OCSP (Online Certificate Status Protocol):
  - [ ] OCSP request/response
  - [ ] OCSP responder
  - [ ] OCSP stapling (server-side)
  - [ ] OCSP Must-Staple extension
- [ ] Revocation in practice:
  - [ ] Browser behavior (soft-fail)
  - [ ] Why revocation often doesn't work
  - [ ] Short-lived certificates as alternative
- [ ] Private CA revocation:
  - [ ] step-ca revocation support
  - [ ] CRL hosting
  - [ ] OCSP responder setup
- [ ] **ADR:** Document revocation strategy (short-lived vs CRL/OCSP)

---

### C.7 Secrets Management Integration

**Goal:** Integrate PKI with secrets management systems

**Learning objectives:**
- Use Vault/OpenBao PKI secrets engine
- Implement just-in-time certificate issuance
- Manage certificate secrets securely

**Tasks:**
- [ ] Create `experiments/scenarios/pki-secrets-integration/`
- [ ] Vault PKI secrets engine:
  - [ ] Root CA generation in Vault
  - [ ] Intermediate CA setup
  - [ ] Roles for certificate issuance
  - [ ] TTL and max TTL configuration
- [ ] Certificate issuance via Vault:
  - [ ] API-based issuance
  - [ ] Vault Agent templating
  - [ ] Dynamic certificates
- [ ] Vault + cert-manager:
  - [ ] Vault issuer configuration
  - [ ] AppRole authentication
  - [ ] Kubernetes auth method
- [ ] OpenBao PKI:
  - [ ] PKI secrets engine (Vault-compatible)
  - [ ] Integration with hub cluster
- [ ] External Secrets Operator:
  - [ ] Syncing certificates to Kubernetes secrets
  - [ ] Refresh intervals
  - [ ] Secret templates
- [ ] Certificate injection patterns:
  - [ ] Init containers
  - [ ] Sidecar agents
  - [ ] Mounted secrets
  - [ ] Environment variables (certificates don't fit well)
- [ ] Key protection:
  - [ ] Private key never leaves source
  - [ ] Key generation location
  - [ ] Transit encryption for keys
- [ ] **ADR:** Document PKI + secrets management architecture

---

### C.8 Certificate Operations & Automation

**Goal:** Operationalize certificate management at scale

**Learning objectives:**
- Monitor certificate inventory
- Automate certificate lifecycle
- Handle certificate emergencies

**Tasks:**
- [ ] Create `experiments/scenarios/cert-operations/`
- [ ] Certificate inventory:
  - [ ] Discovering certificates in cluster
  - [ ] External certificate scanning
  - [ ] Certificate database/CMDB
- [ ] Monitoring:
  - [ ] Expiration alerting (30/14/7/1 day)
  - [ ] Certificate count metrics
  - [ ] Issuance failure alerts
  - [ ] Chain validity checks
- [ ] Grafana dashboards:
  - [ ] cert-manager metrics
  - [ ] Certificate expiry timeline
  - [ ] Issuance success rate
- [ ] Automation:
  - [ ] GitOps for certificate resources
  - [ ] Certificate templating
  - [ ] Namespace onboarding automation
- [ ] Rotation strategies:
  - [ ] Automated rotation (cert-manager)
  - [ ] Blue-green certificate rotation
  - [ ] Client notification for changes
- [ ] Emergency procedures:
  - [ ] Compromised key response
  - [ ] Mass certificate revocation
  - [ ] Emergency re-issuance
  - [ ] Root CA compromise (worst case)
- [ ] Certificate policies:
  - [ ] Maximum validity periods
  - [ ] Allowed key algorithms
  - [ ] Required extensions
  - [ ] Naming conventions
- [ ] Compliance:
  - [ ] Certificate audit trails
  - [ ] Key ceremony documentation
  - [ ] Retention requirements
- [ ] Create operational runbooks
- [ ] **ADR:** Document certificate lifecycle management approach

---

### C.9 Application TLS Configuration

**Goal:** Configure TLS correctly in applications and infrastructure

**Learning objectives:**
- Configure TLS in common applications
- Implement TLS termination patterns
- Avoid common TLS misconfigurations

**Tasks:**
- [ ] Create `experiments/scenarios/app-tls-config/`
- [ ] TLS termination patterns:
  - [ ] Edge termination (load balancer/ingress)
  - [ ] Passthrough (end-to-end)
  - [ ] Re-encryption (terminate and re-encrypt)
- [ ] Ingress TLS:
  - [ ] NGINX TLS configuration
  - [ ] Traefik TLS options
  - [ ] Contour HTTPProxy TLS
  - [ ] Gateway API TLS
- [ ] Backend TLS:
  - [ ] Service-to-service TLS
  - [ ] TLS to databases
  - [ ] TLS to message queues
- [ ] Application configuration:
  - [ ] Go TLS config
  - [ ] Node.js TLS options
  - [ ] Python ssl context
  - [ ] Java TrustStore/KeyStore
- [ ] Trust store management:
  - [ ] Adding custom CAs to containers
  - [ ] update-ca-certificates patterns
  - [ ] JVM trust store updates
  - [ ] Node.js NODE_EXTRA_CA_CERTS
- [ ] Common misconfigurations:
  - [ ] Disabled certificate verification
  - [ ] Weak cipher suites
  - [ ] Missing intermediate certificates
  - [ ] Hostname verification disabled
- [ ] Testing TLS:
  - [ ] SSL Labs scans
  - [ ] testssl.sh automation
  - [ ] TLS in CI/CD pipelines
- [ ] HTTP Strict Transport Security (HSTS):
  - [ ] HSTS headers
  - [ ] Preload lists
  - [ ] Risks of HSTS
- [ ] **ADR:** Document TLS termination strategy

---
