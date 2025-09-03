FROM golang:1.22 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X github.com/onurbalmeida/k8s-cleanup/cmd.Version=${VERSION:-dev} -X github.com/onurbalmeida/k8s-cleanup/cmd.Commit=${COMMIT:-none} -X github.com/onurbalmeida/k8s-cleanup/cmd.Date=${DATE:-unknown} -X github.com/onurbalmeida/k8s-cleanup/cmd.BuiltBy=docker" -o /out/k8s-cleanup .

FROM gcr.io/distroless/static
COPY --from=build /out/k8s-cleanup /k8s-cleanup
ENTRYPOINT ["/k8s-cleanup"]
