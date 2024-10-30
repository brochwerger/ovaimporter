#FROM registry.access.redhat.com/ubi8/go-toolset:1.21 AS build
FROM golang:1.23-alpine AS build
WORKDIR /app/
COPY *.go /app
COPY templates /app/templates
# USER 1001
# RUN chown -R 1001:1001 /app
RUN go mod init github.com/brochwerger/ovaimporter 
RUN go mod tidy
RUN go build -o ovaimporter .

RUN apk add qemu-img

ENTRYPOINT ["/app/ovaimporter"]

# FROM scratch
# WORKDIR /tmp
# # WORKDIR /data
# COPY --from=build /app/ovaimporter /bin/ovaimporter
# COPY --from=build /app/templates /app/templates

# WORKDIR /app/
# ENTRYPOINT ["/bin/ovaimporter"]
