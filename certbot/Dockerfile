FROM golang:1.9.3-alpine AS build
ARG GIT_HEAD
ENV GIT_HEAD ${GIT_HEAD}
COPY src /go/src
WORKDIR /go
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-X main.BuildTime=`date --rfc-3339='seconds'` -X main.Version=$GIT_HEAD"  -o certbotHelper src/github.com/weishi258/nginx-certbot-swarm-docker/certbothelper/main/main.go


FROM certbot/certbot
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=build /go/certbotHelper .
RUN chmod +x /app/certbotHelper
ENTRYPOINT ["/app/certbotHelper", "-certs", "/etc/letsencrypt/certs.json", "-d", "/etc/letsencrypt/live"]
