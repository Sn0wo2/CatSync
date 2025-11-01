FROM alpine:latest

ARG TARGETPLATFORM

RUN mkdir -p /app
WORKDIR /app

COPY ${TARGETPLATFORM}/CatSync /app/CatSync

RUN chmod +x /app/CatSync

ENTRYPOINT ["/app/CatSync"]
