FROM golang as builder

WORKDIR /go/src/racoondev.tk/gitea/racoon/rms-torrent

ARG GIT_PASSWORD=$GIT_PASSWORD

COPY . .

RUN echo "machine racoondev.tk\tlogin racoon\n\tpassword $GIT_PASSWORD" >> ~/.netrc \
    && go get

RUN CGO_ENABLED=0 GOOS=linux go build -o rms-torrent -a -installsuffix cgo rms-torrent.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN mkdir /app
WORKDIR /app
COPY --from=builder /go/src/racoondev.tk/gitea/racoon/rms-torrent .

CMD ["./rms-torrent"]