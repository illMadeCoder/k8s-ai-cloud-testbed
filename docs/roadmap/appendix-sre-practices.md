## Appendix: SRE Practices & Incident Management

*Operational excellence through Site Reliability Engineering practices. From on-call to post-mortems, this appendix covers the human and process side of running reliable systems.*

### K.1 SRE Fundamentals

**Goal:** Understand Site Reliability Engineering principles

**Learning objectives:**
- Understand SRE philosophy and practices
- Balance reliability with velocity
- Establish SRE culture

**Tasks:**
- [ ] Create `experiments/sre-fundamentals/`
- [ ] SRE principles:
  - [ ] Embrace risk (not eliminate)
  - [ ] SLOs as contract
  - [ ] Toil reduction
  - [ ] Automation over manual work
  - [ ] Blameless culture
- [ ] SRE vs DevOps:
  - [ ] Complementary approaches
  - [ ] SRE as DevOps implementation
  - [ ] Focus on reliability
- [ ] SRE team models:
  - [ ] Embedded SREs
  - [ ] Platform SRE team
  - [ ] Consulting model
  - [ ] Hybrid approaches
- [ ] Reliability hierarchy:
  - [ ] Monitoring
  - [ ] Incident response
  - [ ] Post-mortems
  - [ ] Testing and release
  - [ ] Capacity planning
  - [ ] Development
- [ ] Error budgets:
  - [ ] Budget calculation
  - [ ] Budget consumption tracking
  - [ ] Policy when exhausted
  - [ ] Feature velocity trade-off
- [ ] Production readiness:
  - [ ] Readiness reviews
  - [ ] Checklists
  - [ ] Graduation criteria
  - [ ] Ongoing requirements
- [ ] **ADR:** Document SRE adoption approach

---

### K.2 Service Level Objectives (SLOs)

**Goal:** Define and implement meaningful SLOs

**Learning objectives:**
- Design SLIs that matter
- Set appropriate SLO targets
- Implement SLO monitoring

**Tasks:**
- [ ] Create `experiments/slo-implementation/`
- [ ] SLI/SLO/SLA hierarchy:
  - [ ] SLI: Service Level Indicator (metric)
  - [ ] SLO: Service Level Objective (target)
  - [ ] SLA: Service Level Agreement (contract)
- [ ] Choosing SLIs:
  - [ ] Availability (successful requests / total)
  - [ ] Latency (requests faster than threshold)
  - [ ] Throughput (requests per second)
  - [ ] Correctness (correct responses)
- [ ] SLI measurement:
  - [ ] Server-side vs client-side
  - [ ] Synthetic vs real user
  - [ ] Measurement location matters
- [ ] Good SLI properties:
  - [ ] User-centric
  - [ ] Measurable
  - [ ] Actionable
  - [ ] Not too many
- [ ] Setting SLO targets:
  - [ ] Based on user expectations
  - [ ] Historical performance
  - [ ] Business requirements
  - [ ] Achievable but ambitious
- [ ] Common SLOs:
  - [ ] 99.9% availability
  - [ ] p99 latency < 200ms
  - [ ] Error rate < 0.1%
- [ ] Error budget calculation:
  - [ ] Budget = 1 - SLO
  - [ ] Time-based (monthly, quarterly)
  - [ ] Event-based (requests)
- [ ] Error budget policies:
  - [ ] Actions when budget low
  - [ ] Freeze deployments
  - [ ] Focus on reliability
  - [ ] Stakeholder communication
- [ ] SLO monitoring:
  - [ ] Burn rate alerts
  - [ ] Multi-window alerts
  - [ ] SLO dashboards
  - [ ] Budget tracking
- [ ] Tools:
  - [ ] Prometheus + Grafana
  - [ ] Sloth (SLO generator)
  - [ ] OpenSLO specification
- [ ] Implement SLOs for sample service
- [ ] **ADR:** Document SLO framework

---

### K.3 On-Call & Alerting

**Goal:** Build sustainable on-call practices

**Learning objectives:**
- Design effective alerting
- Structure on-call rotations
- Prevent alert fatigue

**Tasks:**
- [ ] Create `experiments/oncall-alerting/`
- [ ] Alerting philosophy:
  - [ ] Alert on symptoms, not causes
  - [ ] Actionable alerts only
  - [ ] Appropriate urgency
  - [ ] Human-friendly messages
