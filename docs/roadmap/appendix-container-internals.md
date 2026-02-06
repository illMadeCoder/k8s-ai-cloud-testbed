## Appendix: Container & Runtime Internals

*Deep dive into how containers actually work. Understanding Linux primitives, OCI specifications, and runtime internals transforms containers from magic to comprehensible technology.*

### G.1 Linux Namespaces

**Goal:** Understand the isolation primitives that make containers possible

**Learning objectives:**
- Understand each namespace type and its purpose
- Create namespaces manually
- Debug namespace-related issues

**Tasks:**
- [ ] Create `experiments/linux-namespaces/`
- [ ] Namespace fundamentals:
  - [ ] Process isolation mechanism
  - [ ] Each namespace type isolates specific resources
  - [ ] Namespaces are hierarchical
- [ ] PID namespace:
  - [ ] Process ID isolation
  - [ ] PID 1 in containers
  - [ ] init process responsibilities
  - [ ] Zombie process reaping
- [ ] Network namespace:
  - [ ] Network stack isolation
  - [ ] Virtual ethernet pairs (veth)
  - [ ] Container networking basics
  - [ ] Host network mode
- [ ] Mount namespace:
  - [ ] Filesystem view isolation
  - [ ] Pivot root for container root filesystem
  - [ ] Bind mounts
  - [ ] Propagation modes (private, shared, slave)
- [ ] UTS namespace:
  - [ ] Hostname isolation
  - [ ] Domain name isolation
  - [ ] Why containers have different hostnames
- [ ] IPC namespace:
  - [ ] System V IPC isolation
  - [ ] POSIX message queues
  - [ ] Shared memory isolation
- [ ] User namespace:
  - [ ] UID/GID mapping
  - [ ] Rootless containers foundation
  - [ ] Capabilities in user namespaces
  - [ ] Security implications
- [ ] Cgroup namespace:
  - [ ] Cgroup hierarchy visibility
  - [ ] Resource limit isolation view
- [ ] Time namespace (newer):
  - [ ] Clock isolation
  - [ ] Use cases (testing, migration)
- [ ] Working with namespaces:
  - [ ] unshare command
  - [ ] nsenter command
  - [ ] /proc/[pid]/ns/ inspection
  - [ ] setns() syscall
- [ ] Build minimal container from namespaces
- [ ] **ADR:** Document namespace security considerations

---

### G.2 Control Groups (cgroups)

**Goal:** Understand resource limiting and accounting

**Learning objectives:**
- Understand cgroups v1 vs v2
- Configure resource limits
- Monitor resource usage via cgroups

**Tasks:**
- [ ] Create `experiments/cgroups/`
- [ ] Cgroups fundamentals:
  - [ ] Resource limiting (CPU, memory, I/O)
  - [ ] Resource accounting
  - [ ] Process grouping
  - [ ] Hierarchical organization
- [ ] Cgroups v1:
  - [ ] Multiple hierarchies
  - [ ] Controller-specific trees
  - [ ] /sys/fs/cgroup/ structure
  - [ ] Why v1 is complex
- [ ] Cgroups v2:
  - [ ] Unified hierarchy
  - [ ] Single tree for all controllers
  - [ ] Improved resource distribution
  - [ ] Modern Linux default
- [ ] CPU controller:
  - [ ] cpu.shares (relative weight)
  - [ ] cpu.cfs_quota_us / cpu.cfs_period_us (hard limits)
  - [ ] cpu.max (v2 combined setting)
  - [ ] CPU pinning (cpuset)
- [ ] Memory controller:
  - [ ] memory.limit_in_bytes (v1) / memory.max (v2)
  - [ ] memory.soft_limit (v1) / memory.high (v2)
  - [ ] OOM killer behavior
  - [ ] Memory accounting (RSS, cache, swap)
- [ ] I/O controller:
  - [ ] blkio (v1) / io (v2)
  - [ ] Bandwidth limiting
  - [ ] IOPS limiting
  - [ ] Weight-based sharing
- [ ] PIDs controller:
  - [ ] Process count limiting
  - [ ] Fork bomb prevention
