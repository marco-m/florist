data_dir = "/opt/nomad/data"
datacenter = "<<.DataCenter>>"

# Bind address of the Nomad protocols (not of the tasks).
# The default is to bind to "0.0.0.0".
# We would like to bind to (1) a private address and (2) localhost, BUT Nomad
# (contrary to Consul) only wants one bind address, otherwise it will fail with
# > Bind address resolution failed: multiple addresses found ("127.0.0.1 10.0.0.11"), please configure one
# So I don't know what to do :-(
# bind_addr = "{{ GetPrivateInterfaces | include \"network\" \"10.0.0.0/8\" | attr \"address\" }}"

addresses {
  http = "127.0.0.1"
  rpc  = "{{ GetPrivateInterfaces | include \"network\" \"10.0.0.0/8\" | attr \"address\" }}"
  serf = "{{ GetPrivateInterfaces | include \"network\" \"10.0.0.0/8\" | attr \"address\" }}"
}

advertise {
  http = "127.0.0.1"
}

ui {
  enabled = true
  label {
    text = "<<.Environment>>"
  }
}

server {
  enabled          = true
  bootstrap_expect = <<.NomadNumServers>>

  server_join {
    retry_join = [ "10.0.0.10", "10.0.0.11", "10.0.0.12", "10.0.0.13" ]
  }
}

tls {
  http = false
  rpc  = true

  ca_file   = "config/NomadAgentCaPub"
  cert_file = "config/GlobalServerNomadKeyPub"
  key_file  = "config/GlobalServerNomadKey"

  verify_server_hostname = true
  verify_https_client    = false
}