- [ ] Alert types:
  - [ ] Pages (immediate action)
  - [ ] Tickets (action needed, not urgent)
  - [ ] Logs (informational)
- [ ] Alert design:
  - [ ] Clear title and description
  - [ ] Runbook links
  - [ ] Relevant context
  - [ ] Expected actions
- [ ] Reducing alert noise:
  - [ ] Alert consolidation
  - [ ] Dependency-aware alerting
  - [ ] Flap detection
  - [ ] Alert routing
- [ ] SLO-based alerting:
  - [ ] Burn rate alerts
  - [ ] Multi-window, multi-burn-rate
  - [ ] Reducing false positives
  - [ ] Catching real issues faster
- [ ] On-call structure:
  - [ ] Rotation schedules
  - [ ] Primary and secondary
  - [ ] Escalation paths
  - [ ] Coverage and handoffs
- [ ] On-call compensation:
  - [ ] Time off in lieu
  - [ ] Financial compensation
  - [ ] Workload balancing
- [ ] On-call health:
  - [ ] Sustainable rotations
  - [ ] Alert volume targets
  - [ ] Interrupt tracking
  - [ ] Burnout prevention
- [ ] Tools:
  - [ ] PagerDuty
  - [ ] Opsgenie
  - [ ] Grafana OnCall
  - [ ] Alertmanager
- [ ] On-call documentation:
  - [ ] On-call guide
  - [ ] Service catalogs
  - [ ] Emergency contacts
  - [ ] Escalation procedures
- [ ] Set up on-call system
- [ ] **ADR:** Document alerting strategy

---

### K.4 Incident Response

**Goal:** Respond effectively to production incidents

**Learning objectives:**
- Structure incident response process
- Coordinate during incidents
- Communicate effectively

**Tasks:**
- [ ] Create `experiments/incident-response/`
- [ ] Incident definition:
  - [ ] What constitutes an incident
  - [ ] Severity levels
  - [ ] Declaration process
  - [ ] vs Problems vs Alerts
- [ ] Severity levels:
  - [ ] SEV1: Critical (all hands)
  - [ ] SEV2: Major (immediate response)
  - [ ] SEV3: Minor (next business day)
  - [ ] Clear criteria for each
- [ ] Incident roles:
  - [ ] Incident Commander (IC)
  - [ ] Technical Lead
  - [ ] Communications Lead
  - [ ] Scribe
- [ ] Incident Commander:
  - [ ] Coordinates response
  - [ ] Makes decisions
  - [ ] Delegates tasks
  - [ ] Manages timeline
- [ ] Response phases:
  - [ ] Detection
  - [ ] Triage
  - [ ] Mitigation
  - [ ] Resolution
  - [ ] Follow-up
- [ ] Communication:
  - [ ] Status page updates
  - [ ] Internal communication
  - [ ] Customer communication
  - [ ] Executive updates
- [ ] War room practices:
  - [ ] Dedicated channel
  - [ ] Video bridge
  - [ ] Shared documents
  - [ ] Timeline tracking
- [ ] Mitigation vs resolution:
  - [ ] Mitigate first (stop bleeding)
  - [ ] Resolve later (fix root cause)
  - [ ] Rollback as mitigation
- [ ] Tools:
  - [ ] Incident management (PagerDuty, FireHydrant)
  - [ ] Communication (Slack, Teams)
  - [ ] Status pages (Statuspage.io)
  - [ ] Timeline tools
- [ ] Practice incidents:
  - [ ] Game days
  - [ ] Tabletop exercises
  - [ ] Fire drills
- [ ] Create incident response playbook
- [ ] **ADR:** Document incident process

---

### K.5 Post-Mortems & Learning

**Goal:** Learn from incidents to prevent recurrence

**Learning objectives:**
- Conduct blameless post-mortems
- Identify systemic issues
- Drive meaningful improvements

**Tasks:**
- [ ] Create `experiments/post-mortems/`
- [ ] Blameless culture:
  - [ ] Focus on systems, not people
  - [ ] Assume good intentions
  - [ ] Psychological safety
  - [ ] Learning over blame
- [ ] Post-mortem triggers:
  - [ ] All SEV1/SEV2 incidents
  - [ ] Near-misses
  - [ ] Novel failures
  - [ ] Customer impact
- [ ] Post-mortem timeline:
  - [ ] Within 48-72 hours
  - [ ] While memory fresh
  - [ ] Preliminary vs final
