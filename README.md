#### ID Generator
distributed ID generator similar to twitter snowflake: https://github.com/twitter/snowflake

thirft server, using framed binary protocol. interface definition as [thrift definition](idgenerator.thrift)

##### Installation
Require Golang environment
```
cd $GOPATH/src
mkdir git.apache.org
cd git.apache.org
git clone https://github.com/apache/thrift thrift.git
cd thrift.git
git checkout 0.10.0
cd ..
go get github.com/liusf/idgenerator
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
##### Note
重新从thrift生成go源码后,需要修改**idgenerator.go**文件如下函数代码:
```
func (p *IdGeneratorProcessor) Process(iprot, oprot thrift.TProtocol) (success bool, err thrift.TException) {
	name, _, seqId, err := iprot.ReadMessageBegin()
	if err != nil {
		return false, err
	}
	if processor, ok := p.GetProcessorFunction(name); ok {
		return processor.Process(seqId, iprot, oprot)
	}
	iprot.Skip(thrift.STRUCT)
	iprot.ReadMessageEnd()
	x11 := thrift.NewTApplicationException(thrift.UNKNOWN_METHOD, "Unknown function "+name)
	oprot.WriteMessageBegin(name, thrift.EXCEPTION, seqId)
	x11.Write(oprot)
	oprot.WriteMessageEnd()
	oprot.Flush()
	// 原来是这样: return false, x11
	return true, x11 // 修改成这样
}
```
