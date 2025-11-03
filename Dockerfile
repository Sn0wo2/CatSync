FROM alpine:latest

LABEL org.opencontainers.image.source=https://github.com/Sn0wo2/CatSync
LABEL org.opencontainers.image.description="Sync the「cat」config."
LABEL org.opencontainers.image.licenses=MIT

RUN mkdir -p /app
WORKDIR /app

COPY CatSync /app/CatSync

RUN chmod +x /app/CatSync

ENTRYPOINT ["/app/CatSync"]