- [ ] Post-mortem structure:
  - [ ] Executive summary
  - [ ] Impact assessment
  - [ ] Timeline
  - [ ] Root cause analysis
  - [ ] Action items
  - [ ] Lessons learned
- [ ] Root cause analysis:
  - [ ] 5 Whys technique
  - [ ] Contributing factors
  - [ ] Systemic issues
  - [ ] Multiple root causes
- [ ] Action items:
  - [ ] Specific and measurable
  - [ ] Assigned owners
  - [ ] Due dates
  - [ ] Priority levels
- [ ] Action item types:
  - [ ] Detect (find faster)
  - [ ] Mitigate (reduce impact)
  - [ ] Prevent (stop recurrence)
  - [ ] Process improvements
- [ ] Post-mortem meeting:
  - [ ] Facilitated discussion
  - [ ] All relevant parties
  - [ ] Safe environment
  - [ ] Focus on learning
- [ ] Post-mortem follow-up:
  - [ ] Track action completion
  - [ ] Review effectiveness
  - [ ] Aggregate learnings
- [ ] Knowledge sharing:
  - [ ] Post-mortem repository
  - [ ] Cross-team sharing
  - [ ] Patterns and anti-patterns
- [ ] Create post-mortem template
- [ ] **ADR:** Document post-mortem process

---

### K.6 Runbooks & Documentation

**Goal:** Create actionable operational documentation

**Learning objectives:**
- Write effective runbooks
- Maintain living documentation
- Enable self-service operations

**Tasks:**
- [ ] Create `experiments/runbooks/`
- [ ] Runbook purpose:
  - [ ] Enable anyone to respond
  - [ ] Reduce mean time to recovery
  - [ ] Capture tribal knowledge
  - [ ] Enable consistent response
- [ ] Runbook structure:
  - [ ] Title and metadata
  - [ ] Description/overview
  - [ ] Prerequisites
  - [ ] Step-by-step instructions
  - [ ] Verification steps
  - [ ] Escalation path
- [ ] Good runbook properties:
  - [ ] Actionable (do this, not understand this)
  - [ ] Copy-paste commands
  - [ ] Current (matches reality)
  - [ ] Tested
- [ ] Runbook types:
  - [ ] Alert response runbooks
  - [ ] Procedure runbooks (deployments, etc.)
  - [ ] Troubleshooting guides
  - [ ] Emergency procedures
- [ ] Alert-linked runbooks:
  - [ ] Every alert has runbook
  - [ ] Link in alert message
  - [ ] Quick access during incident
- [ ] Runbook maintenance:
  - [ ] Review cadence
  - [ ] Update during incidents
  - [ ] Version control
  - [ ] Ownership
- [ ] Runbook automation:
  - [ ] Automated steps where possible
  - [ ] Human checkpoints
  - [ ] Runbook as code
  - [ ] ChatOps integration
- [ ] Service documentation:
  - [ ] Architecture overview
  - [ ] Dependencies
  - [ ] Configuration
  - [ ] Operational characteristics
- [ ] Documentation tools:
  - [ ] Markdown in git
  - [ ] Wiki systems
  - [ ] Backstage TechDocs
  - [ ] Notion, Confluence
- [ ] Create runbook library
- [ ] **ADR:** Document runbook standards

---

### K.7 Toil Reduction

**Goal:** Identify and eliminate operational toil

**Learning objectives:**
- Identify toil vs productive work
- Prioritize toil reduction efforts
- Automate repetitive tasks

**Tasks:**
- [ ] Create `experiments/toil-reduction/`
- [ ] Toil definition:
  - [ ] Manual work
  - [ ] Repetitive
  - [ ] Automatable
  - [ ] Tactical (no lasting value)
  - [ ] Scales with service growth
- [ ] Toil vs overhead:
  - [ ] Toil: operational work that scales
  - [ ] Overhead: necessary non-project work
  - [ ] Not all manual work is toil
- [ ] Toil measurement:
  - [ ] Time tracking
  - [ ] Ticket analysis
  - [ ] Survey teams
  - [ ] Interrupt logs
- [ ] Toil budget:
  - [ ] Target: <50% time on toil
  - [ ] Track over time
  - [ ] Alert on increase
- [ ] Common toil sources:
  - [ ] Manual deployments
  - [ ] Manual scaling
  - [ ] Repetitive alerts
  - [ ] Manual data fixes
  - [ ] Password resets
