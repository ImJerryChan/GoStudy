# redis专题

## 1. redis介绍

### 1.1 什么是redis

redis是C语言开发的一个开源高性能(Key-Value)的NOSQL数据库



### 1.2 redis数据类型

redis提供了五种数据类型

* string（字符串）
* list（链表）
* set（集合）
* zset（有序集合）----典型应用：排序
* hash（哈希类型）



### 1.3 redis和memcached对比

* 持久化
  * redis可以做缓存，也可以做数据持久化（支持aof和rdb两种持久化方式）
  * memcache仅支持做缓存，不支持持久化
* 数据结构
  * redis有五种常用数据类型
  * memcached一般是字符串和对象



### 1.4 redis应用场景

* 内存数据库（登录信息、购物车信息、用户浏览器记录）
* 缓存服务器（商品数据、广告数据等）**（最多）**
* session共享
* 任务队列（秒杀业务）
* 分布式锁实现
* 支持发布订阅的消息模式
* 应用排行榜（有序集合）
* 网站访问统计
* 过期数据处理



## 2. redis工作模式

### 2.1 单机单实例模式

### ① 安装单机单实例redis

```
1. 安装支持依赖
> yum install -y gcc-c++
> yum install -y wget

2. 下载源码包
> wget http://download.redis.io/releases/redis-5.0.7.tar.gz

3. 编译安装redis
> tar xfv redis-5.0.7.tar.gz 
> cd redis-5.0.7/
> make

4. 创建二进制文件及配置文件目录
> mkdir -p /usr/local/redis-5.0.7/{bin,etc}
> mv redis.conf /usr/local/redis-5.0.7/etc/
> mv src/{mkreleasehdr.sh,redis-benchmark,redis-check-aof,redis-check-rdb,redis-cli,redis-server,redis-trib.rb} /usr/local/redis-5.0.7/bin/
  
5. 配置环境变量
  # 持久化配置需要修改/etc/profile文件
> PATH=$PATH:/usr/local/redis-5.0.7/bin

6. 启动redis
> redis-server
```

### ② 守护进程启动

需要修改配置文件：redis.conf

```
> vim etc/redis.conf

# 需要将daemonize 由`no`修改为`yes`
daemonize yes

# 设置监听地址
bind 192.168.186.141

# 设置监听端口
port 6379

# 设置redis数据库总数量
databases 255 （默认16）

# 设置日志文件
logfile /usr/local/redis-5.0.7/log/redis_6379.log

# 设置数据文件存储目录
dir /data/redis

# 是否开启保护模式，由yes改为no
protected-mode no
【这里需要后续测试】
```

修改完毕后，启动redis

```
> redis-server /usr/local/redis-5.0.7/redis.conf
```

关闭后端启动的redis

```
> redis-cli shutdown
```



### 2.2 主从模式

以下操作均在单台服务器中演示

#### ① 主从复制作用

* 主从备份：防止单点故障
* 读写分离：分担master任务
* 任务分离：如从redis分别承担备份和计算工作

#### ② redis主从复制的两种方式

<img src="https://github.com/ImJerryChan/GoStudy/blob/master/res/redis%E4%B8%93%E9%A2%98/2.2.1.png?raw=true"/>



#### ③ redis主从服务通信原理

<img src="https://github.com/ImJerryChan/GoStudy/blob/master/res/redis%E4%B8%93%E9%A2%98/2.2.2.png?raw=true"/>



#### ④ 主从模式配置

 如果主服务器上需要增加redis的密码，增加如下配置: 

```
requirepass xxxxxx
```

 redis的主, 从的安装方式, 步骤都一样, 从的配置文件从主拷贝过来, 然后在**从节点**配置文件中加上如下配置: 

```
slaveof 192.168.186.141 6379

#如果主上有密码, 则从服务器上的配置文件需要增加以下配置:
masterauth xxxxxx

# 如果参数设置为`yes`，当主从复制时/从跟主断开连接时，从服务器可以相应client请求（使用过期的数据）
# 如果参数设置为`no`，当主从复制时/从跟主断开连接时，从服务器会阻塞所有请求，client请求时返回"SYNC with master in progress"报错
# slave-serve-stale-data yes
```

