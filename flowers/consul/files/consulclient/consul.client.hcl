data_dir = "/opt/consul/data"

server     = false
retry_join = ["10.0.0.11"] # The first contoller server

# Bind address of the Consul protocols.
# Explicitly bind on a private address, otherwise it will bind to 0.0.0.0
bind_addr = "{{ GetPrivateInterfaces | include \"network\" \"10.0.0.0/8\" | attr \"address\" }}"

telemetry = {
  # deprecated and resource hungry
  "disable_compat_1.9" = true
}
