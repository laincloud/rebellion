#!/bin/bash
set -ex

# Install official filebeat
wget https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-5.4.0-x86_64.rpm
rpm -ivp filebeat-5.4.0-x86_64.rpm
rm -rf filebeat-5.4.0-x86_64.rpm

# Replace with lain_filebeat
lain_filebeat_version=5.4.0-lain-p3
git clone -b $lain_filebeat_version https://github.com/laincloud/beats.git $GOPATH/src/github.com/elastic/beats
go build -o /usr/share/filebeat/bin/filebeat github.com/elastic/beats/filebeat

# Install rebellion
rebellion_version=2.3.0
git clone -b $rebellion_version https://github.com/laincloud/rebellion.git $GOPATH/src/github.com/elastic/beats


rm -rf hekalain heka-lain.tgz rebellion
tmp_image='rebellion_build'
registry='registry.aliyuncs.com/laincloud'
docker build --no-cache -t $tmp_image .
tmp_container='rebellion_instance'
docker create --name $tmp_container rebellion_build
docker cp $tmp_container:/rebellion ./
docker rm -f $tmp_container
docker rmi -f $tmp_image

# Build hekalain
git clone https://github.com/laincloud/hekalain.git
./hekalain/build.sh
rm -rf hekalain

tar -xvf heka-lain.tgz
rm -f heka-lain.tgz
heka_dir=`ls -d heka-lain-*`
mv $heka_dir heka-lain

docker build --no-cache -t $registry/rebellion:$version -f Dockerfile.release .

rm -rf heka-lain rebellion
docker push $registry/rebellion:$version
