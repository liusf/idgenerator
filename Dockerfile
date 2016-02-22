FROM golang

VOLUME /opt

RUN apt-get update && apt-get install -y wget git make ; \
        cd /opt ; \
        export PATH=$GOROOT/bin:$GOPATH/bin:$PATH ; \
        go get -d github.com/liusf/idgenerator ; \
        cd $GOPATH/src/github.com/liusf/idgenerator ; \
        go build; cp idgenerator /usr/bin/ 

EXPOSE 2357

CMD ["/usr/bin/idgenerator", "-p", "2357"]