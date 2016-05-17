#!/bin/bash
set -e
version=2.0.3
rm -rf hekalain heka-lain.tgz rebellion
tmp_image='rebellion_build'
registry='registry.aliyuncs.com'
docker rm -f $tmp_container
docker rmi -f $tmp_image
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
