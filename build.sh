#!/bin/bash
set -ex

# Install lain filebeat
FILEBEAT_VERSION='5.4.0_lain_p2'
FILEBEAT_TAG='v5.4.0-lain-p2'
curl -L -o filebeat-${FILEBEAT_VERSION}-1.x86_64.rpm https://github.com/laincloud/beats/releases/download/${FILEBEAT_TAG}/filebeat-${FILEBEAT_VERSION}-1.x86_64.rpm
rpm -ivp filebeat-${FILEBEAT_VERSION}-1.x86_64.rpm
rm -f filebeat-${FILEBEAT_VERSION}-1.x86_64.rpm

# Install rebellion
go build -o /usr/local/bin/rebellion github.com/laincloud/rebellion
cp $GOPATH/src/github.com/laincloud/rebellion/templates/filebeat.yml.tmpl /etc/filebeat/
cp $GOPATH/src/github.com/laincloud/rebellion/conf/supervisord.conf /etc/supervisord.conf
rm -rf $GOPATH/src/github.com