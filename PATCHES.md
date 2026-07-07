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

- Base tag: **`v7.2.1`** (OpenCloud release, published 2026-07-06)
- Both `upstream-main` and `timocloud` were branched from this tag.

## Patch table

No patches yet — this fork currently carries zero diffs from `v7.2.1`.

| # | Patch | Purpose | Upstream status | Files touched |
|---|-------|---------|------------------|----------------|
| — | — | — | — | — |

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
