FROM golang:1.20-alpine AS base
WORKDIR /gss
COPY . ./

FROM base AS ci
RUN apk add build-base
RUN go install mvdan.cc/gofumpt@latest
RUN go install github.com/segmentio/golines@latest
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
RUN go install github.com/mfridman/tparse@latest

FROM base AS build
ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOARCH=amd64
ENV GOOS=linux
RUN go build -o gss -ldflags="-s -w" gss.go

FROM scratch AS prod
USER nobody:nobody
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
COPY --from=build /gss ./
ENTRYPOINT ["/gss"]
