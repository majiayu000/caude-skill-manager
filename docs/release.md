# Release Readiness

## Current status

As of 2026-06-17, `majiayu000/caude-skill-manager` has Go module tags up to
`v0.2.0`, but no published GitHub release assets. The repository already has a
GoReleaser config and a tag-triggered GitHub Actions release workflow, but the
install instructions should treat binary archives as unavailable until an
asset-backed release is published.

## Release path

1. Confirm CI is green on `main`.
2. Choose the next SemVer tag. `v0.3.0` is the expected next tag for the
   manifest/shard registry compatibility release.
3. Review `CHANGELOG.md` and move the intended notes out of `Unreleased`.
4. Create and push the tag:

   ```bash
   git tag v0.3.0
   git push origin v0.3.0
   ```

5. Wait for the `Release` workflow to finish.
6. Verify the GitHub release and downloadable assets:

   ```bash
   gh release view v0.3.0
   gh release download v0.3.0 --pattern 'sk_darwin_arm64.tar.gz' --dir /tmp/sk-release-check
   ```

7. Install the downloaded archive on at least one supported platform and run:

   ```bash
   sk --help
   sk doctor
   HOME=/tmp/sk-home-release-check sk search testing
   ```

## Release blockers before advertising binary installs

- An asset-backed GitHub release must exist.
- Release assets must include the archive names used by the README.
- `go install github.com/majiayu000/caude-skill-manager@latest` must resolve to
  the new tag after it is visible to the Go proxy.
- The GoReleaser Homebrew upload path depends on `HOMEBREW_TAP_GITHUB_TOKEN`;
  if that token is unavailable, Homebrew publication should be documented as
  skipped or handled separately.
- `sk update` should remain documented as planned until the command performs an
  actual update.
