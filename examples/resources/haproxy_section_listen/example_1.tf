resource "roxywi_haproxy_section_listen" "example" {
  name = "example-listen"
  binds {
    ip =  "0.0.0.0"
    port = 80
  }
  mode = "http"
  balance = "roundrobin"
  server_id = 1
  backend_servers {
    server = "127.0.0.1"
    port = "8080"
    port_check = "8080"
  }
  backend_servers {
    server = "127.0.0.2"
    port = "8080"
    port_check = "8080"
  }
}
