# build
FROM --platform=$BUILDPLATFORM golang:1.26.6-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -tags timetzdata -ldflags="-s -w" -o /out/core ./cmd/core

# runtime
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=build /out/core /app/core
ENTRYPOINT [ "/app/core" ]