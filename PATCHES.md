# PATCHES.md — timocloud fork of OpenCloud

## Purpose of this fork

This is timocloud's fork of [opencloud-eu/opencloud](https://github.com/opencloud-eu/opencloud),
maintained under the owner's GitHub namespace per the fork-friendly backend policy
(design doc `docs/specs/2026-07-07-timocloud-design.md`, decision D5, §10).

We deploy OpenCloud (posixfs storage driver) as the sole stateful brain of timocloud:
users, spaces, permissions, versioning, trash, search index, thumbnails, and WOPI
sessions. We patch this fork only where necessary — to fix a bug blocking us, to add
a capability upstream doesn't have yet, or to unblock a timocloud-specific
integration point (e.g. WOPI wiring for OnlyOffice, SSE contract needs). Everything
upstreamable should be upstreamed; the goal is to keep our patch surface as close to
zero as practically possible.

## Branch layout

- `upstream-main` — tracking branch, kept in sync with upstream `main` (or the
  upstream release branch/tag we're pinned to). No patches live here.
- `timocloud` — our working branch. Carries only our patches, rebased on top of
  the upstream release we're currently pinned to.

## Pinned upstream version

- Base tag: **`v7.2.0`** — repinned 2026-07-12 to MATCH THE LIVE DEPLOYMENT
  (opencloudeu/opencloud-rolling:7.2.0). The original `timocloud` branch was
  cut from `v7.2.1`, one minor AHEAD of production; the first fork-image
  boot against the production data volume failed owner authentication
  ("could not authenticate ... code 6" from the proxy's basic
  authenticator; graph /users listed only admin) while the on-disk
  idm.boltdb still contained the owner entry — a 7.2.0→7.2.1 IDM
  read-compat issue, not data loss (verified: downgrading to upstream
  7.2.0 restored owner auth untouched). RULE: the fork base tag must equal
  the deployed version; bump both together, testing the IDM upgrade on a
  volume copy first. Active branch: `timocloud-7.2.0`.

## Patch table

| # | Patch | Purpose | Upstream status | Files touched |
|---|-------|---------|------------------|----------------|
| 1 | Expose reva's `custom_mimetypes_json` storageprovider hook via `STORAGE_USERS_CUSTOM_MIMETYPES_JSON` | reva's `mime.RegisterMime` mechanism exists (config key `custom_mimetypes_json`, reads a JSON ext→mime file) but opencloud never wires it, so OOXML Visio (`.vsdx` family) resolves to `application/octet-stream` and never reaches the app registry — root cause + repro in timocloud `docs/validation/upstream-opencloud-vsdx-mimetype.md` | Filed as [opencloud#3120](https://github.com/opencloud-eu/opencloud/issues/3120) (2026-07-12); this patch is the config-driven fix the issue requests — submit as PR candidate | `services/storage-users/pkg/config/config.go`, `services/storage-users/pkg/revaconfig/config.go` |
| 2 | Video poster-frame thumbnails via ffmpeg | Thumbnails service 404s all `video/*` (no supported mimetype, no decoder) — timocloud Phase-0 P4 finding. Adds `VideoDecoder` (spool→`ffmpeg -ss 1 … -frames:v 1`→PNG→existing resize pipeline), five video mimetypes in both build-variant tables, `ffmpeg` in the runtime image | Filed as [opencloud#3119](https://github.com/opencloud-eu/opencloud/issues/3119) (2026-07-12); the ffmpeg dependency may be upstream-contentious — offer as opt-in | `services/thumbnails/pkg/preprocessor/preprocessor_video.go` (new), `services/thumbnails/pkg/preprocessor/preprocessor.go`, `services/thumbnails/pkg/thumbnail/mimetypes.go`, `services/thumbnails/pkg/thumbnail/mimetypes_vips.go`, `Dockerfile` |
| 3 | Emit extracted photo/image/location facets on search REPORT responses | The search engine already extracts EXIF via Tika (`photo`, `image`, `location` — takenDateTime, camera, dimensions, GPS) and stores/returns them through bleve and the search gRPC `Entity`, but `matchToPropResponse` in the webdav service never emitted them, so no WebDAV client could read e.g. a photo's taken-date. Adds `appendPhotoProps`: `oc:photo-*`, `oc:image-{width,height}`, `oc:location-{latitude,longitude,altitude}` props, emit-when-present only | Not yet submitted — clean upstream-PR candidate (pure exposure of already-indexed data, no new deps) | `services/webdav/pkg/service/v0/search.go`, `services/webdav/pkg/service/v0/search_test.go` |

(Update this table as patches land. Each row should point to the commit(s) or PR
that introduced the patch on `timocloud`, and track whether it's been submitted
upstream and its disposition — merged / rejected / not yet submitted.)

## Rebase policy

Per design doc §10:

- Rebase `timocloud` onto new upstream releases as they ship; don't let rebase
  distance grow unbounded.
- Before rebasing: review the patch table above, confirm each patch is still
  needed (check whether upstream has meanwhile absorbed it — if so, drop our
  patch and mark it "merged upstream" in the table rather than carrying a
  redundant diff).
- After rebasing: re-run our verification suite (see design doc §8) against the
  Docker Compose stack before promoting the new `timocloud` HEAD.
- Prefer small, self-contained patches (one concern per commit) so each can be
  dropped or upstreamed independently without entangling the rest of the diff.
- When a patch is accepted upstream, drop it from `timocloud` on the next
  rebase and update its row to "merged upstream — removed from fork".
