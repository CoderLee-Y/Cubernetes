# Cubernetes
Course project for SE3356.

> - 在首次启动，需要下载weave, 可能需要20-30秒。尽量不要在这时候apply api object:)
> - 测试时可以用Script下的clear.sh清理残留的docker容器。[风险提示: Docker和IPtables会被清空]

## Quick start

在开始之前，您需要安装ETCD, 放到合适的目录下(Suggest: `/usr/local/bin`). 建议安装Nginx以获取默认的配置作为Volume. 

```shell
apt-get install resolvconf
```

此外，需要安装Kafka. (在Master)

```shell
wget https://dlcdn.apache.org/kafka/3.2.0/kafka_2.13-3.2.0.tgz
tar -xzvf kafka_2.13-3.2.0.tgz
cp -r ./kafka_2.13-3.2.0/* /usr/local/kafka/
# copy config *service to 
vim /etc/systemd/system/kafka.service
vim /etc/systemd/system/zookeeper.service

systemctl daemon-reload
systemctl restart zookeeper
systemctl status zookeeper
systemctl restart kafka
systemctl status kafka

systemctl enable kafka
systemctl enable zookeeper
```

## Test Routines

测试用例在./example/presentation下。

### Init

```shell
./build/cuberoot init -f ./example/yaml/master-node.yaml
./build/cuberoot join 192.168.1.6 -f ./example/yaml/slave-node.yaml

# 可以展示
weave status peers
```

### Pod

```shell
# 展示基本Pod操作，command和资源限制
./build/cubectl apply -f ./example/yaml/presentation/pod/stress.yaml
# top / log看下CPU占用，应该是2个VM大致平分一个Core

./build/cubectl get pods
./build/cubectl describe pod

# 展示停止pod
./build/cubectl delete pod f8f5ce4e-5819-493e-a9c0-408b5c9b4560
# 已经没有这个Docker了

# 展示挂载卷，容器内localhost，暴露端口
# 记得回忆下tc的index.html有没有写入
# 这个Pod会被调度到另一台机器上
./build/cubectl apply -f ./example/yaml/presentation/pod/tn.yaml
# 展示Volume的网页
curl 127.0.0.1:8095
curl 127.0.0.1:8090

docker exec -it xxx bash
curl 127.0.0.1:8080
# 至此，Pod要求已经全部实现了
```

### GPU

由于GPU任务可能发生排队，可以提前检查。

```shell
./build/cubectl apply -f ./example/yaml/presentation/gpu/gpu-add.yaml
./build/cubectl apply -f ./example/yaml/presentation/gpu/gpu-mult.yaml

# Job也有基本的RR负载均衡
./build/cubectl get gpuJobs
./build/cubectl describe GpuJob
```

### Svc

这部分将和RS, DNS一起检查，先检查SVC与负载均衡，再检查杀死Pod之后的行为；最后检查DNS.

```shell
./build/cubectl apply -f ./example/yaml/presentation/rs.yaml

./build/cubectl get rs
./build/cubectl apply -f ./example/yaml/presentation/svc-1.yaml
./build/cubectl apply -f ./example/yaml/presentation/svc-2.yaml

# 回填svc ID
curl 172.16.0.0:8080
curl 172.16.0.1:80

./build/cubectl apply -f ./example/yaml/presentation/dns.yaml
./build/cubectl describe dns xxx

curl example.cubernetes.weave.local/test/cubernetes/nb
curl example.cubernetes.weave.local/test/cubernetes/very/nb

# 展示docker内也可以访问DNS与Svc
docker exec it ...

# 删除Pod尝试
./build/cubectl get rs
./build/cubectl describe rs x
./build/cubectl delete pod
# Waiting...
./build/cubectl describe rs
```

### AS

```shell
# Keep scaling
./build/cubectl apply -f ./example/yaml/presentation/autoscaler-cpu.yaml

./build/cubectl get autoscaler 
./build/cubectl describe autoscaler 732f4c11-165c-445e-9779-ea8bfb369d62
# Scale to more, then drop to 1
./build/cubectl apply -f ./example/yaml/presentation/autoscaler-memory.yaml

# waiting...
./build/cubectl describe autoscaler
```

### 容错

```shell
./build/cuberoot stop
# 检查Pod IP和Service
curl podIP
curl serviceIP

./build/cuberoot start
# 检查Pod IP和Service
curl podIP
curl serviceIP
```

### Schedule

这部分将运行2个符合仅符合一个node标签的Pod, 观察他们是否被部署到了唯一符合的node上。同时作为容错恢复后的检查。

```shell
./build/cubectl apply -f ./example/yaml/presentation/onemoreRS.yaml

./build/cubectl describe pod
# Only stay in worker 1
```

### Serverless

需要保证所有机器上拥有镜像：`yiyanleee/python-runtime:v1.5` & `yiyanleee/serverless-gateway:v1`

```shell
./build/cuberoot serverless enable
./build/cubectl apply -f example/serverless/hello/hello.yaml
./build/cubectl apply -f example/serverless/hello/ingress.yaml

curl "172.16.0.0:6810/hello?name=serverless" # Change to service IP
```

### Finally

```shell
./build/cuberoot reset
./build/cuberoot stop
bash ./scripts/clear.sh
```



