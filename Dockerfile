# Please use this Dockerfile only if
# you want to build an image from source without
# pnpm and Go installed on your dev machine.

# You can build OpenCloud using this Dockerfile
# by running following command:
# `docker build -t opencloud/opencloud:custom .`

# In most other cases you might want to run the
# following command instead:
# `make -C opencloud dev-docker`
# It will build a `opencloud/opencloud:dev` image for you
# and use your local pnpm and Go caches and therefore
# is a lot faster than the build steps below.


FROM quay.io/opencloudeu/nodejs-ci:24 AS generate

COPY ./ /opencloud/

# timocloud build fix (found live 2026-07-12): running node-generate-prod
# from the opencloud/ submodule skips every OTHER module's asset build —
# services/idp keeps only its tracked assets/.keep, //go:embed embeds that
# placeholder, and the server crash-loops at runtime with 'Could not open
# index template'. The ROOT Makefile's node-generate-prod loops OC_MODULES
# (idp included); run it there. Upstream-issue candidate: the published
# images use docker/Dockerfile.multiarch + a root-level generate in CI, so
# this source Dockerfile appears untested upstream.
WORKDIR /opencloud
RUN make node-generate-prod

FROM quay.io/opencloudeu/golang-ci:1.25 AS build

COPY --from=generate /opencloud /opencloud

WORKDIR /opencloud/opencloud
RUN make go-generate build ENABLE_VIPS=true

FROM alpine:3.24

# timocloud patch #2: ffmpeg for video poster-frame thumbnails.
# timocloud deploy parity: the published opencloud-rolling image creates
# uid/gid 1000 with HOME=/var/lib/opencloud — without it, HOME='/' and the
# server looks for config at /.opencloud/config and crash-loops on
# 'jwt_secret has not been set' (found live 2026-07-12 on first fork-image
# boot). Mirror that user setup here.
RUN apk add --no-cache attr ca-certificates curl mailcap tree vips ffmpeg && \
	echo 'hosts: files dns' >| /etc/nsswitch.conf && \
	addgroup -g 1000 opencloud-group && \
	adduser -D -H -u 1000 -G opencloud-group -h /var/lib/opencloud -s /sbin/nologin opencloud-user && \
	mkdir -p /var/lib/opencloud /etc/opencloud && \
	chown -R opencloud-user:opencloud-group /var/lib/opencloud /etc/opencloud

LABEL maintainer="OpenCloud GmbH <devops@opencloud.eu>" \
        org.opencontainers.image.title="OpenCloud" \
        org.opencontainers.image.vendor="OpenCloud GmbH" \
        org.opencontainers.image.authors="OpenCloud GmbH" \
        org.opencontainers.image.description="OpenCloud is a modern file-sync and share platform" \
        org.opencontainers.image.licenses="Apache-2.0" \
        org.opencontainers.image.documentation="https://github.com/opencloud-eu/opencloud" \
        org.opencontainers.image.source="https://github.com/opencloud-eu/opencloud"

ENTRYPOINT ["/usr/bin/opencloud"]
CMD ["server"]

COPY --from=build /opencloud/opencloud/bin/opencloud /usr/bin/opencloud
