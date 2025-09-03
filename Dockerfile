FROM --platform=$BUILDPLATFORM golang:1.25 AS build
WORKDIR /src
ENV CGO_ENABLED=0 GO111MODULE=on GOTOOLCHAIN=auto GOPROXY=https://proxy.golang.org,direct
ARG TARGETOS TARGETARCH VERSION=dev COMMIT=none DATE=unknown
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -buildvcs=false -trimpath \
      -ldflags="-s -w -X github.com/onurbalmeida/k8s-cleanup/cmd.Version=${VERSION} -X github.com/onurbalmeida/k8s-cleanup/cmd.Commit=${COMMIT} -X github.com/onurbalmeida/k8s-cleanup/cmd.Date=${DATE} -X github.com/onurbalmeida/k8s-cleanup/cmd.BuiltBy=docker" \
      -o /out/k8s-cleanup .

FROM gcr.io/distroless/static
COPY --from=build /out/k8s-cleanup /k8s-cleanup
ENTRYPOINT ["/k8s-cleanup"]
