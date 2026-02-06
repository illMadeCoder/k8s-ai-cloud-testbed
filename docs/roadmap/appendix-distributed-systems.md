## Appendix: Distributed Systems Fundamentals

*The theoretical foundation that makes everything else make sense. Understanding these concepts deeply transforms how you design, debug, and operate distributed systems.*

### E.1 Consensus Algorithms

**Goal:** Understand how distributed systems agree on state

**Learning objectives:**
- Understand the consensus problem and why it's hard
- Learn Raft algorithm in depth
- Understand Paxos conceptually
- Know when consensus is needed vs overkill

**Tasks:**
- [ ] Create `experiments/consensus-algorithms/`
- [ ] The consensus problem:
  - [ ] Agreement, validity, termination properties
  - [ ] FLP impossibility result (async systems)
  - [ ] Why we need consensus (leader election, distributed locks, replicated state)
- [ ] Raft algorithm (understandable consensus):
  - [ ] Leader election (terms, votes, timeouts)
  - [ ] Log replication (AppendEntries, commit index)
  - [ ] Safety guarantees (leader completeness, state machine safety)
  - [ ] Cluster membership changes
  - [ ] Log compaction and snapshots
- [ ] Raft in practice:
  - [ ] etcd (Kubernetes backing store)
  - [ ] Consul
  - [ ] CockroachDB
  - [ ] Visualize Raft (raft.github.io)
- [ ] Paxos family:
  - [ ] Basic Paxos (single-decree)
  - [ ] Multi-Paxos (log replication)
  - [ ] Why Paxos is hard to understand/implement
  - [ ] Flexible Paxos (quorum intersection)
- [ ] Byzantine fault tolerance:
  - [ ] Byzantine generals problem
  - [ ] PBFT basics
  - [ ] When BFT matters (blockchain, adversarial environments)
  - [ ] Why most systems skip BFT (trusted environments)
- [ ] Consensus alternatives:
  - [ ] Viewstamped Replication
  - [ ] Zab (ZooKeeper)
  - [ ] EPaxos (leaderless)
- [ ] Build Raft implementation (simplified)
- [ ] **ADR:** Document when to use consensus vs simpler approaches

---

### E.2 CAP Theorem & Consistency Models

**Goal:** Understand the fundamental trade-offs in distributed data systems

**Learning objectives:**
- Understand CAP theorem and its implications
- Know different consistency models and their trade-offs
- Choose appropriate consistency for different use cases

**Tasks:**
- [ ] Create `experiments/consistency-models/`
- [ ] CAP theorem:
  - [ ] Consistency, Availability, Partition tolerance
  - [ ] Why you must choose (partition happens)
  - [ ] CP systems (sacrifice availability during partition)
  - [ ] AP systems (sacrifice consistency during partition)
  - [ ] CAP is about partitions, not normal operation
- [ ] PACELC extension:
  - [ ] If Partition: Availability vs Consistency
  - [ ] Else: Latency vs Consistency
  - [ ] More practical than CAP alone
- [ ] Consistency models spectrum:
  - [ ] Linearizability (strongest, "real-time")
  - [ ] Sequential consistency
  - [ ] Causal consistency
  - [ ] Eventual consistency (weakest)
  - [ ] Read-your-writes, monotonic reads
- [ ] Linearizability deep dive:
  - [ ] Definition (operations appear atomic, real-time order)
  - [ ] Cost (coordination, latency)
  - [ ] When you need it (counters, locks, leader election)
- [ ] Eventual consistency deep dive:
  - [ ] Definition (replicas converge eventually)
  - [ ] Conflict resolution (last-write-wins, CRDTs, application logic)
  - [ ] Anti-entropy protocols
  - [ ] Read repair, hinted handoff
- [ ] CRDTs (Conflict-free Replicated Data Types):
  - [ ] State-based vs operation-based
  - [ ] G-Counter, PN-Counter
  - [ ] G-Set, OR-Set
  - [ ] LWW-Register
  - [ ] When CRDTs help vs complicate
- [ ] Consistency in practice:
  - [ ] DynamoDB (eventual, strong read option)
  - [ ] Cassandra (tunable consistency)
  - [ ] Spanner (external consistency)
  - [ ] PostgreSQL (linearizable single-node)
- [ ] **ADR:** Document consistency model selection criteria

---

### E.3 Distributed Transactions

**Goal:** Understand how to maintain consistency across multiple services/databases

**Learning objectives:**
- Understand two-phase commit and its limitations
- Implement Saga patterns for distributed transactions
- Know when to avoid distributed transactions entirely

**Tasks:**
- [ ] Create `experiments/distributed-transactions/`
- [ ] Why distributed transactions are hard:
  - [ ] Atomicity across boundaries
  - [ ] Partial failure scenarios
  - [ ] Coordinator failures