#### ⑤ 测试主从同步

在master上新增数据

```
> redis-cli -p 6379
127.0.0.1:6379> set myname dazuo
OK
```

在slave上查看数据

```
> redis-cli -p 6380
127.0.0.1:6380> get myname
"dazuo"
```

redis主从默认是读写分离的

```
127.0.0.1:6380> set yourname jerry
(error) READONLY You can't write against a read only replica.
```



### 2.3 哨兵模式

* redis的哨兵模式（sentinel）是建立在主从模式上的，因为redis-2.8以前，当主redis异常中断服务后，需要手工完成主从redis的切换，难以实现自动化，所以有了哨兵模式。

* 简单的说, 哨兵模式就是增加的投票机制，增加几台服务器作为哨兵节点, 即监控节点, 如果超过半数的哨兵即: 2 / n + 1的个数认为主挂了，就会自动提升从服务器为主服务器，并且，哨兵是可以实时改动redis主从的配置文件的，而自己的配置文件是实时发生变化。
* 优点：可自动完成主从切换，无需人工介入、宕机服务器问题恢复后可自动加入集群
* 缺点：只有一个master，并发写请求较大时，无法缓解写压力、主从切换时，集群处于不可写状态

#### ① redis-sentinel结构

<img src="https://github.com/ImJerryChan/GoStudy/blob/master/res/redis%E4%B8%93%E9%A2%98/2.3.1.png?raw=true"/>



使用哨兵模式时，redis-sentinel架构如图所示，redis-sentinel本身是一个独立的进程，能监控多个redis的master-slave集群**（也就是说，可以在同一个redis-sentinel集群配置中监听多个redis集群）**，在发现master服务不可用之后，可以实现主从自主切换。

* 每个sentinel会以每秒一次的频率向所有已知的master和slave及其他sentinel实例发送一个PING命令
* 如果一个实例最后一次有效回复PING超过了 down-after-milliseconds 设置的值，则此实例会在此sentinel标记为主观下线(SDOWN)
* 只有足够数量的sentinel在指定时间内确认了master的主观下线状态时，此实例才被标记为客观下线(ODOWN)

#### ② sentinel作用

* 监控redis主节点是否正常运行
* 输出集群错误信息（主从切换）
* 当某个mater服务不可用（failover）时，通过选举一个slave升级为master，并修改其他slave的slaveof关系，更新client连接
* client通过sentinel获取redis地址，并在failover时更新地址

#### ③ sentinel集群

因为redis-sentinel本身也是一个独立的进程，单节点的redis-sentinel监控redis集群时不可靠的，因此redis-sentinel也需要实现高可用，即使某一个redis-sentinel宕掉了，依然可以进行redis集群的主备切换（如果有多个redis-sentinel，redis的客户端可以连接任意一个sentinel获取redis集群的信息）

我们至少需要三个redis-sentinel节点来保证集群稳定性。

#### ④ 配置sentinel

参考官网给出的配置

```
> vim /usr/local/redis-5.0.7/etc/sentinel-26379.conf 

port 26379
daemonize yes
logfile /usr/local/redis-5.0.7/log/sentinel-26379.log
dir /data/redis/sentinel/26379
sentinel monitor mymaster 192.168.186.141 6379 2
sentinel down-after-milliseconds mymaster 20000
sentinel failover-timeout mymaster 180000
sentinel parallel-syncs mymaster 1
```

* 第一行配置表示sentinel去监视一个名为mymaster的主服务器，主服务器IP地址为192.168.186.141，端口号为6379，当这个主服务器**至少被两个sentinel判断为主观下线（超过半数的哨兵即: 2 / n + 1的个数认为主挂了）**，此时标记此台服务器为ODOWN后，进行自动故障迁移
* 第二行配置表示当主服务器在给定的多少**毫秒**内没返回sentinel发送的PING命令的回复，或者返回报错，此时sentinel会标记此台服务器为主观下线(SDOWN)，但是自动故障迁移只有在被标记为客观下线（ODOWN）后才会执行，即足够数量的sentinel都将服务器标记为SDOWN时（在第一行处配置）
* 第三行配置表示
* 第四行配置表示在执行故障迁移时，最多可以有几个服务器**同时**对主服务器进行同步，数字越小，完成迁移的时间越长。若此数字较大，当redis.conf配置文件中设置了不允许使用过期数据集（见2.2主从模式配置：slave-serve-stale-data），主从切换时，可能会导致所有从服务器短时间内不可用，因此可以通过将这个值设为 1 来保证每次只有一个从服务器处于不能处理命令请求的状态。 

