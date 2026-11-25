ARG AGENT_BROWSER_VERSION=0.43.0
ARG CLEARCOTE_RELEASE=v0.1.0-pre.25
ARG CLEARCOTE_VERSION=151.0.7922.108

FROM node:24-slim AS agent-browser
ARG AGENT_BROWSER_VERSION
RUN npm install -g --allow-scripts=agent-browser agent-browser@$AGENT_BROWSER_VERSION

FROM debian:stable-slim AS clearcote
ARG CLEARCOTE_RELEASE
ARG CLEARCOTE_VERSION

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates curl xz-utils && \
    rm -rf /var/lib/apt/lists/*

RUN URL="https://github.com/clearcotelabs/clearcote-browser/releases/download/${CLEARCOTE_RELEASE}/clearcote-${CLEARCOTE_VERSION}-linux-x64.tar.xz" && \
    FILE=$(basename "$URL") && \
    cd /tmp && \
    curl -fsSL -O "$URL" && \
    curl -fsSL -O "$URL.sha256" && \
    sha256sum -c "$FILE.sha256" && \
    mkdir -p /clearcote && \
    tar -xJf "$FILE" -C /clearcote && \
    rm -f "$FILE" "$FILE.sha256"

# Builder Image
FROM golang:1.24 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -v -o ./bin/ .

# Distribution Image
FROM debian:stable-slim

ENV HOME=/browserfull

ENV BROWSERFULL_ALLOWED_ORIGINS=*
ENV BROWSERFULL_BROWSER_EXECUTABLE_PATH=/opt/clearcote/chrome
ENV BROWSERFULL_DATA_DIR=/data

# Install system dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    tini \
    fonts-liberation \
    fonts-noto-color-emoji \
    libasound2 \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    libatspi2.0-0 \
    libcairo-gobject2 \
    libcairo2 \
    libcups2 \
    libdbus-1-3 \
    libdrm2 \
    libegl1 \
    libexpat1 \
    libffi8 \
    libfontconfig1 \
    libgbm1 \
    libgdk-pixbuf-2.0-0 \
    libgl1-mesa-dri \
    libglib2.0-0 \
    libgnutls30 \
    libgtk-3-0 \
    libnspr4 \
    libnss3 \
    libpango-1.0-0 \
    libpangocairo-1.0-0 \
    libpcre2-8-0 \
    libselinux1 \
    libudev1 \
    libx11-6 \
    libx11-xcb1 \
    libxau6 \
    libxcb1 \
    libxcomposite1 \
    libxcursor1 \
    libxdamage1 \
    libxext6 \
    libxfixes3 \
    libxi6 \
    libxkbcommon0 \
    libxrandr2 \
    libxrender1 \
    libxshmfence1 \
    libxss1 \
    libxtst6 \
    libz1 \
    zlib1g && \
    rm -rf /var/lib/apt/lists/*

# User and directory setup
RUN useradd browser --uid 1000 --home-dir "$HOME" && \
    mkdir -p "$HOME" "$BROWSERFULL_DATA_DIR" && \
    chown browser:browser "$HOME" "$BROWSERFULL_DATA_DIR"

# Install agent-browser
COPY --from=agent-browser --chown=browser:browser \
     /usr/local/bin/agent-browser /usr/local/bin/agent-browser

COPY --from=clearcote --chown=browser:browser /clearcote/ /opt/clearcote/

# Install browserfull
COPY --from=builder --chown=browser:browser /app/bin/browserfull /usr/local/bin/browserfull

USER browser
WORKDIR "$BROWSERFULL_DATA_DIR"

EXPOSE 8080

ENTRYPOINT ["tini", "--"]
CMD ["/usr/local/bin/browserfull"]
