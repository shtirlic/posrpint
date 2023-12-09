FROM golang:1.21-bookworm

ENV GOOS=linux
ENV GOARM=6
ENV GOARCH=arm
ENV CGO_ENABLED=1
ENV CC=arm-linux-gnueabihf-gcc 
ENV PATH="/go/bin/${GOOS}_${GOARCH}:${PATH}:/opt/cross-pi-gcc/bin/"
ENV PKG_CONFIG_PATH=/usr/lib/arm-linux-gnueabihf/pkgconfig
ENV CGO_LDFLAGS="-Wl,-rpath-link /lib/arm-linux-gnueabihf"

# install build & runtime dependencies
RUN dpkg --add-architecture armhf \
    && apt update && apt upgrade -y \
    && apt install -y --no-install-recommends \
        pkg-config \
        libusb-1.0-0-dev:armhf \
    && rm -rf /var/lib/apt/lists/* 
RUN wget -q https://github.com/Pro/raspi-toolchain/releases/latest/download/raspi-toolchain.tar.gz \
    && tar xfz raspi-toolchain.tar.gz --strip-components=1 -C /opt

