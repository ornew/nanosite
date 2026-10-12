# Adopting the Template

This scaffold supplies documentation and site tooling; choose implementation
language, packaging, license and release policies for each project.

## Project identity

Replace `PROJECT_NAME`, `ONE_SENTENCE_TAGLINE`, installation commands and example
prompts in the root README. Keep useful structure and remove unused sections.
Add only badges and claims whose targets or evidence exist. Choose and add a
project-specific license file before describing its license.

Update `site/config.json` with the same name and description. Set `repository`
and `branch` for source links, and set `siteURL` only after choosing the site's
actual origin. Documentation sections are configurable; add or remove sections
according to the project's readers.

## Working agreement

Keep `AGENTS.md` small. Fill in actual source locations, toolchain and validation
commands. Retain task-specific guidance in `.agents/skills/` only when useful.
The template does not inherit another project's issue numbers, push permissions,
model selection, temporary workflow pauses or worktree locations.

## First documents

Replace the specification introduction with the project's contracts and support
status. Copy authoring templates to finished pages, remove their prompts and add
links to section introductions. Design records preserve decisions; optimization
records preserve scoped measurement evidence. Published documents must not rely
on development conversations or temporary backlog labels.

## Verification and deployment

Run the actual project checks once implementation exists. Build the documentation
site and inspect it before configuring deployment. Keep generated output out of
Git. Treat source links, publication, Netlify setup and release automation as
separate configuration choices.
