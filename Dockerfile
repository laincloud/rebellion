FROM laincloud/centos-lain:20160428.1
# Dockerfile for building rebellion

ENV dest $GOPATH/src/laingit.bdp.cc/lain/rebellion

#Build rebellion
COPY . $dest/
RUN cd $dest && go build -o /rebellion main.go
