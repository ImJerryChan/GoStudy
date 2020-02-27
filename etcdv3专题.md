# etcdv3专题

## 1. etcdv3 搭建

### 1.1 etcd单节点

#### ① 编辑etcd安装脚本

如需使用https，需要自建证书，下面安装脚本是使用http的：

```shell
#!/bin/bash
# example: ./etcd_installer.sh etcd01 192.168.1.10 etcd02=http://192.168.1.11:2380,etcd03=http://192.168.1.12:2380
ETCD_NAME=$1
ETCD_IP=$2
ETCD_CLUSTER=$3

ETCD_VER=v3.3.8
WORK_DIR=/usr/local/etcd-${ETCD_VER}
GITHUB_URL=https://github.com/coreos/etcd/releases/download
DOWNLOAD_URL=${GITHUB_URL}

removeOldPackage() {
    echo "删除历史安装包...."
    rm -f /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz
    [ $? -eq 0 ] && rm -rf /tmp/etcd-download-test && mkdir -p /tmp/etcd-download-test
    [ $? -eq 0 ] && echo "删除历史安装包完成...."
}

downloadPackage() {
    echo "正在安装新安装包...."
    curl -L ${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz -o /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz
    [ $? -eq 0 ] && tar xzvf /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz -C /tmp/etcd-download-test --strip-components=1
    [ $? -eq 0 ] && rm -f /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz
    [ $? -eq 0 ] && echo "安装包下载完毕...."
}

moveBinFile() {
    echo "新建etcd工作目录...."
    if [ ! -d ${WORK_DIR} ]; then
        mkdir -p ${WORK_DIR}/{cfg,bin}
        echo "新建etcd工作目录成功...."
    else
        rm -rf ${WORK_DIR}
        [ $? -eq 0 ] && mkdir -p ${WORK_DIR}/{cfg,bin}
        echo "新建etcd工作目录成功...."
    fi

    echo "移动二进制文件到/usr/local/bin目录...."
    cp /tmp/etcd-download-test/{etcdctl,etcd} ${WORK_DIR}/bin
    [ $? -eq 0 ] && echo "etcd安装完毕，即将完成etcd配置并启动服务，请稍后...."
}

removeOldPackage
downloadPackage
moveBinFile

cat >${WORK_DIR}/cfg/etcd <<EOF 
#[Member]
ETCD_NAME="${ETCD_NAME}"
ETCD_DATA_DIR="/data/etcd"
ETCD_LISTEN_PEER_URLS="http://${ETCD_IP}:2380"
ETCD_LISTEN_CLIENT_URLS="http://${ETCD_IP}:2379"

#[Clustering]
ETCD_INITIAL_ADVERTISE_PEER_URLS="http://${ETCD_IP}:2380"
ETCD_ADVERTISE_CLIENT_URLS="http://${ETCD_IP}:2379"
ETCD_INITIAL_CLUSTER="${ETCD_NAME}=http://${ETCD_IP}:2380,${ETCD_CLUSTER}"
ETCD_INITIAL_CLUSTER_TOKEN="etcd-cluster"
ETCD_INITIAL_CLUSTER_STATE="new"
EOF

cat > /etc/systemd/system/etcd.service << EOF
[Unit]
Description=Etcd Server
Documentation=https://github.com/coreos/etcd
After=network.target
After=network-online.target
Wants=network-online.target

[Service]
Type=notify
Restart=on-failure
LimitNOFILE=65536
EnvironmentFile=${WORK_DIR}/cfg/etcd


ExecStart=${WORK_DIR}/bin/etcd \\
--name \${ETCD_NAME} \\
--data-dir \${ETCD_DATA_DIR} \\
--listen-peer-urls \${ETCD_LISTEN_PEER_URLS} \\
--listen-client-urls \${ETCD_LISTEN_CLIENT_URLS},http://127.0.0.1:2379 \\
--advertise-client-urls \${ETCD_ADVERTISE_CLIENT_URLS} \\
--initial-advertise-peer-urls \${ETCD_INITIAL_ADVERTISE_PEER_URLS} \\
--initial-cluster \${ETCD_INITIAL_CLUSTER} \\
--initial-cluster-token \${ETCD_INITIAL_CLUSTER_TOKEN} \\
--initial-cluster-state \${ETCD_INITIAL_CLUSTER_STATE}

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable etcd
systemctl restart etcd

export ETCDCTL_API=3
```

