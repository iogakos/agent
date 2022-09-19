# syntax=docker/dockerfile:1
FROM golang:1.18.6 as builder

WORKDIR /build

ARG MA_VERSION

# Note: git clone is not required when building locally.
RUN git clone --depth=1 --branch ${MA_VERSION} --single-branch https://github.com/iogakos/agent.git

# Note: uncomment to build from local source.
# COPY . ./agent

WORKDIR /build/agent
ARG MA_PROTOCOL
RUN make build-${MA_PROTOCOL}

FROM bitnami/minideb:bullseye

# copy agent binary and configuration
ARG MA_PROTOCOL
ARG MA_OS_ARCH
COPY --from=builder "/build/agent/metrikad-${MA_PROTOCOL}-linux-${MA_OS_ARCH}" "/opt/metrikad/metrikad-${MA_PROTOCOL}"
COPY --from=builder /build/agent/configs /etc/metrikad/configs

# remove setuid and setgid permission from binaries
RUN find / -perm /6000 -type f -exec chmod a-s {} \; 2>/dev/null || true

WORKDIR /opt/metrikad
RUN ln -s metrikad-${MA_PROTOCOL} metrikad
CMD ./metrikad
