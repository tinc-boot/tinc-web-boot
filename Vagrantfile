# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.define "alfa" do |machine|
    machine.vm.box = "hashicorp/bionic64"
    machine.vm.hostname = "alfa"
    machine.vm.network "forwarded_port", guest: 8686, host: 18686, host_ip: "127.0.0.1"

    machine.vm.network "private_network", ip: "192.168.33.10"

    machine.vm.provider "virtualbox" do |vb|
      vb.memory = "1024"
    end
    machine.vm.provision "file", source: "./build/alfa", destination: "networks"
    machine.vm.provision "shell", inline: <<-SHELL
        cat networks/example-network/tinc.conf | grep Name
        sudo systemd-run --unit tinc-web-boot -p WorkingDirectory=`pwd` --no-block ./tinc-web-boot run --dev --headless --bind 0.0.0.0:8686
    SHELL
  end

  config.vm.define "beta" do |machine|
    machine.vm.box = "hashicorp/bionic64"
    machine.vm.hostname = "beta"
    machine.vm.network "forwarded_port", guest: 8686, host: 28686, host_ip: "127.0.0.1"

    machine.vm.network "private_network", ip: "192.168.33.20"

    machine.vm.provider "virtualbox" do |vb|
      vb.memory = "1024"
    end
    machine.vm.provision "file", source: "./build/beta", destination: "networks"
    machine.vm.provision "shell", inline: <<-SHELL
        cat networks/example-network/tinc.conf | grep Name
        sudo systemd-run --unit tinc-web-boot -p WorkingDirectory=`pwd` --no-block ./tinc-web-boot run --dev --headless --bind 0.0.0.0:8686
      SHELL
  end
  config.vm.provision "shell", inline: <<-SHELL
        apt-get update
        apt-get install -y tinc
        sudo systemctl stop tinc-web-boot || echo "not started"
  SHELL
  config.vm.provision "file", source: "./build/tinc-web-boot", destination: "tinc-web-boot"
end
