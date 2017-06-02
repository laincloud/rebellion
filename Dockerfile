FROM laincloud/centos-lain:20170405
# Dockerfile for building rebellion

RUN pip install supervisor && yum clean all

ENV dest $GOPATH/src/github.com/laincloud/rebellion

#Build rebellion
COPY . $dest/
RUN cd $dest && go build -o /rebellion main.go
