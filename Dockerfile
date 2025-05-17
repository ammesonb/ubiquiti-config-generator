FROM golang:1.24-alpine

WORKDIR /go/src/github.com/ammesonb/ubiquiti-config-generator
COPY go.mod ./
COPY go.sum ./

RUN go mod download

RUN apk add build-base

RUN go install golang.org/x/tools/cmd/goimports@latest
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.0
RUN go install github.com/fzipp/gocyclo/cmd/gocyclo@v0.6.0
RUN go install github.com/go-critic/go-critic/cmd/gocritic@v0.11.0
RUN go install golang.org/x/lint/golint@latest
