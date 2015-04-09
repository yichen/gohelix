#!/bin/sh

set -ex

apt-get update
apt-get install vim curl wget screen -q -y

yes | apt-get install default-jre

export HELIX_INSTALL_ROOT=/opt
export HELIX_HOSTNAME=192.168.44.77
export REPOSITORY_ROOT=/vagrant

sh /vagrant/vagrant/install_cluster.sh
sh /vagrant/vagrant/setup_services.sh
