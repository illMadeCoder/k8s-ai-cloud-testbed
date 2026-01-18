## Phase 10: Messaging & Event Streaming

*Asynchronous communication patterns for event-driven architectures.*

### 10.0 Messaging Decision Framework

**Goal:** Understand messaging patterns and when to use each technology

**Learning objectives:**
- Compare messaging paradigms (queues vs streams vs pub/sub)
- Understand delivery guarantees and trade-offs
- Make informed messaging technology decisions

**Tasks:**
- [ ] Create `docs/messaging-comparison.md`
- [ ] Messaging paradigms:
  - [ ] Message queues (point-to-point, competing consumers)
  - [ ] Event streaming (log-based, replay capability)
  - [ ] Pub/sub (topic-based, fan-out)
  - [ ] Request/reply (synchronous over async)
- [ ] Technology comparison:
  - [ ] **Kafka:** Event streaming, high throughput, log retention
  - [ ] **RabbitMQ:** Traditional queuing, routing flexibility, protocols
  - [ ] **NATS:** Lightweight, low latency, cloud-native
  - [ ] **Cloud queues (SQS/Service Bus):** Managed, serverless integration
- [ ] Decision criteria:
  - [ ] Message ordering requirements
  - [ ] Delivery guarantees (at-most-once, at-least-once, exactly-once)
  - [ ] Throughput and latency requirements
  - [ ] Message replay needs
  - [ ] Operational complexity tolerance
- [ ] Use case mapping:
  - [ ] Event sourcing / CQRS → Kafka
  - [ ] Task queues / work distribution → RabbitMQ
  - [ ] Real-time microservice communication → NATS
  - [ ] Cloud-native serverless → SQS/Service Bus
- [ ] Anti-patterns:
  - [ ] Using Kafka for simple task queues
  - [ ] Using RabbitMQ for event sourcing
  - [ ] Over-engineering with messaging when HTTP suffices
- [ ] Document decision criteria
- [ ] **ADR:** Document messaging technology selection for this lab

*Note: Phase 12.2 provides detailed performance benchmarks after you've learned each system.*

---

### 10.1 Kafka with Strimzi

**Goal:** Deploy and operate Kafka on Kubernetes

**Learning objectives:**
- Understand Kafka architecture (brokers, topics, partitions, consumers)
- Use Strimzi operator for Kafka lifecycle
- Implement common messaging patterns

**Tasks:**
- [ ] Create `experiments/scenarios/kafka-tutorial/`
- [ ] Deploy Strimzi operator via ArgoCD
- [ ] Create Kafka cluster (KafkaCluster CRD)
- [ ] Configure:
  - [ ] Topics (KafkaTopic CRD)
  - [ ] Users and ACLs (KafkaUser CRD)
  - [ ] Replication and partitions
- [ ] Build producer/consumer apps:
  - [ ] Simple pub/sub
  - [ ] Consumer groups
  - [ ] Exactly-once semantics
- [ ] Implement patterns:
  - [ ] Event sourcing
  - [ ] CQRS with Kafka
  - [ ] Dead letter queue
- [ ] Monitoring:
  - [ ] Kafka metrics in Prometheus
  - [ ] Consumer lag monitoring
  - [ ] Grafana dashboards
- [ ] Connect (optional):
  - [ ] Kafka Connect for integrations
  - [ ] Source/sink connectors
- [ ] Document Kafka operational patterns

---

### 10.2 RabbitMQ with Operator

**Goal:** Deploy and operate RabbitMQ for task queues

**Learning objectives:**
- Understand RabbitMQ architecture (exchanges, queues, bindings)
- Use RabbitMQ Cluster Operator
- Compare with Kafka use cases

**Tasks:**
- [ ] Create `experiments/scenarios/rabbitmq-tutorial/`
- [ ] Deploy RabbitMQ Cluster Operator
- [ ] Create RabbitMQ cluster (RabbitmqCluster CRD)
- [ ] Configure:
  - [ ] Exchanges (direct, fanout, topic, headers)
  - [ ] Queues and bindings
  - [ ] Users and permissions
- [ ] Build producer/consumer apps:
  - [ ] Work queues (competing consumers)
  - [ ] Pub/sub (fanout)
  - [ ] Routing (topic exchange)
  - [ ] RPC pattern
- [ ] Implement reliability:
  - [ ] Publisher confirms
  - [ ] Consumer acknowledgments
  - [ ] Dead letter exchanges
  - [ ] Message TTL
- [ ] Monitoring:
  - [ ] RabbitMQ management UI
  - [ ] Prometheus metrics
  - [ ] Queue depth alerting
- [ ] Document RabbitMQ vs Kafka decision guide
- [ ] **ADR:** Document messaging technology selection

---

### 10.3 NATS & JetStream

**Goal:** Learn lightweight, high-performance messaging

**Learning objectives:**
- Understand NATS core vs JetStream
- Implement request-reply patterns
- Compare with Kafka and RabbitMQ

