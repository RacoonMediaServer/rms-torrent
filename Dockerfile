FROM golang as builder
WORKDIR /src/service
COPY . .
RUN apt-get update && apt-get install tzdata fuse libfuse-dev -y
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags "-X main.Version=`git tag --sort=-version:refname | head -n 1`" -o rms-torrent rms-torrent.go
RUN CGO_ENABLED=0 GOOS=linux go build -o rms-torrent-cli ./cli/main.go  \
    && mkdir /app \
    && cp /src/service/rms-torrent /app/ \
    && cp /src/service/rms-torrent-cli /app/ \
    && mkdir -p /etc/rms \
    && cp /src/service/configs/rms-torrent.json /etc/rms/
WORKDIR /app
CMD ["./rms-torrent"]