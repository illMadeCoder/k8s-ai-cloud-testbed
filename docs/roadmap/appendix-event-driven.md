## Appendix: Event-Driven Architecture

*Deep dive into event-driven systems beyond "deploy Kafka." Covers event sourcing, CQRS, Saga patterns, and the architectural patterns that make event-driven systems reliable and maintainable.*

### I.1 Event-Driven Fundamentals

**Goal:** Understand core event-driven architecture concepts

**Learning objectives:**
- Distinguish between event types
- Understand event-driven communication patterns
- Know when event-driven architecture fits

**Tasks:**
- [ ] Create `experiments/event-driven-fundamentals/`
- [ ] Event types:
  - [ ] Domain events (business occurrences)
  - [ ] Integration events (cross-service)
  - [ ] Event notifications (signals without data)
  - [ ] Event-carried state transfer (data included)
- [ ] Event anatomy:
  - [ ] Event ID (unique identifier)
  - [ ] Event type
  - [ ] Timestamp
  - [ ] Payload
  - [ ] Metadata (correlation ID, causation ID)
- [ ] Communication patterns:
  - [ ] Publish-subscribe
  - [ ] Point-to-point queues
  - [ ] Request-reply (async)
  - [ ] Event streaming
- [ ] Coupling considerations:
  - [ ] Temporal coupling (both parties available)
  - [ ] Spatial coupling (knowing location)
  - [ ] Schema coupling (data format)
- [ ] Benefits:
  - [ ] Loose coupling
  - [ ] Scalability
  - [ ] Resilience
  - [ ] Audit trail
- [ ] Challenges:
  - [ ] Eventual consistency
  - [ ] Debugging complexity
  - [ ] Event ordering
  - [ ] Duplicate handling
- [ ] When to use:
  - [ ] Async processing needs
  - [ ] Multiple consumers
  - [ ] Audit requirements
  - [ ] Decoupled services
- [ ] When to avoid:
  - [ ] Simple CRUD
  - [ ] Synchronous requirements
  - [ ] Low complexity systems
- [ ] **ADR:** Document event-driven adoption criteria

---

### I.2 Event Sourcing

**Goal:** Understand event sourcing as a persistence pattern

**Learning objectives:**
- Implement event sourcing correctly
- Understand event store requirements
- Handle event sourcing challenges

**Tasks:**
- [ ] Create `experiments/event-sourcing/`
- [ ] Event sourcing fundamentals:
  - [ ] State from event history
  - [ ] Append-only event log
  - [ ] Current state via replay
  - [ ] Events as source of truth
- [ ] Event store requirements:
  - [ ] Append-only writes
  - [ ] Stream per aggregate
  - [ ] Optimistic concurrency
  - [ ] Ordering guarantees
- [ ] Event store options:
  - [ ] EventStoreDB
  - [ ] Axon Server
  - [ ] Kafka (with considerations)
  - [ ] PostgreSQL (as event store)
- [ ] Aggregate design:
  - [ ] Aggregate root
  - [ ] Event application
  - [ ] Command handling
  - [ ] Invariant enforcement
- [ ] Event versioning:
  - [ ] Schema evolution
  - [ ] Upcasting events
  - [ ] Weak vs strong schema
  - [ ] Event contracts
- [ ] Snapshots:
  - [ ] Why snapshots (replay performance)
  - [ ] Snapshot frequency
  - [ ] Snapshot storage
  - [ ] Snapshot + events replay
- [ ] Projections:
  - [ ] Read models from events
  - [ ] Projection rebuilding
  - [ ] Multiple projections
  - [ ] Projection consistency
- [ ] Challenges:
  - [ ] Event schema evolution
  - [ ] External system integration
  - [ ] Privacy (GDPR, event deletion)
  - [ ] Learning curve
- [ ] Implement event-sourced aggregate
- [ ] **ADR:** Document event sourcing decision

---

### I.3 CQRS Pattern

**Goal:** Understand Command Query Responsibility Segregation

**Learning objectives:**
- Implement CQRS effectively
- Know when CQRS adds value
- Combine CQRS with event sourcing

**Tasks:**
- [ ] Create `experiments/cqrs/`
- [ ] CQRS fundamentals:
  - [ ] Separate read and write models
  - [ ] Command side (writes)
  - [ ] Query side (reads)
  - [ ] Independent scaling
- [ ] Command model:
  - [ ] Command handlers
  - [ ] Domain logic
  - [ ] Validation
  - [ ] Event generation
- [ ] Query model:
  - [ ] Denormalized views
  - [ ] Optimized for queries
  - [ ] Multiple projections
  - [ ] Different storage technologies
- [ ] Synchronization:
  - [ ] Event-based sync
  - [ ] Eventual consistency
  - [ ] Sync lag considerations
  - [ ] Consistency requirements
- [ ] CQRS without event sourcing:
  - [ ] Simpler implementation
  - [ ] Database views
  - [ ] Change data capture
  - [ ] Dual writes (with caution)
