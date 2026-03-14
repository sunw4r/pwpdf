FROM golang:1.26.1-bookworm

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
    libgtk-3-dev \
    libwebkit2gtk-4.1-dev \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace
