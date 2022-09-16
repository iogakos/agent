# syntax=docker/dockerfile:1
FROM golang:1.18.6 as builder

WORKDIR /build

ARG MA_VERSION

# Note: git clone is not required when building locally.
RUN git clone https://github.com/Metrika-Inc/agent.git --branch $MA_VERSION --single-branch

# Note: uncomment when building locally.
# COPY . ./agent

WORKDIR /build/agent
ARG MA_PROTOCOL
RUN make "build-$MA_PROTOCOL"

FROM alpine:3.14

ARG MA_PROTOCOL
ARG MA_OS_ARCH
COPY --from=builder "/build/agent/metrikad-${MA_PROTOCOL}-linux-${MA_OS_ARCH}" "/opt/metrikad/metrikad-${MA_PROTOCOL}" 
COPY --from=builder /build/agent/configs /etc/metrikad/configs

RUN addgroup metrikad
RUN adduser --system --ingroup metrikad --home /opt/metrikad --disabled-password metrikad
RUN chown -R metrikad:metrikad /etc/metrikad /opt/metrikad

EXPOSE 9000/tcp

# remove setuid and setgid permission from binaries
RUN find / -perm +6000 -type f -exec chmod a-s {} \; || true

WORKDIR /opt/metrikad
RUN ln -s metrikad-${MA_PROTOCOL} metrikad
CMD [ "./metrikad" ]
