#### ID Generator
distributed ID generator similar to twitter snowflake: https://github.com/twitter/snowflake

thirft server, using framed binary protocol. interface definition as [thrift definition](idgenerator.thrift)

##### Installation
Require Golang environment
```
go get git.apache.org/thrift.git/lib/go/thrift
go install github.com/liusf/idgenerator
```
##### Usage
show help:
```
idgenerator -h
Usage of idgenerator:
  -dc int
    	data center id (0-7)
  -h	show this help info
  -p int
    	port to listen to
  -w int
    	worker id (0-31)
  -zk string
    	check and register with zookeepers(ip:port,ip:port,..)
    	
如:
idgenerator -p 3456 -w 1 -zk localhost:2181 &
```
启动的服务会自动注册到zookeeper的 /service/idgenerators 路径下:
```
zkCli.sh
[zk: localhost:2181(CONNECTED) 17] ls /service/idgenerators
[member_0, member_1]
[zk: localhost:2181(CONNECTED) 18] get /service/idgenerators/member_0
{"serviceEndpoint":{"host":"liusf-mac","port":3457},"additionalEndpoints":{},"status":"ALIVE","shard":0}
  	
```
##### Java客户端调用
```
安装到本地maven:
mvn install

依赖的Java Project中增加maven依赖:
<dependency>
    <groupId>idgenerator</groupId>
    <artifactId>idgenerator</artifactId>
    <version>0.0.1-SNAPSHOT</version>
</dependency>
```
使用 finagle 调用该服务:
```
maven依赖:
<dependency>
    <groupId>com.twitter</groupId>
    <artifactId>finagle-thrift_2.11</artifactId>
    <version>6.33.0</version>
</dependency>

Java: 
String serverAddr = "zk!localhost:2181!/service/idgenerators"
IdGenerator.FutureIface idGenerator = Thrift.client().
    newIface(serverAddr, "idgenerator_client", IdGenerator.FutureIface.class);
Long id = idGenerator.getId("ORDER").get();
    
```