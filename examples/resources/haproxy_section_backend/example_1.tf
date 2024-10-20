resource "roxywi_haproxy_section_backend" "example" {
  name           = "example-backend"
  mode           = "tcp"
  ssl_offloading = true
  balance        = "roundrobin"
  server_id      = 1
  health_check {
    check = "tcp-check"
  }
  headers {
    path   = "http-response"
    method = "add-header"
    name   = "test-header"
    value  = "test"
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