#### ⑤ 启动sentinel

需要根据实际修改sentinel监听端口port及日志输出路径

```
> redis-sentinel /usr/local/redis-5.0.7/etc/sentinel-26379.conf 
> redis-sentinel /usr/local/redis-5.0.7/etc/sentinel-26380.conf 
> redis-sentinel /usr/local/redis-5.0.7/etc/sentinel-26381.conf 
```

#### ⑥ go语言中哨兵模式使用示例

TODO



### 2.4 Redis-Cluster

#### ① 什么是Redis-Cluster

* Redis-Cluster是Redis的分布式解决方案，当遇到单机内存、并发、流量等瓶颈时，可以采用 Cluster 架构达到负载均衡的目的
* 去中心化，集群可增加1000个节点，性能随节点增加而线性扩展
* 管理方便，后续可自行增加或摘除节点，移动分槽等
* 可缓解哨兵模式单master的高并发写流量压力

#### ② 数据分布理论和redis数据分区

* 分布式数据库首要解决把整个数据集按照分区规则映射到多个节点 的问题，即把数据集划分到多个节点上，每个节点负责整个数据的一个子集。常见的分区规则有哈希分区和顺序分区。**Redis Cluster** 采用哈希分区规则。
* 虚拟槽分区巧妙地使用了哈希空间，使用分散度良好的哈希函数把 所有的数据映射到一个固定范围内的整数集合，整数定义为槽(slot)。比如 **Redis Cluster** 槽的范围是 **0** ~ **16383**。槽是集群内数据管理和迁移的基本单位。
* Redis Cluster 采用虚拟槽分区，所有的键根据哈希函数映射到 0 ~16383，计算公式:slot = CRC16(key)&16383。每一个节点负责维护一部 分槽以及槽所映射的键值数据。
* redis-cluster把所有的物理节点映射到[0-16383] 上,cluster 负责维护node <->slot <-> value

#### ③ Redis-Cluster架构示例(TODO：能否3*6)

我们以六个节点为例，介绍Redis-Cluster的体系架构，其中三个为master节点，三个为slave节点

<img src="https://github.com/ImJerryChan/GoStudy/blob/master/res/redis%E4%B8%93%E9%A2%98/2.4.1.png?raw=true"/>



#### ④ Redis-Cluster一致性（TODO： 更好理解）

 Redis 并不能保证数据的**强一致性**. 这意味着在实际中集群在特定的条件下可能会丢失写操作，原因如下：
 a. 第一个原因是因为集群使用了异步复制. 写操作过程：

- 客户端向主节点B写入一条命令
- 主节点B向客户端回复命令状态
- 主节点将写操作复制给他的从节点 B1, B2 和 B3
- 如果在B崩溃之前，B1没有接收到写入就被提升为主，则此条记录将会永远丢失写入
- 主节点对命令的复制工作发生在返回命令回复之后， 因为如果每次处理命令请求都需要等待复制操作完成的话， 那么主节点处理命令请求的速度将极大地降低 —— 因此必须在性能和一致性之间做出权衡 

 b.  另外一种可能会丢失命令的情况是集群出现了网络分区，并且一个客户端与至少包括一个主节点在内的少数实例被孤立

* 举个例子 假设集群包含 A 、 B 、 C 、 A1 、 B1 、 C1 六个节点， 其中 A 、B 、C 为主节点， A1 、B1 、C1 为A，B，C的从节点， 还有一个客户端 Z1 假设集群中发生网络分区，那么集群可能会分为两方，大部分的一方包含节点 A 、C 、A1 、B1 和 C1 ，小部分的一方则包含节点 B 和客户端 Z1 .