- [ ] CQRS with event sourcing:
  - [ ] Events drive projections
  - [ ] Natural fit
  - [ ] Projection rebuilding
  - [ ] Event replay for new views
- [ ] Query optimization:
  - [ ] Materialized views
  - [ ] Denormalization strategies
  - [ ] Caching patterns
  - [ ] Search integration (Elasticsearch)
- [ ] When CQRS helps:
  - [ ] Complex domains
  - [ ] Different read/write patterns
  - [ ] High read scalability needs
  - [ ] Multiple view requirements
- [ ] When to avoid:
  - [ ] Simple domains
  - [ ] Strong consistency needs
  - [ ] Small teams
- [ ] Implement CQRS system
- [ ] **ADR:** Document CQRS adoption

---

### I.4 Saga Pattern

**Goal:** Implement distributed transactions with Sagas

**Learning objectives:**
- Design saga workflows
- Implement compensation logic
- Choose orchestration vs choreography

**Tasks:**
- [ ] Create `experiments/saga-pattern/`
- [ ] Saga fundamentals:
  - [ ] Sequence of local transactions
  - [ ] Compensation for rollback
  - [ ] No distributed locks
  - [ ] Eventual consistency
- [ ] Choreography-based saga:
  - [ ] Event-driven coordination
  - [ ] Services react to events
  - [ ] Decentralized logic
  - [ ] No single coordinator
- [ ] Choreography implementation:
  - [ ] Event publishing
  - [ ] Event handlers
  - [ ] State machines per service
  - [ ] Correlation handling
- [ ] Choreography challenges:
  - [ ] Hard to understand full flow
  - [ ] Cyclic dependencies risk
  - [ ] Testing complexity
  - [ ] Adding steps requires multiple changes
- [ ] Orchestration-based saga:
  - [ ] Central saga orchestrator
  - [ ] Explicit workflow definition
  - [ ] Command-based coordination
  - [ ] Centralized state tracking
- [ ] Orchestration implementation:
  - [ ] Saga orchestrator service
  - [ ] State machine
  - [ ] Command/reply pattern
  - [ ] Timeout handling
- [ ] Orchestration tools:
  - [ ] Temporal.io
  - [ ] Camunda
  - [ ] AWS Step Functions
  - [ ] Custom implementation
- [ ] Compensation design:
  - [ ] Semantic undo (not always exact reverse)
  - [ ] Idempotent compensations
  - [ ] Compensation ordering
  - [ ] Partial compensation
- [ ] Error handling:
  - [ ] Retries with backoff
  - [ ] Dead letter handling
  - [ ] Manual intervention
  - [ ] Saga timeout
- [ ] Implement saga with Temporal
- [ ] **ADR:** Document saga pattern selection

---

### I.5 Outbox Pattern

**Goal:** Ensure reliable event publishing

**Learning objectives:**
- Implement transactional outbox
- Understand change data capture
- Handle publisher failures gracefully

**Tasks:**
- [ ] Create `experiments/outbox-pattern/`
- [ ] The problem:
  - [ ] Database write + message publish
  - [ ] Dual write problem
  - [ ] Partial failures
  - [ ] Inconsistency risk
- [ ] Outbox pattern:
  - [ ] Write event to outbox table
  - [ ] Same transaction as business data
  - [ ] Separate process publishes
  - [ ] At-least-once delivery
- [ ] Outbox implementation:
  - [ ] Outbox table schema
  - [ ] Transaction scope
  - [ ] Publisher process
  - [ ] Delivery tracking
- [ ] Polling publisher:
  - [ ] Poll outbox table
  - [ ] Publish messages
  - [ ] Mark as published
  - [ ] Polling interval tuning
- [ ] Change Data Capture (CDC):
  - [ ] Database log monitoring
  - [ ] Debezium
  - [ ] No polling overhead
  - [ ] Lower latency
- [ ] Debezium setup:
  - [ ] Kafka Connect
  - [ ] Database connectors
  - [ ] Outbox transformation
  - [ ] Event routing
- [ ] Idempotency handling:
  - [ ] Duplicate detection
  - [ ] Message IDs
  - [ ] Consumer idempotency
- [ ] Ordering considerations:
  - [ ] Per-aggregate ordering
  - [ ] Partition keys
  - [ ] Sequence numbers
- [ ] Failure scenarios:
  - [ ] Publisher crash
  - [ ] Database failure
  - [ ] Message broker failure
  - [ ] Recovery procedures
- [ ] Implement outbox with Debezium
- [ ] **ADR:** Document reliable messaging approach

---

### I.6 Event Schema Evolution

**Goal:** Evolve event schemas without breaking consumers

**Learning objectives:**
- Design evolvable event schemas
- Handle schema versioning
- Implement schema registry

**Tasks:**
- [ ] Create `experiments/event-schema-evolution/`
- [ ] Schema evolution challenges:
  - [ ] Multiple consumers
  - [ ] Different deployment times
  - [ ] Event history (event sourcing)
  - [ ] Forward/backward compatibility
- [ ] Compatibility types:
  - [ ] Backward compatible (new consumer, old events)
  - [ ] Forward compatible (old consumer, new events)
  - [ ] Full compatibility (both)
