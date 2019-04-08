#!/bin/bash
#Disable upgrade related services
systemctl disable locksmithd
systemctl stop locksmithd
systemctl disable update-engine
systemctl stop update-engine

#Fix mis-configuration of dockerd
mkdir -p /etc/docker
echo '{ "storage-driver": "devicemapper" }' > /etc/docker/daemon.json
sed -i '/Environment=DOCKER_SELINUX=--selinux-enabled=true/s/^/#/g' /run/systemd/system/docker.service

mkdir -p '/'
cat << EOF | base64 -d > '/foo'
YmFy
EOF
chmod '0600' '/foo'

cat << EOF | base64 -d > '/etc/systemd/system/docker.service'
dW5pdA==
EOF

mkdir -p '/etc/systemd/system/docker.service.d'
cat << EOF | base64 -d > '/etc/systemd/system/docker.service.d/10-docker-opts.conf'
b3ZlcnJpZGU=
EOF

META_EP=http://100.100.100.200/latest/meta-data
PROVIDER_ID=`curl -s $META_EP/region-id`.`curl -s $META_EP/instance-id`
echo PROVIDER_ID=$PROVIDER_ID > $DOWNLOAD_MAIN_PATH/provider-id
echo PROVIDER_ID=$PROVIDER_ID >> /etc/environment

systemctl daemon-reload
systemctl restart docker
systemctl enable 'docker.service' && systemctl restart 'docker.service'