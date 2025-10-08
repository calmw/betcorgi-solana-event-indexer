## 镜像构建

``` shell
# relayer node 
docker build -t harbor.devops.tantin.com/chain/tt_bridge:0.1.0 .
docker push harbor.devops.tantin.com/chain/tt_bridge:0.1.0

docker buildx build --tag harbor.devops.tantin.com/chain/tt_bridge:0.1.0 --push .
```

``` shell
# api 
docker buildx build --platform linux/amd64 --tag harbor.devops.tantin.com/chain/tt_bridge_api:0.2.5 --push .
```

``` shell
# bridge 
docker buildx build --platform linux/amd64 --tag harbor.devops.tantin.com/chain/tt_bridge:0.4.1 --push .
```

``` shell
# dfr 
docker buildx build --platform linux/amd64 --tag harbor.devops.tantin.com/chain/tt_bridge_dfr:0.2.1 --push .
```

``` shell
# 启动 
docker compose -f docker/docker-compose_dev.yml up -d

```

``` shell
# 进入容器
docker exec -it tt_bridge_api /bin/bash
# 然后执行
cd / && bash -c './bridge_api 2> >(tee /proc/1/fd/2 >&2) | tee /proc/1/fd/1'
```

``` shell
# 进入容器
docker exec -it tt_bridge_one /bin/bash
# 查看是否已经运行
ps -ef | grep tt_bridge | grep -v grep
# 然后执行
cd / && bash -c './tt_bridge 2> >(tee /proc/1/fd/2 >&2) | tee /proc/1/fd/1'
```

``` shell
# 启进入容器
mysql -h 127.0.0.1 -P 3306 -u root -p bridge
```

## 命令行程序构建

``` shell
# relayer node 
go build -o tb -trimpath cmd/deploy/main.go
```

``` shell
# EVM使用示例 
./tb tron --admin_address 'TFBymbm7LrbRreGtByMPRD2HUyneKabsqb' --fee_address 'TFBymbm7LrbRreGtByMPRD2HUyneKabsqb' --server_address 'TFBymbm7LrbRreGtByMPRD2HUyneKabsqb' --relayer_one_address  'TTgY73yj5vzGM2HGHhVt7AR7avMW4jUx6n'   --relayer_two_address  'TSARBFH6PW6jEuf8chd1DxZGW6JEmHuv6g' --relayer_three_address 'TEz4CMzy3mgtVECcYxu5ui9nJfgv3oXhyx' --fee 4 --passphrase '123456' --key 'XXXXXXX'
```

``` shell
# TRON使用示例 
./tb tron --admin_address 'TFBymbm7LrbRreGtByMPRD2HUyneKabsqb' --fee_address 'TFBymbm7LrbRreGtByMPRD2HUyneKabsqb' --server_address 'TFBymbm7LrbRreGtByMPRD2HUyneKabsqb' --relayer_one_address  'TTgY73yj5vzGM2HGHhVt7AR7avMW4jUx6n'   --relayer_two_address  'TSARBFH6PW6jEuf8chd1DxZGW6JEmHuv6g' --relayer_three_address 'TEz4CMzy3mgtVECcYxu5ui9nJfgv3oXhyx' --fee 4 --key 'XXXXXXX'
```

