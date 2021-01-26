# imanager

#### 介绍
基于属性基加密的轻量级用户身份认证鉴权微服务

#### 编译

```shell script
bash build/build.sh
```

#### 部署
将build/deploy目录下的文件拷贝至master节点，执行以下命令
```shell script
kubectl create secret generic imanager -n eec --from-file=master_key --from-file=pub_key

kubectl create -f imanager-deployment.yaml

kubectl create -f imanager-service.yaml
```

#### 调试相关
本地启动容器
```shell script
docker run -it -p 8080:8080 10.5.26.86:8080/zjlab/cpabe:1.0 bash
mkdir -p /home/zjlab/secret
```
另开一个后端，将相关的密钥文件拷贝至容器内部
```shell script
imanagerPath="/mnt/hgfs/GOProject/src/imanager"
cd ${imanagerPath}/build/deploy
containerID=$(docker ps | grep cpabe | awk '{print $1}')
docker cp master_key ${containerID}:/home/zjlab/secret
docker cp pub_key ${containerID}:/home/zjlab/secret
```
编译二进制，并拷贝进容器
```shell script
cd ${imanagerPath}/cmd
go build  -o imanager .; docker cp imanager ${containerID}:/home/zjlab/
```
在容器内部，启动调试进程
```shell script
HarborAddress=http://10.5.26.86:8080 HarborUser=admin HarborPassword=Harbor12345 \
/home/zjlab/imanager --encryptDir /home/zjlab/secret/ --httpport 8080 --logtostderr
```