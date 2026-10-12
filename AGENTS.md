# AGENTS.md

Nanosite is the Go documentation site generator. Its root module is
`github.com/ornew/nanosite`; `site/` configures its own documentation site.

User instructions for the current task take precedence over these defaults.

## Documentation map

- `README.md`: product overview, install, quick start and documentation index.
- `spec/`: normative public contracts; `docs/design/`: decisions and rationale.
- `docs/tutorial/`, `docs/guide/`, `docs/cookbook/`: learning, features, recipes.
- `docs/optimizations/`: mechanisms, applicability and reproducible measurements.
- `docs/dev/`: repository operations; `site/`: Nanosite configuration and wrapper.

Keep detailed contracts, status and evidence in docs. Use skills when relevant:

- `documentation`: authoring docs and design/spec changes.
- `performance`: optimization and benchmark evidence.
- `backlog`: maintaining an explicitly selected backlog.
- `nanosite`: site publication and configuration.

## Working conventions

- Write repository docs, comments and commit messages in English.
- Write commit, issue and PR titles/bodies in the owner's voice: describe the
  problem, rationale and result. Omit references to agents, AI assistance, the
  owner/user, their requests, delegation or work history, and agent signatures.
- Reader-facing docs must stand alone without chat history or backlog IDs.
- Update affected contracts, examples and guides with implementation changes;
  regenerate generated files rather than editing them by hand.
- Ask for necessary maintainer decisions, including unspecified public semantics,
  with alternatives, tradeoffs and a recommendation. Decide routine choices locally.
- After asynchronous questions, keep the turn active and continue independent
  work; wait for answers before dependent actions.
- For high-impact changes, prefer critical review by an agent with context
  independent of the worker; reuse existing reviewers. Share requirements, diffs
  and validation evidence, not the worker's reasoning or a desired verdict.
  Evaluate findings and resolve material issues before completion.
- Verify affected behavior; distinguish local checks, CI, deployment and live
  integration evidence. Claim only checks actually run.
- Follow [commit messages](docs/dev/commit-message.md); include a body explaining
  the problem and rationale. Omit coding-agent co-author trailers.
- Keep logical changes reviewable. Commit, push and external edits require user
  authorization; this file grants none.

## Checks

For site or Markdown changes, run `sh site/build.sh -check` and `sh site/build.sh`.

Run `gofmt` on changed Go files, then `go test ./...` and `go vet ./...`.
`main.go` implements the CLI; `sitegen/` owns standard config, Markdown and the
shared build pipeline, while `render/` and `publish/` provide supporting APIs.
`sitegen/layout.html`, `sitegen/static/` and the Nanosite skill are embedded.
The CLI and `sitegen.BuildWith` share standard config and publication behavior.
Keep exported skill/config guidance aligned with behavior.
