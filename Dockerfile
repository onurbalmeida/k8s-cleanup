FROM golang:1.25 AS build
WORKDIR /src
ENV CGO_ENABLED=0 GO111MODULE=on GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w -X github.com/onurbalmeida/k8s-cleanup/cmd.Version=${VERSION} -X github.com/onurbalmeida/k8s-cleanup/cmd.Commit=${COMMIT} -X github.com/onurbalmeida/k8s-cleanup/cmd.Date=${DATE} -X github.com/onurbalmeida/k8s-cleanup/cmd.BuiltBy=docker" -o /out/k8s-cleanup .

FROM gcr.io/distroless/static
COPY --from=build /out/k8s-cleanup /k8s-cleanup
ENTRYPOINT ["/k8s-cleanup"]
