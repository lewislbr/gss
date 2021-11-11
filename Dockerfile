FROM golang:1.17-alpine AS build
WORKDIR /
COPY . ./
ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOARCH=amd64
ENV GOOS=linux
RUN go build -o gss -ldflags="-s -w" gss.go

FROM scratch
USER nobody:nobody
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
COPY --from=build /gss ./
ENTRYPOINT ["/gss"]
