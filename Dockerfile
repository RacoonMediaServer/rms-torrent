FROM golang as builder
WORKDIR /go/src/git.rms.local/RacoonMediaServer/rms-torrent
ARG GIT_PASSWORD=$GIT_PASSWORD
COPY . .
RUN go env -w GOPRIVATE=git.rms.local \
  && go env -w GOINSECURE=git.rms.local  \
  && rm -rf .git \
  && echo "192.168.1.133	git.rms.local" > /etc/hosts \
  && git config --global url."http://racoon:$GIT_PASSWORD@git.rms.local/".insteadOf "http://git.rms.local/"  \
  && go get
RUN CGO_ENABLED=0 GOOS=linux go build -o rms-torrent -a -installsuffix cgo rms-torrent.go
FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN mkdir /app
WORKDIR /app
COPY --from=builder /go/src/git.rms.local/RacoonMediaServer/rms-torrent/rms-torrent .
CMD ["./rms-torrent"]