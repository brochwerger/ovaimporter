#FROM registry.access.redhat.com/ubi8/go-toolset:1.21 AS build
FROM golang:1.20-alpine AS build
WORKDIR /app/
COPY main.go /app
COPY templates /app/templates
# USER 1001
# RUN chown -R 1001:1001 /app
RUN go mod init github.com/brochwerger/ovaimporter 
RUN go mod tidy
RUN go build -o ovaimporter main.go

# FROM scratch
# COPY --from=build /app/ovaimporter /bin/ovaimporter
# ENTRYPOINT ["/bin/ovaimporter"]
ENTRYPOINT ["/app/ovaimporter"]
