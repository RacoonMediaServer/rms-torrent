FROM golang as builder
WORKDIR /src/service
COPY . .
RUN apt-get update && apt-get install libsqlite3-dev
RUN go build -tags libsqlite3 -ldflags "-X main.Version=`git tag --sort=-version:refname | head -n 1`" -o rms-torrent rms-torrent.go
RUN CGO_ENABLED=0 GOOS=linux go build -o rms-torrent-cli ./cli/main.go
FROM frolvlad/alpine-glibc
RUN apk update && apk upgrade && apk --no-cache add ca-certificates tzdata sqlite libstdc++
RUN mkdir /app
WORKDIR /app
COPY --from=builder /src/service/rms-torrent .
COPY --from=builder /src/service/rms-torrent-cli .
COPY --from=builder /src/service/configs/rms-torrent.json /etc/rms/
CMD ["./rms-torrent"]