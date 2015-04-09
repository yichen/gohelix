#/bin/sh

set -ex

apt-get install supervisor -q -y
cp /vagrant/vagrant/zookeeper.conf /etc/supervisor/conf.d/
cp /vagrant/vagrant/helixcontroller.conf /etc/supervisor/conf.d/

supervisorctl reload