**Tasks:**
- [ ] Create `experiments/scenarios/nats-tutorial/`
- [ ] Deploy NATS with JetStream enabled
- [ ] Core NATS patterns:
  - [ ] Pub/sub (fire and forget)
  - [ ] Request/reply
  - [ ] Queue groups (load balancing)
- [ ] JetStream (persistence):
  - [ ] Streams and consumers
  - [ ] At-least-once delivery
  - [ ] Message replay
  - [ ] Key-value store
  - [ ] Object store
- [ ] Build demo apps showcasing each pattern
- [ ] Compare with Kafka/RabbitMQ:
  - [ ] Latency
  - [ ] Throughput
  - [ ] Operational complexity
  - [ ] Use case fit
- [ ] Document NATS patterns and when to use

---

### 10.4 Cloud Messaging with Crossplane

**Goal:** Abstract cloud message queues with Crossplane XRDs

**Learning objectives:**
- Use Crossplane for managed messaging services
- Create portable queue abstractions
- Compare managed vs self-hosted

**Tasks:**
- [ ] Create `experiments/scenarios/cloud-messaging/`
- [ ] Create XRD: SimpleQueue
  - [ ] Abstracts AWS SQS and Azure Service Bus
  - [ ] Common interface for both clouds
- [ ] Deploy same producer/consumer app to both clouds
- [ ] Compare:
  - [ ] Message visibility handling
  - [ ] Dead letter queue behavior
  - [ ] FIFO vs standard queues
  - [ ] Pricing models
- [ ] Test failover (queue in different region)
- [ ] Document cloud queue patterns

---

### 10.5 Distributed Coordination & ZooKeeper

**Goal:** Understand distributed coordination primitives and when to use them

**Learning objectives:**
- Understand ZooKeeper architecture and use cases
- Compare coordination systems (ZooKeeper vs etcd vs Consul)
- Implement common coordination patterns

**Tasks:**
- [ ] Create `experiments/scenarios/distributed-coordination/`
- [ ] ZooKeeper deep dive:
  - [ ] Deploy ZooKeeper ensemble (3+ nodes)
  - [ ] Understand znodes, watches, ephemeral nodes
  - [ ] Leader election pattern
  - [ ] Distributed locks
  - [ ] Configuration management
  - [ ] ZooKeeper with Kafka (legacy mode)
- [ ] etcd comparison:
  - [ ] Deploy etcd cluster
  - [ ] Key-value operations
  - [ ] Watch API
  - [ ] etcd as Kubernetes backing store
  - [ ] Compare with ZooKeeper use cases
- [ ] Consul comparison:
  - [ ] Deploy Consul cluster
  - [ ] Service discovery features
  - [ ] Key-value store
  - [ ] Connect (service mesh features)
  - [ ] Multi-datacenter capabilities
- [ ] Modern alternatives:
  - [ ] Kafka KRaft (ZooKeeper-less Kafka)
  - [ ] When to use coordination services vs embedded consensus
- [ ] Use case mapping:
  - [ ] Leader election → ZooKeeper/etcd
  - [ ] Service discovery → Consul/Kubernetes DNS
  - [ ] Configuration → etcd/Consul KV
  - [ ] Distributed locks → ZooKeeper/etcd
- [ ] Operational considerations:
  - [ ] Quorum and failure tolerance
  - [ ] Performance characteristics
  - [ ] Backup and recovery
  - [ ] Monitoring and alerting
- [ ] Document coordination patterns and selection criteria
- [ ] **ADR:** Document coordination service selection

---

### 10.6 Messaging Cost Optimization

**Goal:** Optimize messaging infrastructure costs

*FinOps consideration: Messaging systems like Kafka can be resource-intensive. Right-size brokers and implement retention policies.*

**Learning objectives:**
- Understand messaging cost drivers
- Implement retention and compaction policies
- Right-size messaging infrastructure

**Tasks:**
- [ ] Broker resource analysis:
  - [ ] CPU/memory/storage per broker
  - [ ] Cost comparison: Kafka vs RabbitMQ vs NATS
  - [ ] Self-managed vs cloud-managed cost comparison
- [ ] Storage optimization:
  - [ ] Retention policy cost impact
  - [ ] Log compaction for reduced storage
  - [ ] Tiered storage (hot/cold) for Kafka
- [ ] Throughput optimization:
  - [ ] Partition count vs resource cost
  - [ ] Consumer group efficiency
  - [ ] Batch size optimization
- [ ] Cloud messaging costs:
  - [ ] Per-message pricing analysis
  - [ ] Data transfer costs (cross-AZ, cross-region)
  - [ ] Self-managed breakeven analysis
- [ ] Right-sizing:
  - [ ] Broker count optimization
  - [ ] Instance type selection
  - [ ] Storage type selection (SSD vs HDD)
- [ ] Document messaging cost patterns
- [ ] **ADR:** Document messaging cost optimization strategy

---