- [ ] Two-Phase Commit (2PC):
  - [ ] Prepare phase (voting)
  - [ ] Commit phase (decision)
  - [ ] Coordinator role
  - [ ] Blocking problem (coordinator failure)
  - [ ] Performance implications
- [ ] Three-Phase Commit (3PC):
  - [ ] Pre-commit phase addition
  - [ ] Non-blocking under certain failures
  - [ ] Why rarely used in practice
- [ ] XA transactions:
  - [ ] XA protocol standard
  - [ ] Database support (PostgreSQL, MySQL)
  - [ ] Message queue support (JMS)
  - [ ] Performance and operational costs
- [ ] Saga pattern:
  - [ ] Sequence of local transactions
  - [ ] Compensating transactions for rollback
  - [ ] Choreography vs Orchestration
- [ ] Choreography-based Saga:
  - [ ] Events trigger next steps
  - [ ] Decentralized coordination
  - [ ] Harder to track/debug
  - [ ] Better for simple flows
- [ ] Orchestration-based Saga:
  - [ ] Central coordinator (saga orchestrator)
  - [ ] Explicit workflow definition
  - [ ] Easier to understand and modify
  - [ ] Single point of coordination
- [ ] Saga implementation patterns:
  - [ ] Temporal.io for orchestration
  - [ ] Event-driven choreography
  - [ ] State machines
  - [ ] Idempotency requirements
- [ ] Compensation design:
  - [ ] Semantic rollback (not always exact inverse)
  - [ ] Forward recovery vs backward recovery
  - [ ] Compensation ordering
- [ ] Alternatives to distributed transactions:
  - [ ] Design to avoid them (single database)
  - [ ] Eventual consistency acceptance
  - [ ] Reservation pattern
  - [ ] Try-Confirm-Cancel (TCC)
- [ ] Implement Saga with Temporal
- [ ] **ADR:** Document distributed transaction strategy

---

### E.4 Time & Clocks in Distributed Systems

**Goal:** Understand the challenges of time in distributed systems

**Learning objectives:**
- Understand why physical clocks are unreliable
- Implement logical clocks for ordering
- Know hybrid approaches used in practice

**Tasks:**
- [ ] Create `experiments/distributed-time/`
- [ ] Physical clock problems:
  - [ ] Clock drift and skew
  - [ ] NTP accuracy limits
  - [ ] Leap seconds
  - [ ] Why you can't trust timestamps across machines
- [ ] Logical clocks:
  - [ ] Happened-before relation (Lamport)
  - [ ] Lamport timestamps
  - [ ] Limitations (concurrent events)
- [ ] Vector clocks:
  - [ ] Tracking causality per node
  - [ ] Detecting concurrent events
  - [ ] Size growth problem
  - [ ] Dynamo-style conflict detection
- [ ] Version vectors:
  - [ ] Similar to vector clocks
  - [ ] Used for replica synchronization
  - [ ] Dotted version vectors (Riak)
- [ ] Hybrid Logical Clocks (HLC):
  - [ ] Combining physical and logical
  - [ ] CockroachDB implementation
  - [ ] Bounded clock skew assumption
- [ ] TrueTime (Google Spanner):
  - [ ] GPS and atomic clocks
  - [ ] Bounded uncertainty intervals
  - [ ] Wait-out uncertainty for consistency
  - [ ] Why most can't replicate this
- [ ] Ordering guarantees:
  - [ ] Total order vs partial order
  - [ ] FIFO ordering
  - [ ] Causal ordering
  - [ ] Total order broadcast
- [ ] Practical implications:
  - [ ] Event ordering in logs
  - [ ] Conflict resolution timestamps
  - [ ] Debugging distributed systems
  - [ ] Distributed tracing correlation
- [ ] **ADR:** Document time/ordering approach

---

### E.5 Failure Detection & Membership

**Goal:** Understand how distributed systems detect failures and manage group membership

**Learning objectives:**
- Understand failure detection challenges
- Implement membership protocols
- Handle partial failures gracefully

**Tasks:**
- [ ] Create `experiments/failure-detection/`
- [ ] Failure models:
  - [ ] Crash failures (stop responding)
  - [ ] Omission failures (dropped messages)
  - [ ] Timing failures (too slow)
  - [ ] Byzantine failures (arbitrary behavior)
- [ ] Failure detection challenges:
  - [ ] Can't distinguish slow from dead
  - [ ] False positives vs false negatives
  - [ ] Network partitions
- [ ] Heartbeat-based detection:
  - [ ] Push vs pull heartbeats
  - [ ] Timeout configuration
  - [ ] Cascading false positives
