# syntax=docker/dockerfile:1
FROM golang:1.18.6 as builder

WORKDIR /build

ARG MA_VERSION

# Note: git clone is not required when building locally.
RUN git clone --depth=1 --branch $MA_VERSION --single-branch https://github.com/iogakos/agent.git

# Note: uncomment to build from local source.
# COPY . ./agent

WORKDIR /build/agent
ARG MA_PROTOCOL
RUN make "build-$MA_PROTOCOL"

FROM alpine:3.14

ARG MA_PROTOCOL
ARG MA_OS_ARCH

# copy agent binary and configuration
COPY --from=builder "/build/agent/metrikad-${MA_PROTOCOL}-linux-${MA_OS_ARCH}" "/opt/metrikad/metrikad-${MA_PROTOCOL}"
COPY --from=builder /build/agent/configs /etc/metrikad/configs

# create metrikad:metrikad user group
# and configure permissions for agent directories.
RUN addgroup metrikad
RUN adduser --system --ingroup metrikad --home /opt/metrikad --disabled-password metrikad
RUN chown -R metrikad:metrikad /etc/metrikad /opt/metrikad

# remove setuid and setgid permission from binaries
RUN find / -perm +6000 -type f -exec chmod a-s {} \; || true

WORKDIR /opt/metrikad
RUN ln -s metrikad-${MA_PROTOCOL} metrikad
CMD ./metrikad
