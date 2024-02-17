data_dir = "/opt/consul/data"
datacenter = "<<.DataCenter>>"

server           = true
bootstrap_expect = <<.ConsulNumServers>>
retry_join       = ["10.0.0.11"] # The first controller server

# Multiple private IPv4 addresses found. Please configure one with 'bind'
bind_addr = "{{ GetPrivateInterfaces | include \"network\" \"10.0.0.0/8\" | attr \"address\" }}"

# Enable the built-in web UI server
ui_config = {
  enabled = true
}
