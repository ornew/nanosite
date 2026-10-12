---
name: performance
description: Evaluate optimizations and benchmark claims with reproducible baseline and candidate measurements, applicability and correctness checks.
---

# Performance Evidence

Use `docs/optimizations/000-template.md`. Identify the operation, workloads and
execution profiles. A mechanism in one backend is not automatically in another.
Establish correctness before interpreting timing gains.

Serialize runs on the measurement host. Use paired or alternating baseline and
candidate samples; pin revisions, toolchains, inputs, configuration, units and
hardware. Measure affected operations separately and check normal paths.

Keep focused raw logs outside Git unless the project deliberately versions a
result fixture. Record locations, commands and provenance. Report time,
allocations and live memory as relevant, including noise and unmeasured paths.
Full snapshots describe combined performance; focused evidence supports attribution.

Update the optimization catalog with applicability and evidence. Preserve
rejected alternatives and measured tradeoffs. Separate measurement plans,
local results and production observations; avoid incomparable speed claims.