* Z1仍然能够向主节点B中写入, 如果网络分区发生时间较短,那么集群将会继续正常运作,如果分区的时间足够让大部分的一方将B1选举为新的master，那么Z1写入B中的数据便丢失了.

* 注意， 在网络分裂出现期间， 客户端 Z1 可以向主节点 B 发送写命令的最大时间是有限制的， 这一时间限制称为节点超时时间（node timeout）， 是 Redis 集群的一个重要的配置选项

  #### ⑤ Redis-Cluster搭建

Redis-Cluster至少需要三台主服务器，三台从服务器（本机测试可以使用一台机器的不同端口进行实验）

* 安装Ruby环境和Ruby Redis接口 （redis 5.X可以省略此步骤）

  ```
  1. 安装ruby环境
  > yum -y install ruby ruby-devel rubygems rpm-build
  
  
  2. 安装ruby redis接口
  > gem install redis
  
  【问题1】此步骤可能会出现命令无反应，那是因为科学上网，需要更改为国内源
  > gem sources --add https://gems.ruby-china.com/ --remove https://rubygems.org/
  > gem sources -l
  
  【问题2】安装过程中出现redis requires Ruby version >= 2.3.0，查看此链接解决（https://blog.csdn.net/chenxinchongcn/article/details/78666374）
  ```

* 以6个节点为例，安装部署Redis-Cluster

  ```
  # 主节点: 6379, 6380, 6381
  # 从节点: 6382, 6383, 6384
  # 每个配置文件需要修改成对应端口的配置
  
  > vim /usr/local/redis-5.0.7/etc/redis-6379.conf
  
  daemonize yes
  
  # 各个节点的端口不同
  port 6379
  
  # 开启集群服务
  cluster-enabled yes
  
  # 节点的配置文件名字, 需要更改成不同的端口
  cluster-config-file nodes/nodes-6379.conf 
  cluster-node-timeout 15000
  
  # rdb 文件名字改成不同的端口
  dbfilename dump6379.rdb
  appendonly yes
  
  # aof 文件名字改成不同的端口
  appendfilename "appendonly6379.aof"
  ```

* 启动redis实例

  ```
  redis-server /usr/local/redis-5.0.7/etc/redis-6379.conf
  redis-server /usr/local/redis-5.0.7/etc/redis-6380.conf
  redis-server /usr/local/redis-5.0.7/etc/redis-6381.conf
  redis-server /usr/local/redis-5.0.7/etc/redis-6382.conf
  redis-server /usr/local/redis-5.0.7/etc/redis-6383.conf
  redis-server /usr/local/redis-5.0.7/etc/redis-6384.conf
  ```

  确认redis进程已启动

  ```
  ps aux|grep redis  # 确认服务已启动，若未启动，查看log定位问题原因
  ```

  <img src="https://github.com/ImJerryChan/GoStudy/blob/master/res/redis%E4%B8%93%E9%A2%98/2.4.2.png?raw=true"/>

* 使用 redis-trib.rb 自动部署方式 

  ```
  # redis 5.X之前
  redis-trib.rb create --replicas 1 192.168.186.141:6379 192.168.186.141:6380 192.168.186.141:6381 192.168.186.141:6382 192.168.186.141:6383 192.168.186.141:6384 --cluster-replicas 1
  ```

  **“--replicate ”：**表示为集群中每一个主节点创建一个从节点

  ```
  # redis 5.X之后，官方推荐使用redis-cli
  redis-cli --cluster create 192.168.186.141:6379 192.168.186.141:6380 192.168.186.141:6381 192.168.186.141:6382 192.168.186.141:6383 192.168.186.141:6384 --cluster-replicas 1
  ```

  **“--cluster-replicas ”：**表示为集群中每一个主节点创建一个从节点

  <img src="https://github.com/ImJerryChan/GoStudy/blob/master/res/redis%E4%B8%93%E9%A2%98/2.4.3.png?raw=true"/>开始配置集群

  <img src="https://github.com/ImJerryChan/GoStudy/blob/master/res/redis%E4%B8%93%E9%A2%98/2.4.4.png?raw=true"/>

  分配完成后，查看redis进程，发现六个redis-server已成功加入到集群当中

  <img src="https://github.com/ImJerryChan/GoStudy/blob/master/res/redis%E4%B8%93%E9%A2%98/2.4.5.png?raw=true"/>

