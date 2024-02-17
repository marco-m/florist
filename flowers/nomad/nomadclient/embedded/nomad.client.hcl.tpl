data_dir = "/opt/nomad/data"
datacenter = "<<.Workspace>>"

# Bind address of the Nomad protocols (not of the tasks).
# The default is to bind to "0.0.0.0".
# We would like to bind to (1) a private address and (2) localhost, BUT Nomad
# (contrary to Consul) only wants one bind address, otherwise it will fail with
# > Bind address resolution failed: multiple addresses found ("127.0.0.1 10.0.0.11"), please configure one
# So I don't know what to do :-(
# bind_addr = "{{ GetPrivateInterfaces | include \"network\" \"10.0.0.0/8\" | attr \"address\" }}"

addresses {
  http = "127.0.0.1"
}

advertise {
  http = "127.0.0.1"
}

ui {
  enabled = true
  label {
    text = "<<.Workspace>>"
  }
}

client {
  enabled = true

  server_join {
    retry_join = [ "10.0.0.10", "10.0.0.11", "10.0.0.12", "10.0.0.13" ]
  }

  # Interface to allocate address:port for tasks.
  # By default Nomad will select the interface with the default route, while
  # in our case we want tasks to listen on a private address.
  network_interface = "{{ GetPrivateInterfaces | include \"network\" \"10.0.0.0/8\" | attr \"name\" }}"
}

tls {
  http = false
  rpc  = true

  ca_file   = "config/NomadAgentCaPub"
  cert_file = "config/GlobalClientNomadKeyPub"
  key_file  = "config/GlobalClientNomadKey"

  verify_server_hostname = true
  verify_https_client    = false
}