etcd的配置文件参数请参考：[ etcdv3 集群的搭建和使用 ]( https://yq.aliyun.com/articles/618674?spm=a2c4e.11155435.0.0.8a9042cfi1rr1y )

#### ② 安装并启动etcd服务

这里采用的是etcd单节点安装，因此只需要单台机器即可测试

```
> sh etcd_installer.sh etcd01 192.168.186.141
```

#### ③ 测试etcd服务是否正常

```
> etcdctl version
etcdctl version: 3.3.8
API version: 3.3

> etcdctl --endpoints="http://192.168.186.141:2379" put /test/url "hello"
OK
> etcdctl --endpoints="http://192.168.186.141:2379" get /test/url
/test/url
hello
```



### 1.2 etcdv3集群搭建

#### ① 安装etcd

这里是etcdv3集群安装，集群搭建建议使用https，需要准备三台虚拟机进行测试，安装脚本见上一小节:

```
> sh etcd_installer.sh etcd01 192.168.186.141 etcd02=https://192.168.186.142:2380,etcd03=https://192.168.186.143:2380
```

```
> sh etcd_installer.sh etcd02 192.168.186.142 etcd01=https://192.168.186.141:2380,etcd03=https://192.168.186.143:2380
```

```
> sh etcd_installer.sh etcd03 192.168.186.143 etcd01=https://192.168.186.141:2380,etcd02=https://192.168.186.142:2380
```

* 需要注意，使用https，需要在配置文件中加上证书配置，具体请看后续更新

  ```
  --cert-file=${WORK_DIR}/ssl/server.pem \
  --key-file=${WORK_DIR}/ssl/server-key.pem \
  --peer-cert-file=${WORK_DIR}/ssl/server.pem \
  --peer-key-file=${WORK_DIR}/ssl/server-key.pem \
  --trusted-ca-file=${WORK_DIR}/ssl/ca.pem \
  --peer-trusted-ca-file=${WORK_DIR}/ssl/ca.pem
  ```

#### ② 查看etcd集群状态

```she&amp;#39;l
> etcdctl --endpoints="https://192.168.186.141:2379,https://192.168.186.142:2379,https://192.168.186.143:2379" cluster-health

> etcdctl --endpoints="https://192.168.186.141:2379,https://192.168.186.142:2379,https://192.168.186.143:2379" endpoints status  --write-out="table"
```



## 2. etcd api 常用命令

以下命令均为ETCD的V3版本命令，因此需要设置如下参数

```shell
export ETCTCTL_API=3
```

### 2.1 增删查

#### ① 增

```go
etcdctl --endpoints=http://127.0.0.1:2379 put /node/111 value1
etcdctl --endpoints=http://127.0.0.1:2379 put /node/222 value2
```

#### ② 删

```go
// 删除一个key
etcdctl --endpoints=http://127.0.0.1:2379 del /node/222

// 删除以/node前缀开头的key
etcdctl --endpoints=http://127.0.0.1:2379 del /node --prefix
```

#### ③ 查

```go
etcdctl --endpoints=http://127.0.0.1:2379 get /node/111
etcdctl --endpoints=http://127.0.0.1:2379 get /node/222 --write-out="json"

//  -w, --write-out="simple" set the output format (fields, json, protobuf, simple, table)
```

其中列出所有etcd中存储的key

```go
etcdctl --endpoints=http://127.0.0.1:2379 get / --prefix --keys-only
```



### 2.2 租约相关

什么是租约lease，它和redis的expire time非常类似，我们可以申请一个TTL=10秒的租约，它会返回一个lease ID标志定时器，让我们在set一个key的时候我们可以带上这个lease ID，那么我们就可以实现一个自动过期的key，这样当过了10秒之后，携带这个租约的key（一个lease可以被多个key携带）就会过期并从etcd中删除。

当然，和redis不同的是，我们可以对etcd中定时的key进行续租，也就是重置它的TTL，这样就可以保证key永远不会过期**(基于这个特性，我们可以用来它做服务发现)**

#### ① 申请租约

```go
 etcdctl --endpointss=http://127.0.0.1:2379 lease grant 100
```

#### ② 租约绑定

```go
etcdctl --endpointss=http://127.0.0.1:2379 put --lease=2303706bddc07659 /node/333 value3
```

#### ③ 租约撤销

```go
etcdctl --endpointss=http://127.0.0.1:2379 lease revoke 2303706bddc07659
```

#### ④ 租约续租

```go
// 设置之后，每当到期就会自动续租
etcdctl --endpointss=http://127.0.0.1:2379 lease keep-alive 2303706bddc07659
```



## 3. 基于etcd的最简单的服务注册及服务发现

在go语言中，我们可以通过`github.com/coreos/etcd`包来进行etcd的管理，下面我们将使用此模块来说明etcd是如何工作的。基于etcd的服务注册及服务发现，可以用几句话大概描述：

**服务注册**

* 为服务注册新建一个etcd client
* 根据服务名称(prefix)注册我们所需要的服务，etcd的本质是一个有序的KV存储，因此我们可以模拟Linux文件系统的命名方式从而实现父子目录关系（如` /micro/user-srv，/micro/user-srv/loginService，/micro/user-srv/registerService `）
* 为注册到etcd的每个服务绑定一个TTL=N的lease，并定期重新注册(KeepAlive)

**服务发现**

* 为服务发现注册一个etcd client
* 根据服务名称(prefix)查找我们所需要的服务，并使用watch监听etcd中服务的变动

### 3.1 基本概念

* etcd client，我们服务注册或者服务发现，都需要先新建一个etcd客户端进行通信。其中服务注册会不定期地发送心跳信息给etcd服务并更新服务信息，告诉etcd我们的自定义的服务没有挂掉，而服务发现会watch在etcd中注册的服务，如果服务信息有更新(新增/下线)，我们都可以感知到
* lease，如果我们的服务因为异常宕掉，当我们没有给服务绑定TTL=N的租约时，订阅了这个服务的客户端可能就会返回一个异常结果，因此我们需要当这个服务宕掉之后它会在注册中心自动下线，etcd的租约lease可以实现这个功能

### 3.2 源码解析

#### ① 服务注册

下面是一个最简单的服务注册示例

```go
package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

var (
	err           error
	config        clientv3.Config
	client        *clientv3.Client
	keepResp      *clientv3.LeaseKeepAliveResponse
	keepRespChan  <-chan *clientv3.LeaseKeepAliveResponse
)

func main() {
	// 服务注册
	config = clientv3.Config{
		endpointss: []string{"127.0.0.:2379"},
	}
	// 新建一个etcd client客户端
	client, _ = clientv3.New(config)

	// 申请租约
	lease := clientv3.NewLease(client)
	// 设置租约过期时间为5秒
	leaseGrantResponse, _ := lease.Grant(context.TODO(), 5)
	leaseID := leaseGrantResponse.ID

	// KeepAlive会使得租约永不过期
	if keepRespChan, err = lease.KeepAlive(context.TODO(), leaseID); err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		for {
			select {
			case keepResp = <-keepRespChan:
				if keepResp == nil {
					fmt.Println("租约已失效")
					goto END
				} else {
					fmt.Println("收到自动续约应答， ", keepResp.ID)
				}
			}
		}
	END:
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	_, err = client.Put(ctx, "/myservice/userService", "I'm User Service, I can do a lot of things!", clientv3.WithLease(leaseID))

	// 防止context泄露
	cancel()

	if err != nil {
		fmt.Println(err)
		return
	}

	select {}
}
```



#### ② 服务发现

下面是一个最简单的服务发现示例

```go
package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

var (
	e           error
	clientDis   *clientv3.Client
	configDis   clientv3.Config
)


func main() {
	configDis = clientv3.Config{
		endpointss: []string{"127.0.0.1:2379"},
	}
	clientDis, e = clientv3.New(configDis)

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	getResp, err := clientDis.Get(ctx, "/myservice/userService")
	cancel()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(getResp.Kvs)
}
```

