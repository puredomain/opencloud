# THIRD-PARTY.md — license tracking for the timocloud OpenCloud fork

Per design doc `docs/specs/2026-07-07-timocloud-design.md` §10 (license hygiene).
This is a personal/family deployment, which keeps AGPL network-use obligations
trivial in practice, but we track them anyway for hygiene and in case deployment
scope ever changes.

## This repository

- **OpenCloud** (this fork of [opencloud-eu/opencloud](https://github.com/opencloud-eu/opencloud))
  — mixed licensing: predominantly Apache-2.0, with some AGPL-licensed components
  within the OpenCloud project. Confirm exact per-component licensing against
  upstream's own `LICENSE` / `NOTICE` files before relying on this summary; this
  stub will be filled in with a precise component-by-component breakdown as we
  patch or vendor anything.
  - Upstream base: `v7.2.1`.
  - Our patches: see `PATCHES.md` (currently none).

## Related third-party components used elsewhere in the timocloud stack

(Tracked here for visibility even though they live in other repos/containers —
see the system architecture in the design doc §3.)

- **OnlyOffice Docs CE** — AGPL. Runs as a separate container behind OpenCloud's
  WOPI service; not vendored into this fork.
- **ownCloud desktop client** — Apache-2.0 / GPL mixed parts. Referenced/tested
  for Windows VFS on-demand sync (design doc §6, Phase B); not vendored here.
- **kDrive desktop client** — GPLv3. Reference only for the macOS File Provider
  agent (design doc §6, Phase D); not vendored here, and only relevant if a
  future timocloud agent adopts GPL-compatible licensing.

## Maintenance note

Update this file whenever:
- A patch to this fork adds or touches code under a different license than the
  surrounding file.
- A new third-party library/service is vendored into this fork (not just run
  alongside it in Docker Compose).
- The deployment scope changes from personal/family use (which would warrant a
  fresh look at AGPL source-availability obligations).
