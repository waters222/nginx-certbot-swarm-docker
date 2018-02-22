FROM golang:1.9.3-alpine AS build
ARG GIT_HEAD
ENV GIT_HEAD ${GIT_HEAD}
COPY nginx-proxy/src /go/src
WORKDIR /go
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-X main.BuildTime=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.Version=$GIT_HEAD"  -o generator src/github.com/weishi258/nginx-certbot-swarm-docker/generator/main/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-X main.BuildTime=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.Version=$GIT_HEAD"  -o watcher src/github.com/weishi258/nginx-certbot-swarm-docker/watcher/main/main.go

FROM nginx:mainline-alpine
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=build /go/generator .
COPY --from=build /go/watcher .
COPY nginx-proxy/templates /app/templates
