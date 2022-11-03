#!/bin/bash

cd $(dirname $0)
cd ..

echo "==> Updating this repo"

git pull

echo "==> Updating kubecare cluster manager"

wget $(curl -s https://api.github.com/repos/krecik/cluster-manager/releases/latest \
| grep browser_download_url \
| grep amd64 \
| grep linux \
| cut -d '"' -f 4) -O kubecare-cluster-manager.tgz \
&& tar -zxvf kubecare-cluster-manager.tgz \
&& chmod +x kubecare-cluster-manager

echo "==> Updating addons"

rm -fr addons
curl -LJO https://github.com/krecik/cluster-manager-addons/archive/refs/heads/master.zip
unzip cluster-manager-addons-master.zip
mv cluster-manager-addons-master addons
chmod -R a+w addons

echo "==> Done"

exit 0