- [ ] Toil reduction strategies:
  - [ ] Automation
  - [ ] Self-service
  - [ ] Elimination (stop doing it)
  - [ ] Process improvement
- [ ] Automation prioritization:
  - [ ] Frequency × time saved
  - [ ] Error reduction
  - [ ] Team satisfaction
  - [ ] Risk reduction
- [ ] Self-service platforms:
  - [ ] Developer portals
  - [ ] ChatOps
  - [ ] API-driven operations
  - [ ] Backstage integration
- [ ] Measuring success:
  - [ ] Toil reduction over time
  - [ ] Team velocity increase
  - [ ] Incident reduction
  - [ ] Team satisfaction
- [ ] Identify and automate top toil
- [ ] **ADR:** Document automation priorities

---

### K.8 Capacity Management

**Goal:** Ensure sufficient capacity for reliability

**Learning objectives:**
- Forecast capacity needs
- Plan for growth
- Handle capacity emergencies

**Tasks:**
- [ ] Create `experiments/capacity-management/`
- [ ] Capacity planning cycle:
  - [ ] Current state assessment
  - [ ] Demand forecasting
  - [ ] Capacity modeling
  - [ ] Provisioning
  - [ ] Monitoring
- [ ] Demand signals:
  - [ ] Historical trends
  - [ ] Business projections
  - [ ] Seasonality
  - [ ] Marketing events
- [ ] Resource tracking:
  - [ ] CPU utilization
  - [ ] Memory usage
  - [ ] Storage consumption
  - [ ] Network bandwidth
- [ ] Headroom targets:
  - [ ] Target utilization (e.g., 70%)
  - [ ] Burst capacity
  - [ ] Failure scenarios
  - [ ] Lead time buffer
- [ ] Scaling strategies:
  - [ ] Vertical scaling limits
  - [ ] Horizontal scaling
  - [ ] Auto-scaling configuration
  - [ ] Manual intervention points
- [ ] Capacity alerts:
  - [ ] Utilization thresholds
  - [ ] Growth rate alerts
  - [ ] Projection-based alerts
- [ ] Capacity reviews:
  - [ ] Regular cadence
  - [ ] Cross-team coordination
  - [ ] Budget alignment
  - [ ] Long-term planning
- [ ] Emergency capacity:
  - [ ] Rapid scaling procedures
  - [ ] Reserved capacity
  - [ ] Degraded mode operation
  - [ ] Load shedding
- [ ] Cloud capacity:
  - [ ] Quotas and limits
  - [ ] Reserved instances
  - [ ] Spot/preemptible capacity
  - [ ] Multi-region capacity
- [ ] Create capacity dashboard
- [ ] **ADR:** Document capacity management process

---

### K.9 Reliability Testing

**Goal:** Proactively test system reliability

**Learning objectives:**
- Design reliability tests
- Implement game days
- Build confidence in systems

**Tasks:**
- [ ] Create `experiments/reliability-testing/`
- [ ] Testing types:
  - [ ] Load testing
  - [ ] Chaos engineering (Phase 12)
  - [ ] Disaster recovery testing
  - [ ] Failover testing
- [ ] Game days:
  - [ ] Planned incident simulations
  - [ ] Full team participation
  - [ ] Realistic scenarios
  - [ ] Learning focus
- [ ] Game day planning:
  - [ ] Scenario selection
  - [ ] Scope definition
  - [ ] Safety measures
  - [ ] Success criteria
- [ ] Game day execution:
  - [ ] Inject failure
  - [ ] Observe response
  - [ ] Take notes
  - [ ] Controlled environment
- [ ] Disaster recovery testing:
  - [ ] Failover to backup
  - [ ] Data recovery
  - [ ] Service restoration
  - [ ] Communication testing
- [ ] Tabletop exercises:
  - [ ] Discussion-based scenarios
  - [ ] No actual systems affected
  - [ ] Process validation
  - [ ] Team coordination
- [ ] Wheel of misfortune:
  - [ ] Random incident scenarios
  - [ ] On-call training
  - [ ] Response practice
  - [ ] Runbook validation
- [ ] Testing frequency:
  - [ ] Regular schedule
  - [ ] After major changes
  - [ ] Before peak periods
  - [ ] Annual DR tests
- [ ] Measuring effectiveness:
  - [ ] Recovery time
  - [ ] Process adherence
  - [ ] Communication quality
  - [ ] Issues discovered
- [ ] Plan and execute game day
- [ ] **ADR:** Document reliability testing strategy

---