- [ ] Kubernetes and cgroups:
  - [ ] Pod cgroups structure
  - [ ] QoS classes and cgroups
  - [ ] Burstable vs Guaranteed
  - [ ] Resource requests vs limits
- [ ] Cgroups inspection:
  - [ ] /sys/fs/cgroup/ exploration
  - [ ] systemd-cgls
  - [ ] cgroupfs driver vs systemd
- [ ] Configure resource limits manually
- [ ] **ADR:** Document cgroups v2 migration considerations

---

### G.3 Container Runtimes

**Goal:** Understand the container runtime landscape

**Learning objectives:**
- Understand OCI runtime specification
- Know the role of high-level vs low-level runtimes
- Debug runtime issues

**Tasks:**
- [ ] Create `experiments/container-runtimes/`
- [ ] Runtime landscape overview:
  - [ ] High-level runtimes (containerd, CRI-O)
  - [ ] Low-level runtimes (runc, crun, gVisor)
  - [ ] CRI (Container Runtime Interface)
- [ ] OCI Runtime Specification:
  - [ ] config.json structure
  - [ ] Filesystem bundle
  - [ ] Lifecycle operations (create, start, kill, delete)
  - [ ] State management
- [ ] runc:
  - [ ] Reference OCI implementation
  - [ ] Written in Go
  - [ ] Used by containerd and CRI-O
  - [ ] runc spec, runc create, runc start
- [ ] containerd:
  - [ ] Industry-standard high-level runtime
  - [ ] Image management
  - [ ] Container lifecycle
  - [ ] Snapshotter architecture
  - [ ] ctr and crictl commands
- [ ] CRI-O:
  - [ ] Kubernetes-native runtime
  - [ ] Minimal footprint
  - [ ] OCI-only focus
  - [ ] Comparison with containerd
- [ ] Alternative low-level runtimes:
  - [ ] crun (C, faster startup)
  - [ ] youki (Rust)
  - [ ] Performance comparisons
- [ ] gVisor (runsc):
  - [ ] Application kernel in userspace
  - [ ] Syscall interception
  - [ ] Security isolation
  - [ ] Performance trade-offs
- [ ] Kata Containers:
  - [ ] Lightweight VMs
  - [ ] Hardware isolation
  - [ ] OCI compatibility
  - [ ] Use cases (multi-tenancy, untrusted workloads)
- [ ] Firecracker:
  - [ ] MicroVM technology
  - [ ] AWS Lambda foundation
  - [ ] Fast boot times
- [ ] Runtime selection:
  - [ ] Security requirements
  - [ ] Performance needs
  - [ ] Compatibility
- [ ] Compare runtimes hands-on
- [ ] **ADR:** Document runtime selection

---

### G.4 Container Images & OCI Image Spec

**Goal:** Understand container image format and distribution

**Learning objectives:**
- Understand OCI image specification
- Optimize image builds
- Work with image registries

**Tasks:**
- [ ] Create `experiments/container-images/`
- [ ] OCI Image Specification:
  - [ ] Image manifest
  - [ ] Image index (multi-arch)
  - [ ] Layer format (tar+gzip)
  - [ ] Config blob
- [ ] Image layers:
  - [ ] Union filesystem concept
  - [ ] Layer stacking
  - [ ] Copy-on-write
  - [ ] Layer deduplication
- [ ] Dockerfile optimization:
  - [ ] Layer caching
  - [ ] Multi-stage builds
  - [ ] Ordering instructions (cache efficiency)
  - [ ] Minimizing layer size
- [ ] Base image selection:
  - [ ] Distroless images
  - [ ] Alpine considerations
  - [ ] Scratch images
  - [ ] Security vs convenience
- [ ] Image distribution:
  - [ ] OCI Distribution Specification
  - [ ] Registry API
  - [ ] Content-addressable storage
  - [ ] Manifest lists for multi-arch
- [ ] Snapshotters:
  - [ ] overlayfs (default)
  - [ ] native
  - [ ] stargz (lazy pulling)
  - [ ] Performance characteristics
- [ ] Lazy pulling:
  - [ ] eStargz format
  - [ ] SOCI (AWS)
  - [ ] Faster container starts
- [ ] Image security:
  - [ ] Vulnerability scanning
  - [ ] Image signing (cosign, Notary)
  - [ ] SBOM generation
  - [ ] Provenance attestations
