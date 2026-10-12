# Maintain the Nanosite repository template

Nanosite's documentation conventions and site tooling originated in an OSS
project template. Maintain the template as a reviewed set of reusable files;
keep Nanosite's own generator implementation and site behavior authoritative.

## Preserve project ownership

The root Go module owns the `nanosite` command and its `render` and `publish`
packages. The repository's `site/config.json` and `site/build.sh` configure and
build this checkout's site from the root module. Preserve this source ownership
when preparing files for reuse; consumer instructions should pin a specific
Nanosite revision rather than depend on a neighboring checkout.

Keep Nanosite-specific contracts and links in its README, specification, guides,
and development docs. Reusable template files should avoid inventing a consumer's
license, package name, support policy, installation command, or product claims.
Use clearly marked authoring prompts only where the adopter must supply a real
project-specific value.

## Review template changes

When adapting files from a template or another repository, apply and review the
intended patch against the target file. Preserve target-specific configuration,
design, and documentation. Do not import unrelated Git history or overwrite
project contracts and working rules as part of copying a template.

Validate the source checkout using the checks in [development](README.md) and
rebuild its site using [site maintenance](site.md). Keep generated output out
of source control. Follow the [commit-message conventions](commit-message.md). Review publication
and deployment separately from local builds.
