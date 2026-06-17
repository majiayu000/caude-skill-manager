# Changelog

All notable changes to `sk` should be recorded here.

This project has Go module tags, but does not yet have published GitHub release
assets. Until the first asset-backed release is cut, use the `Unreleased`
section as the source of truth for release notes.

## Unreleased

### Added

- Documented release status, initial release checklist, and current limitations.
- Linked the README to release notes and release-readiness documentation.
- Added a registry consumer readiness spec for manifest/shard compatibility,
  release verification, and follow-up issue tracking.

### Known limitations

- No binary GitHub release is published yet.
- `sk update` is a placeholder command and does not update installed skills.
- Registry-backed commands depend on the configured registry URL and network
  access.
- Private GitHub repositories, enterprise GitHub hosts, authenticated downloads,
  signature verification, and sandboxing are not documented as supported.

## 0.0.0 - Not released

- Initial public release has not been cut.