* 测试Redis-Cluster

  使用命令行客户端登录

  ```
  > redis-cli -c -p 6379
  
  # 这里需要带上“-c”参数，如果不带上，可能会出现报错“(error) MOVED 8172 192.168.186.141:6380”
  ```

  <img src="https://github.com/ImJerryChan/GoStudy/blob/master/res/redis%E4%B8%93%E9%A2%98/2.4.6.png?raw=true"/>
  
* 测试高可用

  <img src="https://github.com/ImJerryChan/GoStudy/blob/master/res/redis%E4%B8%93%E9%A2%98/2.4.7.png?raw=true"/>

  关闭6380前，6380为其中一个master节点；关闭6380测试后，发现6382被选举成了新的master节点，各节点配置文件更新为集群最新主备配置

  <img src="https://github.com/ImJerryChan/GoStudy/blob/master/res/redis%E4%B8%93%E9%A2%98/2.4.8.png?raw=true"/>

#### ⑥ 维护节点

* 添加主节点

  a. 新增一个redis-server服务，具体方法参考2.1守护进程启动，这里启动一个端口为7006的redis-server

  b. 新增主节点

  redis 5.X之前：

  ```
  > redis-trib.rb add-node 192.168.186.141:7006 192.168.186.141:6379
  ```

  redis 5.X之后：

  ```
  redis-cli --cluster add-node 192.168.186.141:7006 192.168.186.141:6379
  ```

  c. 确认是否添加成功

  ```
  > redis 127.0.0.1:7006> cluster nodes
  ```

  新节点现在已经连接上了集群， 成为集群的一份子， 并且可以对客户端的命令请求进行转向了， 但是和其他主节点相比， 新节点还有两点区别：

  **d. 新节点没有包含任何数据， 因为它没有包含任何哈希槽.**

  **e. 尽管新节点没有包含任何哈希槽， 但它仍然是一个主节点， 所以在集群需要将某个从节点升级为新的主节点时， 这个新节点不会被选中**（可以参考2.4搭建Redis-Cluster中的测试高可用，7006并未被选择成为新的master）

* hash槽重新分配(数据迁移)

  a. 连接集群任意一个节点，记录节点信息

  ```
  > redis-cli -c -p 6379
  ```

  <img src="https://github.com/ImJerryChan/GoStudy/blob/master/res/redis%E4%B8%93%E9%A2%98/2.4.9.png?raw=true"/>

  b. 从6379迁移数据到7006

  ```
  > redis-cli --cluster reshard 192.168.186.141:6379 --cluster-from af692bd8b8be1f76deba6608aa0734080283b0a3 --cluster-to eb7d976aadb72b6061469ed978d3125542e08630
  ```

  c. 确认数据迁移

  <img src="https://github.com/ImJerryChan/GoStudy/blob/master/res/redis%E4%B8%93%E9%A2%98/2.4.10.png?raw=true"/>

  d. 确认是否迁移完成

  <img src="https://github.com/ImJerryChan/GoStudy/blob/master/res/redis%E4%B8%93%E9%A2%98/2.4.11.png?raw=true"/>

* 添加从节点

  同添加主节点，添加从节点步骤完全一致，可参考上面

  redis 5.X之前：

  ```
  > redis-trib.rb add-node --slave 192.168.186.141:7007 192.168.186.141:7006
  ```

  redis 5.X之后：

  ```
  > redis-cli --cluster add-node 192.168.186.141:7007 192.168.186.141:7006 --cluster-slave
  ```

* 删除节点

  redis 5.X之前：
  
  ```
  > redis-trib.rb del-node 192.168.186.141:6379 ${node_id}
  ```
  
redis 5.X之后：
  
  ```
  > redis-cli --cluster del-node 192.168.186.141:6379 ${node_id}
  ```
  
  如果删除hash槽不为空的节点，会报错“[ERR] Node 127.0.0.1:7005 is not empty!Reshard data away and try again.”，需要先把hash槽分配出去后再进行删除
  
  

