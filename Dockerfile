FROM scratch

LABEL org.opencontainers.image.title=CatSync
LABEL org.opencontainers.image.description="Sync the「cat」config."
LABEL org.opencontainers.image.source=https://github.com/Sn0wo2/CatSync
LABEL org.opencontainers.image.url=https://github.com/Sn0wo2/CatSync
LABEL org.opencontainers.image.vendor=Sn0wo2
LABEL org.opencontainers.image.authors=Sn0wo2
LABEL org.opencontainers.image.licenses=MIT

COPY CatSync /CatSync

ENTRYPOINT ["/CatSync"]