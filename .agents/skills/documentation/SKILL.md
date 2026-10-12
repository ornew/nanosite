---
name: documentation
description: Write repository documentation with distinct user guides, developer records and durable specifications.
---

# Documentation

- **User-facing:**
  - `docs/tutorial/`: First-time starting point; prerequisites, progressive steps, a useful result and next steps.
  - `docs/guide/`: Per-feature explanation; usage, options, behavior, limits and errors.
  - `docs/cookbook/`: Per-use-case recipe; complete runnable solution, verified output and variations.
  - `docs/reference/`: Concise API/type/command/option lookup; not a normative specification.
- **Developer-facing:**
  - `docs/design/`: Problem, decision, rationale, alternatives and validation; distinguish proposed, accepted and implemented.
  - `docs/optimizations/`: Cost, mechanism, applicability, reproducible measurements and tradeoffs.
  - `docs/dev/`: Repository structure, contributor setup, checks, generation and maintenance.
- **Audience-independent:** `spec/` defines durable normative contracts, requirements, errors and compatibility.
- Keep usage and contributor workflows separate. Specs contain contracts, not tutorials, progress or development history.
- Use available templates and remove prompts. Verify examples; label pseudocode and proposed APIs.
- Update contracts and user guidance with features. Accepted does not mean implemented.
- Keep finished docs independent of chats and backlog IDs; link to contracts, update indexes and check site links.
- Prefer a dedicated sidekick using the least expensive available model for documentation retrieval and updates; request concise summaries and focused diffs to conserve the primary agent's context.
- The primary agent must supervise edits, review their correctness and scope, and verify the saved results rather than trusting the sidekick's report alone.
