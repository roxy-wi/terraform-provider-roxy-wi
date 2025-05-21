resource "roxywi_nginx_section_upstream" "example" {
  name      = "example"
  server_id = 1
  balance   = "ip_hash"
  keepalive = 33
  backend_servers {
    server       = "127.0.0.1"
    port         = 8080
    max_fails    = 33
    fail_timeout = 3
  }
  backend_servers {
    server       = "127.0.0.2"
    port         = 8081
    max_fails    = 3
    fail_timeout = 13
  }
}
