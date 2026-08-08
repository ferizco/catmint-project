# syntax=docker/dockerfile:1

ARG GO_VERSION=1.25.12

FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags "-s -w -X catmint/cmd.version=${VERSION}" -o /out/catmint .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /work

COPY --from=build /out/catmint /usr/local/bin/catmint

ENTRYPOINT ["catmint"]
CMD ["--help"]
