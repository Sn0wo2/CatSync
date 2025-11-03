FROM alpine:latest

LABEL org.opencontainers.image.description="Sync the「cat」config."

RUN mkdir -p /app
WORKDIR /app

COPY CatSync /app/CatSync

RUN chmod +x /app/CatSync

ENTRYPOINT ["/app/CatSync"]
