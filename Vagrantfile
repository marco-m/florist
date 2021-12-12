Vagrant.require_version ">= 2.2.18"
Vagrant.configure("2") do |config|
  config.vm.box = "debian/bullseye64" # debian/11
  config.vm.hostname = "florist-dev"
  config.vm.synced_folder ".", "/vagrant", disabled: true
  config.vm.synced_folder ".", "/home/vagrant/florist"
  config.vm.box_check_update = false

  config.vm.provider "virtualbox" do |v|
    v.name = "florist-dev" # name in VirtualBox GUI
    # v.linked_clone = true
    v.check_guest_additions = false
    v.memory = 2048
    v.cpus = 2
  end
end
