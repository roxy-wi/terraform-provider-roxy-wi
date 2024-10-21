resource "roxywi_haproxy_section_listen" "example" {
  name           = "example-listen"
  mode           = "tcp"
  ssl_offloading = true
  balance        = "roundrobin"
  server_id      = 1
  waf            = true
  headers {
    path   = "http-response"
    method = "add-header"
    name   = "test-header"
    value  = "test"
  }
  binds {
    ip   = "0.0.0.0"
    port = 8088
  }
  backend_servers {
    server     = "127.0.0.1"
    port       = "8080"
    port_check = "8080"
  }
  backend_servers {
    server     = "127.0.0.2"
    port       = "8080"
    port_check = "8080"
  }
}