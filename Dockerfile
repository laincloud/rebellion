FROM laincloud/centos-lain:20170103
# Dockerfile for building rebellion

ENV dest $GOPATH/src/github.com/laincloud/rebellion

#Build rebellion
COPY . $dest/
RUN cd $dest && go build -o /rebellion main.go