- [ ] Phi Accrual Failure Detector:
  - [ ] Probabilistic suspicion level
  - [ ] Adaptive to network conditions
  - [ ] Used in Akka, Cassandra
- [ ] SWIM protocol:
  - [ ] Scalable Weakly-consistent Infection-style Membership
  - [ ] Probe and probe-request
  - [ ] Infection-style dissemination
  - [ ] Used in Consul, Memberlist
- [ ] Gossip protocols:
  - [ ] Epidemic-style information spread
  - [ ] Probabilistic guarantees
  - [ ] Anti-entropy vs rumor mongering
  - [ ] Cracking gossip (Serf, Memberlist)
- [ ] Membership views:
  - [ ] Consistent vs eventually consistent membership
  - [ ] View changes
  - [ ] Virtual synchrony
- [ ] Split-brain handling:
  - [ ] Detecting partition
  - [ ] Quorum-based decisions
  - [ ] Fencing (STONITH)
  - [ ] Merge after partition heals
- [ ] Lease-based approaches:
  - [ ] Time-bounded leadership
  - [ ] Lease renewal
  - [ ] Lease expiration handling
- [ ] **ADR:** Document failure detection configuration

---

### E.6 Replication Strategies

**Goal:** Understand data replication patterns and trade-offs

**Learning objectives:**
- Understand different replication topologies
- Know consistency implications of each approach
- Choose appropriate replication for different needs

**Tasks:**
- [ ] Create `experiments/replication-strategies/`
- [ ] Why replicate:
  - [ ] Availability (survive failures)
  - [ ] Performance (read scaling, latency)
  - [ ] Data locality
- [ ] Single-leader replication:
  - [ ] All writes go to leader
  - [ ] Leader replicates to followers
  - [ ] Synchronous vs asynchronous replication
  - [ ] Failover (manual, automatic)
  - [ ] Split-brain prevention
- [ ] Replication lag:
  - [ ] Reading from follower inconsistencies
  - [ ] Read-your-writes guarantees
  - [ ] Monotonic reads
  - [ ] Consistent prefix reads
- [ ] Multi-leader replication:
  - [ ] Multiple nodes accept writes
  - [ ] Conflict detection and resolution
  - [ ] Use cases (multi-datacenter, offline clients)
  - [ ] Conflict resolution strategies
- [ ] Leaderless replication:
  - [ ] Any node accepts reads/writes
  - [ ] Quorum reads and writes
  - [ ] Sloppy quorums and hinted handoff
  - [ ] Read repair and anti-entropy
- [ ] Quorum systems:
  - [ ] R + W > N for consistency
  - [ ] Tunable consistency
  - [ ] Trade-offs (availability vs consistency)
- [ ] Chain replication:
  - [ ] Linear chain of replicas
  - [ ] Strong consistency
  - [ ] CRAQ (Chain Replication with Apportioned Queries)
- [ ] Replication in practice:
  - [ ] PostgreSQL streaming replication
  - [ ] MySQL group replication
  - [ ] MongoDB replica sets
  - [ ] Cassandra (leaderless)
  - [ ] Kafka replication
- [ ] **ADR:** Document replication topology selection

---

### E.7 Partitioning & Sharding

**Goal:** Understand how to distribute data across multiple nodes

**Learning objectives:**
- Understand partitioning strategies
- Handle partition rebalancing
- Design for partition-aware queries

**Tasks:**
- [ ] Create `experiments/partitioning/`
- [ ] Why partition:
  - [ ] Data too large for single node
  - [ ] Throughput beyond single node
  - [ ] Distributing load
- [ ] Partitioning strategies:
  - [ ] Range partitioning (key ranges)
  - [ ] Hash partitioning (hash of key)
  - [ ] Consistent hashing (ring-based)
  - [ ] Directory-based partitioning
- [ ] Range partitioning:
  - [ ] Natural ordering preserved
  - [ ] Range queries efficient
  - [ ] Hot spots on sequential keys
  - [ ] Partition boundaries
- [ ] Hash partitioning:
  - [ ] Even distribution (with good hash)
  - [ ] Range queries inefficient
  - [ ] Hash function selection
- [ ] Consistent hashing:
  - [ ] Virtual nodes
  - [ ] Minimal redistribution on changes
  - [ ] Load balancing challenges
  - [ ] Used in: Cassandra, DynamoDB, Riak
- [ ] Partition rebalancing:
  - [ ] Fixed partitions (hash mod N)
  - [ ] Dynamic partitioning (split/merge)
  - [ ] Rebalancing without downtime
  - [ ] Data movement costs
- [ ] Secondary indexes:
  - [ ] Local indexes (partition-local)
  - [ ] Global indexes (partitioned separately)
  - [ ] Scatter-gather queries
- [ ] Cross-partition operations:
  - [ ] Distributed joins
  - [ ] Cross-partition transactions
  - [ ] Denormalization to avoid
