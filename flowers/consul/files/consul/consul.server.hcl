data_dir = "/opt/consul/data"

server           = true
bootstrap_expect = 1
retry_join       = ["10.0.0.11"] # The first contoller server

# Enable the built-in web UI server
ui_config = {
  enabled = true
}

# Bind address of the Consul protocols.
# Explicitly bind on a private address, otherwise it will bind to 0.0.0.0
bind_addr = "{{ GetPrivateInterfaces | include \"network\" \"10.0.0.0/8\" | attr \"address\" }}"

telemetry = {
  # deprecated and resource hungry
  "disable_compat_1.9" = true
}