## 3. redis数据类型

## 4. redis的消息模式

## 5. redis持久化

Redis提供了多种不同级别的持久化方式

* **RDB**持久化可以在指定的时间间隔内生成数据集的时间点快照(point-in-time snapshot) 
* **AOF(Append-only file)** 持久化记录服务器执行的所有写操作命令，并在服务器启动时，通过重新执行这些命令来还原数据集。AOF 文件中的命令全部以 Redis 协议的格式来保存，新命令会被追加到文件的末尾。Redis 还可以在后台对 AOF 文件进行重写(rewrite)， 使得 AOF 文件的体积不会超出保存数据集状态所需的实际大小<img src="https://img2018.cnblogs.com/blog/651008/201901/651008-20190111113443240-859893500.png" style="zoom:50%;" />
* Redis还可以同时使用RDB和AOF持久化，此配置下Redis重启时优先使用AOF还原数据集，因为通常AOF比RDB保存的数据集更完整
* 你甚至可以关闭持久化功能，让数据只在服务器运行时存在

### 5.1 rdb

#### ① save策略

 SNAPSHOTTING的持久化方式有多种save策略可供选择，而且支持混用，在默认的配置文件redis.conf中，配置如下：

```
save 900 1
save 300 100
save 60  10000
```

 上述配置的效果是：snapshotting会在3个条件中的**任何一个满足时被触发**： 

```
a. 900s内至少1个key有变化；
b. 300s内至少100个key有变化；
c. 60s内至少有10000个key有变化
```

RDB也支持主动备份策略

```
redis 127.0.0.1:6379> SAVE
OK

# 在后台执行备份
127.0.0.1:6379> BGSAVE
Background saving started
```

#### ② RDB优点

- RDB是一种表示某个即时点的Redis数据的紧凑文件。RDB文件适合用于备份。例如，你可能想要每小时归档最近24小时的RDB文件，每天保存近30天的RDB快照。这允许你很容易的恢复不同版本的数据集以容灾。
- RDB非常适合于灾难恢复，作为一个紧凑的单一文件，可以被传输到远程的数据中心，或者是Amazon S3(可能得加密)。
- RDB最大化了Redis的性能，因为Redis父进程持久化时唯一需要做的是启动(fork)一个子进程，由子进程完成所有剩余工作。父进程实例不需要执行像磁盘IO这样的操作。
- RDB在重启保存了大数据集的实例时比AOF要快。
- 总的来说，就是**适用于“有准备的灾难恢复”**

#### ③ RDB缺点

- 当你需要在Redis停止工作(例如停电)时最小化数据丢失，RDB可能不太好。你可以配置不同的保存点(save point)来保存RDB文件(例如，至少5分钟和对数据集100次写之后，但是你可以有多个保存点)。然而，你通常每隔5分钟或更久创建一个RDB快照，所以一旦Redis因为任何原因**没有正确关闭**而停止工作，你就得做好最近几分钟**数据丢失**的准备了。
- RDB需要经常调用fork()子进程来持久化到磁盘。**如果数据集很大的话，fork()比较耗时**，结果就是，当数据集非常大并且CPU性能不够强大的话，Redis会停止服务客户端几毫秒甚至一秒。AOF也需要fork()，但是你可以调整多久频率重写日志而不会有损(trade-off)持久性(durability)。

### 5.2 aof

#### ① 配置aof

AOF持久化方式在默认的配置文件redis.conf中，配置如下：

```
appendonly       yes
appendfilename   "appendonly6379.aof"
appendfsync      everysec
```

对于参数 appendfsync 解析如下：

```
no：      redis不主动调用fsync,何时刷盘由os来调度
always：  redis针对每个写入命令俊辉主动调用fsync刷磁盘
eversec： 每秒调用一次fsync刷盘
```

#### ② AOF优点

- 使用AOF Redis会更具有可持久性(durable)：你可以有很多不同的fsync策略：没有fsync，每秒fsync，每次请求时fsync。使用默认的每秒fsync策略，写性能也仍然很不错(fsync是由后台线程完成的，主线程继续努力地执行写请求)，即便你也就仅仅只损失一秒钟的写数据。