- [ ] Hot spots and skew:
  - [ ] Celebrity problem
  - [ ] Salting keys
  - [ ] Application-level sharding
- [ ] Partitioning in practice:
  - [ ] Vitess (MySQL sharding)
  - [ ] Citus (PostgreSQL sharding)
  - [ ] CockroachDB (automatic)
  - [ ] MongoDB sharding
- [ ] **ADR:** Document partitioning strategy

---

### E.8 Distributed Locking & Coordination

**Goal:** Understand coordination primitives in distributed systems

**Learning objectives:**
- Implement distributed locks correctly
- Use coordination services effectively
- Know the pitfalls of distributed coordination

**Tasks:**
- [ ] Create `experiments/distributed-locking/`
- [ ] Why distributed locks are hard:
  - [ ] No global time
  - [ ] Partial failures
  - [ ] Network partitions
  - [ ] Process pauses (GC, paging)
- [ ] Lock requirements:
  - [ ] Mutual exclusion (safety)
  - [ ] Deadlock freedom (liveness)
  - [ ] Fault tolerance
- [ ] Naive approaches and failures:
  - [ ] SET NX (Redis) problems
  - [ ] Lock expiration race conditions
  - [ ] Process pause scenarios
- [ ] Redlock algorithm:
  - [ ] Multi-instance Redis locking
  - [ ] Quorum-based approach
  - [ ] Criticisms (Martin Kleppmann analysis)
  - [ ] When Redlock is/isn't appropriate
- [ ] Fencing tokens:
  - [ ] Monotonic token with lock
  - [ ] Resource validates token
  - [ ] Prevents stale lock holder operations
- [ ] ZooKeeper recipes:
  - [ ] Sequential ephemeral nodes
  - [ ] Watch for predecessor
  - [ ] Leader election pattern
  - [ ] Distributed barriers
- [ ] etcd coordination:
  - [ ] Lease-based locking
  - [ ] Election API
  - [ ] Watch for changes
- [ ] Chubby (Google):
  - [ ] Lock service with consensus
  - [ ] Sessions and keepalives
  - [ ] Sequencer for fencing
- [ ] When to avoid distributed locks:
  - [ ] Idempotent operations instead
  - [ ] Optimistic concurrency
  - [ ] Single-writer patterns
  - [ ] Eventual consistency acceptance
- [ ] Coordination services:
  - [ ] ZooKeeper
  - [ ] etcd
  - [ ] Consul
  - [ ] When to use vs build your own
- [ ] **ADR:** Document coordination approach

---

### E.9 Exactly-Once & Idempotency

**Goal:** Understand message delivery semantics and idempotent design

**Learning objectives:**
- Understand delivery guarantees (at-most-once, at-least-once, exactly-once)
- Implement idempotent operations
- Design systems that handle duplicates gracefully

**Tasks:**
- [ ] Create `experiments/exactly-once/`
- [ ] Delivery semantics:
  - [ ] At-most-once (fire and forget)
  - [ ] At-least-once (retry until ack)
  - [ ] Exactly-once (the holy grail)
- [ ] Why exactly-once is hard:
  - [ ] Two Generals Problem
  - [ ] Can't distinguish lost message from lost ack
  - [ ] Network is unreliable
- [ ] Exactly-once semantics:
  - [ ] It's really "effectively once"
  - [ ] Deduplication at receiver
  - [ ] Idempotent processing
- [ ] Idempotency fundamentals:
  - [ ] Same operation, same result
  - [ ] Natural idempotency (SET vs INCREMENT)
  - [ ] Designing for idempotency
- [ ] Idempotency keys:
  - [ ] Client-generated unique IDs
  - [ ] Server-side deduplication
  - [ ] Key storage and expiration
  - [ ] Stripe API pattern
- [ ] Deduplication strategies:
  - [ ] Idempotency key lookup
  - [ ] Message ID tracking
  - [ ] Bloom filters for efficiency
  - [ ] Time-windowed deduplication
- [ ] Outbox pattern:
  - [ ] Write to outbox in same transaction
  - [ ] Separate process publishes
  - [ ] Guaranteed delivery with dedup
- [ ] Transactional messaging:
  - [ ] Kafka exactly-once semantics
  - [ ] Transactions across produce/consume
  - [ ] Read-committed isolation
- [ ] Idempotent consumers:
  - [ ] Processing with idempotency key
  - [ ] Storing processed message IDs
  - [ ] Exactly-once in stream processing
- [ ] Practical patterns:
  - [ ] Request IDs in APIs
  - [ ] ETags for updates
  - [ ] Version numbers
  - [ ] Conditional updates
- [ ] **ADR:** Document idempotency implementation

---
