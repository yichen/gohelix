# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "hashicorp/precise64"
  config.vm.box_url = "https://vagrantcloud.com/hashicorp/boxes/precise64/versions/1.1.0/providers/virtualbox.box"

  config.vm.provision :shell, path: "vagrant/provision.sh"
  
  config.vm.network "private_network", ip: "192.168.44.77"
  config.vm.network :forwarded_port, guest: 2181, host: 2181

end
