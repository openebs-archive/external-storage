# -*- mode: ruby -*-
# vi: set ft=ruby :

$script = <<SCRIPT
# Update apt and get dependencies
sudo apt-get update
sudo apt-get install -y zip unzip curl wget vim

# Setup go
GO_VERSION="1.8"

# Get the ARCH
ARCH=`uname -m | sed 's|i686|386|' | sed 's|x86_64|amd64|'`

# Install Go
cd /tmp
wget -q https://storage.googleapis.com/golang/go${GO_VERSION}.linux-${ARCH}.tar.gz
sudo tar -xf go${GO_VERSION}.linux-${ARCH}.tar.gz -C /usr/local
rm go*.tar.gz

# Setup Go Path
SRCROOT="/usr/local/go"
SRCPATH="/home/vagrant/go"
sudo mkdir -p $SRCPATH
sudo chown -R vagrant:vagrant $SRCPATH
mkdir -p $SRCPATH/bin
mkdir -p $SRCPATH/src
mkdir -p $SRCPATH/pkg

cat <<EOF >/tmp/gopath.sh
export GOPATH="$SRCPATH"
export GOBIN="$SRCPATH/bin"
export GOROOT="$SRCROOT"
export PATH="$SRCROOT/bin:$SRCPATH/bin:\$PATH"
EOF
sudo mv /tmp/gopath.sh /etc/profile.d/gopath.sh
sudo chmod 0755 /etc/profile.d/gopath.sh
source /etc/profile.d/gopath.sh

#Install build dependencies
go get github.com/Masterminds/glide

SCRIPT

required_plugins = %w(vagrant-cachier)

required_plugins.each do |plugin|
  need_restart = false
  unless Vagrant.has_plugin? plugin
    system "vagrant plugin install #{plugin}"
    need_restart = true
  end
  exec "vagrant #{ARGV.join(' ')}" if need_restart
end


# All Vagrant configuration is done below. The "2" in Vagrant.configure
# configures the configuration version (we support older styles for
# backwards compatibility). Please don't change it unless you know what
# you're doing.
Vagrant.configure(2) do |config|
  vmName = "openebs-k8s-provisioner-dev"

  # Every Vagrant development environment requires a box. You can search for
  # boxes at https://atlas.hashicorp.com/search.
  config.vm.box = "bento/ubuntu-16.04"

  # Disable automatic box update checking. If you disable this, then
  # boxes will only be checked for updates when the user runs
  # `vagrant box outdated`. This is not recommended.
  if Vagrant.has_plugin?("vagrant-vbguest") then
    config.vbguest.auto_update = false
  end

  config.vm.provider "virtualbox" do |vb|
      #vb.gui = true
      vb.memory = 2048
      vb.cpus = 4
      vb.customize ["modifyvm", :id, "--cableconnected1", "on"]
  end

  config.vm.define vmName do |vmCfg|
      vmCfg.vm.hostname = vmName
      #Adding Vagrant-cachier
      if Vagrant.has_plugin?("vagrant-cachier")
        vmCfg.cache.scope = :machine
        vmCfg.cache.enable :apt
        vmCfg.cache.enable :gem
      end
  end

  config.vm.provision "shell", inline: $script, privileged: false
  config.vm.provision "docker" # Just install it
  config.vm.synced_folder ".", "/home/vagrant/go/src/github.com/openebs/openebs-k8s-provisioner/"
end
