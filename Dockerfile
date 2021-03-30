FROM golang:1-alpine AS build
ENV CGO_ENABLED=0 \
    GO111MODULE=on \
    GOARCH=amd64 \
    GOOS=linux
WORKDIR /
COPY . ./
RUN go build -o gss -ldflags="-s -w" gss.go

FROM scratch
COPY --from=build /gss ./
ENTRYPOINT ["/gss"]
