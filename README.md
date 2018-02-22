# Build Certbot Docker
```bash
    docker build -t --build-arg GIT_HEAD=$(git rev-parse HEAD) water258/auto-certbot -f certbot/Dockerfile .
```

# Build Nginx Docker
```bash
    docker build --build-arg GIT_HEAD=$(git rev-parse HEAD) -t water258/swarm-nginx -f nginx-proxy/Dockerfile .
```