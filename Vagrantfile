Vagrant.require_version ">= 2.4.1"
Vagrant.configure("2") do |config|
  config.vm.define "florist-dev" # name in vagrant CLI
  config.vm.box = "debian/bookworm64" # debian/12
  config.vm.hostname = "florist-dev"
  config.vm.synced_folder ".", "/vagrant", disabled: true
  # Must be disabled otherwise Goland ssh integration goes into an infinite loop
  # copying files into itself.
  #config.vm.synced_folder ".", "/home/vagrant/florist", disabled: true
  config.vm.synced_folder ".", "/home/vagrant/florist"
  config.vm.box_check_update = false
  config.vm.provider "virtualbox" do |v|
    v.name = "florist-dev" # name in VirtualBox GUI
    v.check_guest_additions = false
    v.memory = 2048
    v.cpus = 2
  end
end
