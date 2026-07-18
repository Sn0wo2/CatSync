FROM alpine:latest AS certs
RUN apk add --no-cache ca-certificates

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

LABEL org.opencontainers.image.title=CatSync
LABEL org.opencontainers.image.description="Sync the「cat」config."
LABEL org.opencontainers.image.source=https://github.com/Sn0wo2/CatSync
LABEL org.opencontainers.image.url=https://github.com/Sn0wo2/CatSync
LABEL org.opencontainers.image.vendor=Sn0wo2
LABEL org.opencontainers.image.authors=Sn0wo2
LABEL org.opencontainers.image.licenses=MIT
LABEL org.opencontainers.image.license.name="MIT License"
LABEL org.opencontainers.image.license.spdx=MIT

COPY CatSync /CatSync
WORKDIR /app
ENTRYPOINT ["/CatSync"]
