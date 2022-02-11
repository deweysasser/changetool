ARG GO_VERSION=1.17
FROM golang:${GO_VERSION} as builder
ARG PROGRAM=nothing
ARG VERSION=development

RUN mkdir /src /output

WORKDIR /src

COPY . .
RUN GOBIN=/output make install VERSION=$VERSION
RUN PROGRAM=$(ls /output); echo "#!/bin/sh\nexec '/usr/bin/$PROGRAM' \"\$@\"" > /docker-entrypoint.sh && chmod +x /docker-entrypoint.sh


FROM alpine:latest
RUN apk add --no-cache libc6-compat ca-certificates

COPY --from=builder /output/* /usr/bin
COPY --from=builder /docker-entrypoint.sh  /docker-entrypoint.sh
ENTRYPOINT ["/docker-entrypoint.sh"]
