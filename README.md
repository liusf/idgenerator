#### ID Generator
distributed ID generator similar to twitter snowflake: https://github.com/twitter/snowflake

thirft server, using compact protocol. interface definition as [thrift definition](idgenerator.thrift)

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
Usage of ./idgenerator:
  -consul string
    	check peers with consul server/agent(ip:port,ip:port,...)
  -dc int
    	data center id (0-7)
  -h	show this help info
  -p int
    	port to listen to
  -w int
    	worker id (0-31)
```