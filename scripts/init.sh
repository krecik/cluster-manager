#!/bin/bash

cd $(dirname $0)
cd ..

echo "==> Updating this repo"

git pull

echo "==> Updating kubecare cluster manager"

wget $(curl -s https://api.github.com/repos/kubecare/cluster-manager/releases/latest \
| grep browser_download_url \
| grep amd64 \
| grep linux \
| cut -d '"' -f 4) -O kubecare-cluster-manager.tgz \
&& tar -zxvf kubecare-cluster-manager.tgz \
&& chmod +x kubecare-cluster-manager

echo "==> Updating addons"

if [ -e addons ]
then
  cd addons
  git pull
else
  git clone https://github.com/kubecare/cluster-manager-addons.git addons
  chmod -R a+w addons # to allow future updates of the repo
fi

echo "==> Done"

exit 0