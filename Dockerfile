FROM alpine:latest

RUN mkdir -p /opt/CatSync
WORKDIR /opt/CatSync

COPY CatSync /opt/CatSync/CatSync

RUN chmod +x /opt/CatSync/CatSync

ENTRYPOINT ["/opt/CatSync/CatSync"]
