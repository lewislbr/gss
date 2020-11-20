FROM golang:1-alpine AS build
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64
WORKDIR /
COPY . ./
RUN go build -o gss -ldflags '-s -w' src/main.go

FROM scratch
COPY --from=build /gss ./
