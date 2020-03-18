FROM golang as builder

WORKDIR /go/src/racoondev.tk/gitea/racoon/rtorrent

COPY . .

RUN go get
RUN CGO_ENABLED=0 GOOS=linux go build -o rtorrent -a -installsuffix cgo rtorrent.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN mkdir /app
WORKDIR /app
COPY --from=builder /go/src/racoondev.tk/gitea/racoon/rtorrent .

CMD ["./rtorrent"]