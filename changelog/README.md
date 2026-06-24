# Changelog

This `changelog/` directory is the single source of truth for the project's
in-repo changelog. Per-major-version files live here, one Markdown file per major
series:

- `CHANGELOG-v0.x.md` — the `0.x` series (active)
- `CHANGELOG-v1.x.md` — created when the `1.x` series begins
- …

**How updates work:**

- The active changelog file (e.g. `CHANGELOG-v0.x.md`) is maintained automatically
  by [release-please](https://github.com/googleapis/release-please) in its release
  PR. [`release-please-config.json`](../release-please-config.json) sets
  `"changelog-path": "changelog/CHANGELOG-v0.x.md"`. When a new major series
  begins, create a new `CHANGELOG-vN.x.md` and bump `changelog-path`.
- GitHub Release notes (shown on the Releases page for each tag) are rendered
  separately by [GoReleaser](https://goreleaser.com) from the commit history in
  the release workflow. That is distinct from this in-repo changelog.

**Do not hand-edit:** These files are auto-generated and excluded from `oxfmt`
(see [`.oxfmtignore`](../.oxfmtignore)). Reformatting them causes spurious diffs
on every PR between releases.
