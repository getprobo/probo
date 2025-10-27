FROM ubuntu:24.04

LABEL org.opencontainers.image.source="https://github.com/getprobo/probo"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.vendor="Probo Inc"

RUN useradd -m probo && \
    apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y ca-certificates libcap2-bin && \
    rm -rf /var/lib/apt/lists/*

COPY probod /usr/local/bin/probod
COPY entrypoint.sh /usr/local/bin/entrypoint.sh

RUN chmod +x /usr/local/bin/probod && \
    chmod +x /usr/local/bin/entrypoint.sh && \
    setcap CAP_NET_BIND_SERVICE=+eip /usr/local/bin/probod && \
    mkdir -p /etc/probod && \
    chown probo:probo /etc/probod

USER probo

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
