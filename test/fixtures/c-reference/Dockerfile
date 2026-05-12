# Dockerfile for the onvif_simple_server C reference build.
#
# Used by P0.1 (stand up the C reference server) to produce parity
# fixtures the Go server is diffed against. Pinned to debian:bookworm
# because the upstream Makefile links -lrt (Linux-only) and expects
# libtomcrypt / json-c / zlib system packages.
#
# Build context expected: the playground onvif_simple_server checkout.
# Example:
#   docker build \
#     -f test/fixtures/c-reference/Dockerfile \
#     -t onvif-c-reference:latest \
#     /Users/fawadmazhar/github/codes/oma/playground/onvif_simple_server
#
# Run (with canonical fixture config mounted):
#   docker run --rm -p 18080:8080 \
#     -v $(pwd)/test/fixtures/config/server.conf:/etc/onvif_simple_server.conf:ro \
#     onvif-c-reference:latest

FROM debian:bookworm-slim AS build

RUN apt-get update && apt-get install -y --no-install-recommends \
        build-essential \
        libtomcrypt-dev \
        libjson-c-dev \
        zlib1g-dev \
        ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY . /src

# Build all three C binaries (onvif_simple_server, onvif_notify_server,
# wsd_simple_server). Default Makefile links -ltomcrypt; that's fine
# inside this Linux container.
RUN make -j"$(nproc)"

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
        libtomcrypt1 \
        libjson-c5 \
        zlib1g \
        ca-certificates \
        curl \
    && rm -rf /var/lib/apt/lists/*

# Install the C binaries and template directories at the paths the
# binaries expect by default.
COPY --from=build /src/onvif_simple_server /usr/local/bin/onvif_simple_server
COPY --from=build /src/onvif_notify_server /usr/local/bin/onvif_notify_server
COPY --from=build /src/wsd_simple_server /usr/local/bin/wsd_simple_server

COPY --from=build /src/device_service_files   /etc/onvif_simple_server/device_service_files
COPY --from=build /src/media_service_files    /etc/onvif_simple_server/media_service_files
COPY --from=build /src/media2_service_files   /etc/onvif_simple_server/media2_service_files
COPY --from=build /src/ptz_service_files      /etc/onvif_simple_server/ptz_service_files
COPY --from=build /src/events_service_files   /etc/onvif_simple_server/events_service_files
COPY --from=build /src/deviceio_service_files /etc/onvif_simple_server/deviceio_service_files
COPY --from=build /src/generic_files          /etc/onvif_simple_server/generic_files
COPY --from=build /src/notify_files           /etc/onvif_simple_server/notify_files
COPY --from=build /src/wsd_files              /etc/onvif_simple_server/wsd_files

# CGI front-end. The C server expects to be invoked as a CGI by an HTTP
# daemon (e.g. busybox httpd). For fixture capture we use a minimal
# busybox httpd config that maps /onvif/<service>_service to the binary.
RUN apt-get update && apt-get install -y --no-install-recommends busybox \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /www/cgi-bin/onvif /www/onvif

# CGI wrappers: the C server picks the service name from argv[argc-1]
# when argv[0] does not contain "_service" (see onvif_simple_server.c:266-278),
# so we wrap each service in a tiny shell script that invokes the binary
# with `-c <conf> <service>`. busybox httpd auto-treats /cgi-bin/* as CGI.
# We also expose /onvif/<svc> via an httpd.conf URL alias so the ONVIF
# client sees the canonical /onvif/<svc> path.
RUN for svc in device_service media_service media2_service ptz_service \
               events_service deviceio_service ; do \
        printf '#!/bin/sh\ncd /etc/onvif_simple_server\nexec /usr/local/bin/onvif_simple_server -c /etc/onvif_simple_server.conf %s\n' "$svc" \
            > /www/cgi-bin/onvif/$svc ; \
        chmod +x /www/cgi-bin/onvif/$svc ; \
    done

# Symlink the template directory tree where the binary expects it relative
# to the conf directory.
RUN ln -s /etc/onvif_simple_server.conf /etc/onvif_simple_server/onvif_simple_server.conf || true

EXPOSE 8080

# busybox httpd serving CGI from /www. The C server is reached via
#   POST http://<host>:<port>/cgi-bin/onvif/<service>
# (URL is only used for fixture capture; the captured response body is
# what gets diffed, not the URL.)
CMD ["busybox", "httpd", "-f", "-p", "8080", "-h", "/www"]
