# Planning Documents for Roadmap Review

**Branch:** `claude/review-project-roadmap-psMLb`
**Status:** Planning in progress - not yet committed to main roadmap
**Date:** 2026-01-17

---

## Overview

This directory contains planning documents and proposals being considered for the roadmap consolidation. These are **working documents** that will be refined before being committed to the actual roadmap.

---

## Documents

### 1. Advanced Metrics & eBPF Strategy
**File:** `advanced-metrics-ebpf-strategy.md`

**Summary:** Proposal to expand metrics beyond CPU/RAM to include I/O and network metrics using eBPF tools.

**Key Points:**
- Current metrics (CPU/RAM) miss critical bottlenecks
- I/O metrics: Disk latency, IOPS, throughput, queue depth
- Network metrics: TCP retransmits, socket buffers, connection overhead
- eBPF tools: biosnoop, tcptop, tcpretrans, cachestat
- Integration with Pixie, Parca, Tetragon

**Impact on Roadmap:**
- Phase 3: Add eBPF & System Metrics sub-phase
- Phase 5: Enhance database benchmark with I/O analysis
- Phase 7: Add network I/O overhead to mesh benchmark
- Phase 10: Full stack I/O attribution in capstone

**Questions to Answer:**
- Should Pixie be Phase 3.6 or separate phase?
- How much eBPF in core vs appendix?
- Should we have an "eBPF Deep Dive" appendix?

**Status:** ✅ Drafted, awaiting review

---

### 2. The "Chi" Observability Stack
**File:** `chi-observability-stack.md`

**Summary:** Philosophical framework for visualizing traffic as energy flow through service meshes, with security as identity verification.

**Key Concepts:**
- **Traffic = Energy flow** (not just requests/second)
- **Latency = Resistance** (friction in the system)
- **Queue depth = Energy reservoirs** (backup/pressure)
- **CPU = Heat** (byproduct, not the primary constraint)
- **Service Mesh = Distributed sensors + valves + armor**

**4 Phases:**
1. **Glass Window** - Visualize flow with Hubble/Pixie
2. **Gauge** - Measure friction with USE Method (Utilization, Saturation, Errors)
3. **Valve & Armor** - Control flow and prove identity with Linkerd
4. **Federation** - Multi-cluster trust boundaries

**Impact on Roadmap:**
- Phase 7: Enhanced Service Mesh section with Chi framework
- Integration with eBPF strategy (complementary)
- New experiments: chi-glass-window, chi-gauge-saturation, chi-valve-smart-routing
- FinOps: Mesh ROI analysis (cost vs benefits)

**Questions to Answer:**
- Should Chi be Phase 7 or separate phase 7.5?
- Which mesh for Chi lab? (Linkerd vs Istio vs Cilium)
- How much multi-cluster in core vs appendix?

**Status:** ✅ Drafted, awaiting review

---

## Main Roadmap Documents (In Branch Root)

These have already been created and pushed:

1. **`docs/strategic-review-2026-01.md`**
   - Initial strategic assessment
   - Options analysis (A, B, C)
   - Questions for decision

2. **`docs/roadmap-consolidation-analysis.md`** ⭐ **MAIN DOCUMENT**
   - Detailed phase-by-phase analysis
   - 10-phase structure with FinOps integration
   - Dependency mapping
   - Migration plan

3. **`docs/roadmap-final-structure.md`**
   - Complete 10-phase roadmap
   - AI-powered tech discovery
   - Timeline and success metrics

4. **`docs/roadmap-visual-summary.md`**
   - Visual diagrams
   - FinOps integration examples
   - Timeline breakdown

5. **`docs/roadmap-consolidation-summary.md`**
   - Quick before/after comparison
   - Visual structure changes

---

## Decisions Made So Far

### ✅ Confirmed
1. **16 phases → 10 core phases** (38% reduction)
2. **FinOps as first-class metric** in every phase
3. **Phase 10 as capstone** (Performance & Cost Engineering)
4. **AI-powered tech discovery** post-Phase 10
5. **Moved to appendices:**
   - Deployment Strategies (Appendix G)
   - gRPC Deep Dive (Appendix H)
   - Chaos Engineering (Appendix P)
   - Advanced Workflows (Appendix Q)
   - Backstage IDP (Appendix R)
   - Web Serving Internals (Appendix S)

### 🤔 Under Consideration
1. **eBPF integration** - Where and how much?
2. **I/O metrics** - Which tools and which phases?
3. **Pixie deployment** - Separate phase or sub-phase?

---

## How to Use This Directory

**For planning:**
1. Create new `.md` files for proposals
2. Keep them in this branch-specific planning directory
3. Review and refine before committing to main roadmap

**For discussion:**
1. Each file should have clear "Open Questions" section
2. Mark status (Drafted, Under Review, Approved, Rejected)
3. Link related files

**For approval:**
1. Once approved, integrate into main roadmap files
2. Archive planning doc or mark as "Implemented"
3. Update this README with outcome

---

## Next Actions

**Immediate:**
- [ ] Review eBPF strategy document
- [ ] Decide on Pixie integration approach
- [ ] Determine I/O metrics scope

**After Review:**
- [ ] Update `docs/roadmap-consolidation-analysis.md` with eBPF decisions
- [ ] Update phase files (3, 5, 7, 10) with I/O metrics
- [ ] Create new appendix if needed (Appendix T: eBPF Deep Dive?)

**Before Merging to Main:**
- [ ] Finalize all planning documents
- [ ] Update main `docs/roadmap.md`
- [ ] Create migration checklist
- [ ] Archive this planning directory or mark complete

---

## Branch Status

**Current branch:** `claude/review-project-roadmap-psMLb`

**Commits so far:**
1. Initial strategic review and consolidation analysis
2. 10-phase structure with FinOps integration
3. Visual summaries and diagrams
4. Advanced metrics & eBPF strategy (this planning doc)

**Ready to merge?** No - still in planning phase

---

## Contact / Notes

This is a working directory for the roadmap consolidation effort. Documents here represent proposals and may change significantly before being committed to the final roadmap.

**Philosophy:** Plan thoroughly, commit confidently.
