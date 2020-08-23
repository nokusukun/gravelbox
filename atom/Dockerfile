FROM alpine:latest

LABEL "source"="gravelbox"

# Base
RUN apk update
RUN apk add python3
RUN apk add go
RUN apk add nodejs
RUN apk add mono --update-cache --repository http://dl-3.alpinelinux.org/alpine/edge/testing/ --allow-untrusted

# Additionals
COPY executor/executor.go /bin/executor.go
RUN go build -o /bin/executor /bin/executor.go