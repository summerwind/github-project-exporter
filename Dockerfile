FROM golang:1.13 AS build

ENV GO111MODULE=on

WORKDIR /go/src/github.com/summerwind/github-project-exporter
COPY go.mod go.sum ./
RUN go mod download

COPY . /workspace
WORKDIR /workspace

ARG VERSION
ARG COMMIT

RUN go vet ./...
RUN go test -v ./...
RUN CGO_ENABLED=0 go build -ldflags "-X main.VERSION=${VERSION} -X main.COMMIT=${COMMIT}" .

###################

FROM quay.io/prometheus/busybox:latest

COPY --from=build /workspace/github-project-exporter /bin/github-project-exporter

EXPOSE 9410
ENTRYPOINT ["/bin/github-project-exporter"]
