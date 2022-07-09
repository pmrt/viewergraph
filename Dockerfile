FROM golang:1.18-alpine AS build

WORKDIR /src

# deps
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# source code
COPY . .

# Build
# Don't use libc, the resulting binary will be statically linked against the
# libraries
ENV CGO_ENABLED=0
RUN go build -o /usr/local/bin/vgserver ./cmd/vgserver

ENTRYPOINT ["vgserver"]

