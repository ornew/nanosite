# NNN. Decision title

- **Status**: Proposed
- **Author**: MAINTAINER
- **Date**: YYYY-MM-DD

Copy to an unused `NNN-topic.md` and remove authoring prompts. Use Accepted for
approved decisions, Implemented for integrated behavior, Superseded with a
successor link, or Rejected with a reason.

## Summary

State the problem and observable behavior. Label proposed APIs explicitly.

## Motivation

Describe the use case and current gap. Cite contracts, reproduced failures and
pinned measurements. Separate facts from hypotheses.

## Goals and scope

State outcomes, acceptance criteria and meaningful boundaries.

## Design

Define public inputs, outputs, defaults, errors and edge cases. Explain the
invariants preserving the contract; consider ownership, mutation, concurrency,
cancellation and resource accounting where relevant.

## Compatibility and support

Identify callers, serialized formats, generated consumers and backend differences.
Describe migrations and documents that must change with the implementation.

## Alternatives

Explain benefits, costs and reasons for choosing or rejecting credible options.
Keep unresolved maintainer decisions explicit.

## Validation and implementation

Describe regressions, conformance checks, dependencies and reviewable stages.
Distinguish completed checks from plans.

## Results and open questions

Include reproducible performance evidence when relevant. State limitations and
remaining gates; acceptance does not establish implementation or release.
