# systemd sandbox for xmtpd
FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive
ENV XMTDP_VERSION=1.0.0

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      systemd \
      ca-certificates \
      curl && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Fetch xmtpd release tarball and install the binary
RUN curl -L "https://github.com/xmtp/xmtpd/releases/download/v${XMTDP_VERSION}/xmtpd_${XMTDP_VERSION}_linux_amd64.tar.gz" \
      -o /tmp/xmtpd.tar.gz && \
    mkdir -p /usr/local/bin && \
    tar -xzf /tmp/xmtpd.tar.gz \
      -C /usr/local/bin \
      --strip-components=1 \
      "xmtpd_${XMTDP_VERSION}_linux_amd64/xmtpd" && \
    chmod 0755 /usr/local/bin/xmtpd && \
    rm -f /tmp/xmtpd.tar.gz

# Fetch the testnet config file
RUN mkdir -p /etc/xmtpd && \
    curl -L "https://github.com/xmtp/smart-contracts/releases/download/v2025.12.03-1/testnet.json" \
      -o /etc/xmtpd/testnet.json

COPY systemd/ /etc/systemd/system/

# Run systemd as PID 1
CMD ["/lib/systemd/systemd"]