- [ ] Build tools:
  - [ ] Docker build
  - [ ] BuildKit
  - [ ] Kaniko (in-cluster)
  - [ ] Buildah
  - [ ] ko (Go applications)
- [ ] Build and analyze images
- [ ] **ADR:** Document image build strategy

---

### G.5 Container Security Primitives

**Goal:** Understand security mechanisms for containers

**Learning objectives:**
- Configure seccomp profiles
- Understand Linux capabilities
- Implement security contexts

**Tasks:**
- [ ] Create `experiments/container-security/`
- [ ] Linux capabilities:
  - [ ] Root privilege decomposition
  - [ ] Common capabilities (NET_ADMIN, SYS_PTRACE, etc.)
  - [ ] Default container capabilities
  - [ ] Capability dropping
  - [ ] Capability adding (when necessary)
- [ ] Seccomp (Secure Computing):
  - [ ] Syscall filtering
  - [ ] Default Docker/containerd profile
  - [ ] Custom seccomp profiles
  - [ ] Audit mode for profile development
- [ ] AppArmor:
  - [ ] Mandatory Access Control
  - [ ] Profile structure
  - [ ] Container-specific profiles
  - [ ] Default profiles
- [ ] SELinux:
  - [ ] Type enforcement
  - [ ] MCS (Multi-Category Security)
  - [ ] Container SELinux labels
  - [ ] spc_t (super privileged container)
- [ ] Read-only root filesystem:
  - [ ] Immutable container filesystem
  - [ ] tmpfs for writable paths
  - [ ] Implementation patterns
- [ ] No-new-privileges:
  - [ ] Preventing privilege escalation
  - [ ] setuid/setgid blocking
  - [ ] When to enable
- [ ] User namespaces (rootless):
  - [ ] Running as non-root
  - [ ] UID mapping
  - [ ] Security benefits
  - [ ] Compatibility considerations
- [ ] Kubernetes security context:
  - [ ] runAsUser, runAsGroup
  - [ ] fsGroup
  - [ ] allowPrivilegeEscalation
  - [ ] Pod vs container level
- [ ] Kubernetes Pod Security Standards:
  - [ ] Privileged, Baseline, Restricted
  - [ ] Pod Security Admission
  - [ ] Migration from PodSecurityPolicy
- [ ] Create hardened container configuration
- [ ] **ADR:** Document container security baseline

---

### G.6 Container Networking Internals

**Goal:** Understand how container networking works

**Learning objectives:**
- Understand Linux networking primitives for containers
- Debug container networking issues
- Know CNI plugin architecture

**Tasks:**
- [ ] Create `experiments/container-networking/`
- [ ] Linux networking primitives:
  - [ ] Network namespaces
  - [ ] Virtual ethernet pairs (veth)
  - [ ] Linux bridges
  - [ ] iptables/nftables
  - [ ] IPVS
- [ ] Basic container networking:
  - [ ] Bridge mode
  - [ ] Host mode
  - [ ] None mode
  - [ ] Container mode (shared namespace)
- [ ] CNI (Container Network Interface):
  - [ ] CNI specification
  - [ ] Plugin architecture
  - [ ] ADD/DEL operations
  - [ ] Chained plugins
- [ ] CNI plugins:
  - [ ] Bridge plugin
  - [ ] IPAM plugins (host-local, dhcp)
  - [ ] Portmap plugin
  - [ ] Bandwidth plugin
- [ ] Advanced CNI implementations:
  - [ ] Calico (BGP-based)
  - [ ] Cilium (eBPF-based)
  - [ ] Flannel (overlay)
  - [ ] Weave
- [ ] Overlay networking:
  - [ ] VXLAN encapsulation
  - [ ] Geneve
  - [ ] Performance overhead
  - [ ] MTU considerations
- [ ] eBPF networking:
  - [ ] Dataplane acceleration
  - [ ] Bypassing iptables
  - [ ] Cilium architecture
  - [ ] Socket-level load balancing
- [ ] Service networking:
  - [ ] kube-proxy modes (iptables, IPVS, eBPF)
  - [ ] ClusterIP implementation
  - [ ] NodePort implementation
  - [ ] LoadBalancer implementation
