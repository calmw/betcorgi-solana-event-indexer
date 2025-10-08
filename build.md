## 镜像构建

``` shell
# indexer 
docker buildx build --platform linux/amd64 --tag calmw/betcorgi_solana_indexer:0.0.1 --push .
```

``` shell
# indexer_api 
docker buildx build --platform linux/amd64 --tag calmw/betcorgi_solana_indexer_api:0.0.1 --push .
```

``` shell
# 启动 
docker compose -f docker-compose.yml up -d

```
