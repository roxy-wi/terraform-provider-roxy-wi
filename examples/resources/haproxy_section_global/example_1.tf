provider "roxywi" {
  base_url = "https://..."
  login    = "test"
  password = "testpass"
}

resource "roxywi_haproxy_section_global" "example" {
  server_id = 1
  chroot    = "/var/lib/haproxy"
  daemon    = true
  socket    = ["*:1999 level admin", "/var/run/haproxy.sock mode 600 level admin", "/var/lib/haproxy/stats"]
  log       = ["127.0.0.1 local1", "127.0.0.1 local1 notice"]
  maxconn   = 5000
  action    = "restart"
}