- [ ] DNS in containers:
  - [ ] /etc/resolv.conf injection
  - [ ] CoreDNS configuration
  - [ ] ndots and search domains
- [ ] Debug networking issues hands-on
- [ ] **ADR:** Document CNI selection criteria

---

### G.7 Container Storage

**Goal:** Understand container storage mechanisms

**Learning objectives:**
- Understand storage drivers and their trade-offs
- Configure persistent storage for containers
- Debug storage performance issues

**Tasks:**
- [ ] Create `experiments/container-storage/`
- [ ] Union filesystems:
  - [ ] Copy-on-write concept
  - [ ] Layer stacking
  - [ ] Write performance implications
- [ ] Storage drivers:
  - [ ] overlay2 (recommended)
  - [ ] devicemapper (legacy)
  - [ ] btrfs
  - [ ] zfs
  - [ ] Performance characteristics
- [ ] overlayfs deep dive:
  - [ ] Lower and upper directories
  - [ ] Work directory
  - [ ] Whiteout files
  - [ ] Opaque directories
- [ ] Volume types:
  - [ ] Named volumes
  - [ ] Bind mounts
  - [ ] tmpfs mounts
  - [ ] Volume drivers
- [ ] CSI (Container Storage Interface):
  - [ ] CSI specification
  - [ ] Controller vs node plugins
  - [ ] Identity service
  - [ ] Volume lifecycle
- [ ] Kubernetes storage:
  - [ ] PersistentVolume and PersistentVolumeClaim
  - [ ] StorageClass and dynamic provisioning
  - [ ] Access modes (RWO, ROX, RWX)
  - [ ] Volume expansion
- [ ] Storage performance:
  - [ ] I/O isolation (cgroups)
  - [ ] Storage driver overhead
  - [ ] Local vs network storage
  - [ ] IOPS and throughput considerations
- [ ] Ephemeral storage:
  - [ ] emptyDir volumes
  - [ ] Container writable layer limits
  - [ ] Ephemeral storage requests/limits
- [ ] Storage security:
  - [ ] fsGroup and supplemental groups
  - [ ] SELinux labels for volumes
  - [ ] Encryption at rest
- [ ] Debug storage issues
- [ ] **ADR:** Document storage driver selection

---

### G.8 Rootless & Unprivileged Containers

**Goal:** Run containers without root privileges

**Learning objectives:**
- Understand rootless container architecture
- Configure rootless container runtimes
- Know limitations and workarounds

**Tasks:**
- [ ] Create `experiments/rootless-containers/`
- [ ] Why rootless:
  - [ ] Security (no root on host)
  - [ ] Multi-tenancy
  - [ ] Reduced attack surface
  - [ ] Container escape mitigation
- [ ] User namespace mapping:
  - [ ] /etc/subuid and /etc/subgid
  - [ ] UID ranges for users
  - [ ] newuidmap and newgidmap
- [ ] Rootless containerd:
  - [ ] containerd-rootless-setuptool.sh
  - [ ] XDG_RUNTIME_DIR
  - [ ] Socket location
- [ ] Rootless Podman:
  - [ ] Native rootless support
  - [ ] Configuration differences
  - [ ] Storage location
- [ ] Rootless Docker:
  - [ ] dockerd-rootless-setuptool.sh
  - [ ] Docker context management
  - [ ] Socket location
- [ ] Kubernetes rootless:
  - [ ] Rootless kind
  - [ ] Usernetes
  - [ ] Sysbox runtime
- [ ] Networking in rootless:
  - [ ] slirp4netns (userspace networking)
  - [ ] pasta (more performant)
  - [ ] Port forwarding limitations
  - [ ] Performance implications
- [ ] Storage in rootless:
  - [ ] fuse-overlayfs
  - [ ] Native overlay (kernel 5.11+)
  - [ ] Performance comparison
- [ ] Limitations:
  - [ ] Privileged ports (<1024)
  - [ ] Certain syscalls
  - [ ] Network performance
  - [ ] Compatibility issues
- [ ] Set up rootless container environment
- [ ] **ADR:** Document rootless adoption strategy

---