- AOF日志是一个追加文件，所以不需要定位，在断电时也没有损坏问题。即使由于某种原因文件末尾是一个写到一半的命令(磁盘满或者其他原因),redis-check-aof工具也可以很轻易的修复。

- 当AOF文件变得很大时，Redis会自动在后台进行重写。重写是绝对安全的，因为Redis继续往旧的文件中追加，使用创建当前数据集所需的最小操作集合来创建一个全新的文件，一旦第二个文件创建完毕，Redis就会切换这两个文件，并开始往新文件追加。

- AOF文件里面包含一个接一个的操作，以易于理解和解析的格式存储。你也可以轻易的导出一个AOF文件。例如，即使你不小心错误地使用FLUSHALL命令清空一切，如果此时并没有执行重写，你仍然可以保存你的数据集，你只要停止服务器，删除最后一条命令，然后重启Redis就可以，具体操作方法如下：

  ```
  # 如果是单节点redis（没做主从或者是redis-cluster），在不小心执行了`FLUSHALL`后
    1. 执行命令`SHUTDOWN NOSAVE`
    2. 找到对应的aof文件，删除文件最后的`FLUSHALL`记录，保存文件
    3. 重启redis服务
  
  # 如果做了redis主从或Redis-Cluster
    1. 执行命令`INTO REPLICATION`找到此节点的对应关系（假设当前节点是主节点）
    2. 执行命令`SHUTDOWN NOSAVE`
    3. 到从节点上执行命令`SHUTDOWN NOSAVE`
    4. 分别找到对应的aof文件，删除文件最后的`FLUSHALL`记录，保存文件
    5. 重启主从节点redis服务
  ```

#### ③ AOF缺点

- 对同样的数据集，AOF文件通常要大于等价的RDB文件。
- AOF可能比RDB慢，这取决于准确的fsync策略。通常fsync设置为每秒一次的话性能仍然很高，如果关闭fsync，即使在很高的负载下也和RDB一样的快。不过，即使在很大的写负载情况下，RDB还是能提供能好的最大延迟保证。
- 在过去，我们经历了一些针对特殊命令(例如，像BRPOPLPUSH这样的阻塞命令)的罕见bug，导致在数据加载时无法恢复到保存时的样子。这些bug很罕见，我们也在测试套件中进行了测试，自动随机创造复杂的数据集，然后加载它们以检查一切是否正常，但是，这类bug几乎不可能出现在RDB持久化中。为了说得更清楚一点：Redis AOF是通过递增地更新一个已经存在的状态，像MySQL或者MongoDB一样，而RDB快照是一次又一次地从头开始创造一切，概念上更健壮。但是，1)要注意Redis每次重写AOF时都是以当前数据集中的真实数据从头开始，相对于一直追加的AOF文件(或者一次重写读取老的AOF文件而不是读内存中的数据)对bug的免疫力更强。2)我们还没有收到一份用户在真实世界中检测到崩溃的报告。

#### ④ AOF持久性如何

你可以配置 Redis 多久才将数据 fsync 到磁盘一次。有三个选项：

- 每次有新命令追加到 AOF 文件时就执行一次 fsync ：非常慢，也非常安全。
- 每秒 fsync 一次：足够快（和使用 RDB 持久化差不多），并且在故障时只会丢失 1 秒钟的数据。
- 从不 fsync ：将数据交给操作系统来处理。更快，也更不安全的选择。

推荐（并且也是默认）的措施为每秒 fsync 一次， 这种 fsync 策略可以兼顾速度和安全性。

总是 fsync 的策略在实际使用中非常慢， 即使在 Redis 2.0 对相关的程序进行了改进之后仍是如此 —— 频繁调用 fsync 注定了这种策略不可能快得起来。





## 6. redis事务

## 7. redis的分布式锁

## 8. redis故障诊断和优化

## 9. redis和go开发



参考链接：[故事凌-redis专题]( https://mp.weixin.qq.com/s/3ER1vUSqz195-8OXSHGpxw )
