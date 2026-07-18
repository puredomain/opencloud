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
| 4 | Honest thumbnail 404 (thread the grpc "unsupported file type" detail through webdav's StatusNotFound branches) | The other half of [opencloud#3119](https://github.com/opencloud-eu/opencloud/issues/3119) — the thumbnails grpc service (`services/thumbnails/pkg/service/grpc/v0/service.go`) already returns `merrors.NotFound(..., "Unsupported file type")` for files whose mimetype `thumbnail.IsMimeTypeSupported` rejects, but all four `StatusNotFound` branches in the webdav service (`SpacesThumbnail`, `Thumbnail`, `PublicThumbnail`, `PublicThumbnailHead`) discarded that detail and always rendered the generic "File with name … could not be located", even though the sibling branches for `StatusTooEarly`/`StatusBadRequest`/`StatusForbidden` already pass `e.Detail` straight through. Adds `thumbnail.UnsupportedFileTypeDetail` (exported constant, shared by both services so the string isn't duplicated) and a `thumbnailNotFoundMsg` helper in webdav that renders the honest detail when it matches, falling back to the generic message for genuinely-missing files — no new status codes, single shared helper over the 4 call sites | Not yet submitted — small, self-contained cleanup on top of #3119 | `services/thumbnails/pkg/thumbnail/thumbnail.go`, `services/thumbnails/pkg/service/grpc/v0/service.go`, `services/webdav/pkg/service/v0/service.go`, `services/webdav/pkg/service/v0/service_test.go` |
| 5 | Media duration extraction for video, exposed as `oc:audio-duration` | Tika's `xmpDM:duration` extraction (`tika.go`) was gated to `Content-Type` prefix `audio/` only, so video files (probed: corpus `video_01.mp4`, `video/mp4`, Tika reports `xmpDM:duration=2.0`) never got a duration read even though Tika emits the same `xmpDM` key for video. Widens the gate to also match `video/*` and reuses `getAudio` as-is — the search proto's `Entity` has no separate `Video` message, only `Audio` (field 15), and `getAudio`'s field reads key off `xmpDM:*` meta keys rather than content-type, so for a video file the audio-only fields (album, artist, …) simply stay nil while `Duration` populates normally; a thin `getVideo` wrapper would just re-read the same key into the same struct with no behavioural difference. Duration is **milliseconds** (`int64`) — Tika's raw `xmpDM:duration` is seconds as a float string; `getAudio` already does `math.Round(f * 1000)` (see `tika_test.go`'s pre-existing "225.5"→`225500` case, now joined by a video case: "2.0"→`2000`). Patch #3's `appendPhotoProps` (webdav) is extended to emit `oc:audio-duration` from `entity.GetAudio().GetDuration()` when present — emit-when-present, same pattern as the rest of that function; no proto change, `Audio.duration` (field 9) already existed | Not yet submitted — small, self-contained extension of #3's exposure pattern to an existing-but-ungated extractor path | `services/search/pkg/content/tika.go`, `services/search/pkg/content/tika_test.go`, `services/search/pkg/bleve/bleve.go` (read-side mime gate — found live on tc4: index carried the field, `getAudioValue` dropped it at query time for `video/*`), `services/search/pkg/bleve/backend_test.go`, `services/webdav/pkg/service/v0/search.go`, `services/webdav/pkg/service/v0/search_test.go` |

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
