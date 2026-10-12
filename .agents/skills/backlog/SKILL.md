---
name: backlog
description: Maintain a selected issue backlog with categorized IDs and synchronized plan, progress and completion comments.
---

# Backlog Maintenance

- Resolve the tracker and management comments from the request or project docs; do not assume issue numbers, comment IDs or write permission.
- Categories: **B** bug fixes, **O** optimization, **F** features, **Q** quality/tooling, **I** small improvements, **D** documentation.
  - Classify by outcome: new CLI capability is F, CI validation is Q, a bug fix remains B even if faster.
  - Add project-specific categories with unique prefixes and defined scope. Distinguish reproduced bugs and measured gains from hypotheses.
- IDs: number independently per category from 001; preserve existing formats. Check open items, archives and reservations; never reuse allocated or cancelled IDs.
  - Keep IDs stable; allocate new ones for separate follow-ups and retain incomplete parents. Use IDs only in the tracker, not code, commits or repository docs.
- Keep surfaces distinct and synchronized whenever scope, priority or execution changes:
  - **Body:** Rules, incomplete categorized items, acceptance criteria and management links.
  - **Plan/priority:** Order, dependencies and deferrals; priority tables cover every open ID exactly once, with updated counts.
  - **Progress:** Current work, actual validation state, pending decisions and blockers; no conversation log.
  - **Archive(s):** Item ID, verified integration commit link and one outcome sentence.
- Substantial features need design records and user docs. Priority changes or pauses do not change acceptance criteria; accepted designs and local tests do not establish completion.
- Complete only after acceptance gates pass and integration is verified on the designated remote branch:
  - Append and verify the archive entry without duplicates, then remove the open item and synchronize plan, progress and dependencies.
  - Retain unfinished parent scope; pushed WIP is not completion. Retire obsolete IDs with a reason, not an implementation claim.
- Fetch before edits, preserve concurrent changes, then read back affected surfaces and check IDs, counts, links and status together.
  - Recover comment locations from live links; after partial failure, re-fetch and retry only missing edits. Keep local checks, CI, integration and deployment distinct.
- Prefer a dedicated sidekick using the least expensive available model for backlog retrieval and updates; request concise summaries and focused diffs to conserve the primary agent's context.
- The primary agent must supervise edits, review their correctness and scope, and verify the saved results rather than trusting the sidekick's report alone.