- [ ] Safe changes:
  - [ ] Adding optional fields
  - [ ] Adding new event types
  - [ ] Deprecating fields (not removing)
- [ ] Breaking changes:
  - [ ] Removing fields
  - [ ] Changing field types
  - [ ] Renaming fields
  - [ ] Changing semantics
- [ ] Schema evolution strategies:
  - [ ] Tolerant reader pattern
  - [ ] Schema versioning
  - [ ] Event upcasting
  - [ ] Event transformation
- [ ] Schema registries:
  - [ ] Confluent Schema Registry
  - [ ] AWS Glue Schema Registry
  - [ ] Apicurio Registry
- [ ] Schema registry usage:
  - [ ] Schema registration
  - [ ] Compatibility checks
  - [ ] Schema evolution rules
  - [ ] CI/CD integration
- [ ] Serialization formats:
  - [ ] Avro (schema required)
  - [ ] Protobuf (schema required)
  - [ ] JSON Schema
  - [ ] Format selection criteria
- [ ] Handling event history:
  - [ ] Upcasting old events
  - [ ] Lazy vs eager migration
  - [ ] Snapshot rebuilding
- [ ] Set up schema registry workflow
- [ ] **ADR:** Document schema evolution strategy

---

### I.7 Event Ordering & Delivery

**Goal:** Understand ordering guarantees and delivery semantics

**Learning objectives:**
- Configure ordering guarantees appropriately
- Handle out-of-order events
- Implement exactly-once processing

**Tasks:**
- [ ] Create `experiments/event-ordering/`
- [ ] Ordering levels:
  - [ ] Total ordering (all events)
  - [ ] Partition ordering (within partition)
  - [ ] No ordering guarantees
- [ ] Kafka ordering:
  - [ ] Partition-level ordering
  - [ ] Partition key selection
  - [ ] Producer ordering guarantees
  - [ ] Consumer group ordering
- [ ] Delivery semantics:
  - [ ] At-most-once
  - [ ] At-least-once
  - [ ] Exactly-once (effectively once)
- [ ] Exactly-once in Kafka:
  - [ ] Idempotent producer
  - [ ] Transactional producer
  - [ ] Consumer read committed
- [ ] Out-of-order handling:
  - [ ] Event timestamps
  - [ ] Sequence numbers
  - [ ] Buffering and reordering
  - [ ] Late event handling
- [ ] Consumer design:
  - [ ] Idempotent consumers
  - [ ] Deduplication
  - [ ] Offset management
  - [ ] Error handling
- [ ] Partitioning strategies:
  - [ ] Entity-based partitioning
  - [ ] Round-robin (no ordering)
  - [ ] Custom partitioners
  - [ ] Partition count impact
- [ ] Consumer groups:
  - [ ] Parallel consumption
  - [ ] Rebalancing
  - [ ] Sticky assignment
  - [ ] Cooperative rebalancing
- [ ] Dead letter queues:
  - [ ] Poison message handling
  - [ ] DLQ processing
  - [ ] Retry strategies
  - [ ] Manual intervention
- [ ] Implement ordered event processing
- [ ] **ADR:** Document ordering requirements

---

### I.8 Event Mesh & CloudEvents

**Goal:** Understand event standards and multi-cluster eventing

**Learning objectives:**
- Use CloudEvents specification
- Implement event mesh patterns
- Enable cross-cluster eventing

**Tasks:**
- [ ] Create `experiments/event-mesh/`
- [ ] CloudEvents specification:
  - [ ] Standard event format
  - [ ] Required attributes (id, source, type, specversion)
  - [ ] Optional attributes (time, subject, datacontenttype)
  - [ ] Extension attributes
- [ ] CloudEvents benefits:
  - [ ] Interoperability
  - [ ] Tooling compatibility
  - [ ] Transport agnostic
  - [ ] Ecosystem support
- [ ] CloudEvents bindings:
  - [ ] HTTP binding
  - [ ] Kafka binding
  - [ ] AMQP binding
  - [ ] NATS binding
- [ ] Event mesh concepts:
  - [ ] Multi-cluster eventing
  - [ ] Event routing
  - [ ] Protocol translation
  - [ ] Event filtering
- [ ] Knative Eventing:
  - [ ] Brokers and triggers
  - [ ] Channels and subscriptions
  - [ ] Event sources
  - [ ] CloudEvents native
- [ ] Event routing:
  - [ ] Content-based routing
  - [ ] Header-based routing
  - [ ] Broker patterns
  - [ ] Topic hierarchies
- [ ] Cross-cluster eventing:
  - [ ] Kafka MirrorMaker
  - [ ] Event bridges
  - [ ] Geographic distribution
  - [ ] Latency considerations
- [ ] Event observability:
  - [ ] Event tracing
  - [ ] Event lineage
  - [ ] Event analytics
  - [ ] Dead letter monitoring
- [ ] Set up Knative Eventing
- [ ] **ADR:** Document event mesh architecture

---
