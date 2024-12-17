provider "roxywi" {
  base_url = "https://..."
  login    = "test"
  password = "testpass"
}

resource "roxywi_haproxy_list" "example" {
  name      = "example"
  server_ip = "127.0.0.1"
  content   = <<EOF
10.0.0.1
10.0.0.2
10.0.0.3
EOF
  action    = "reload"
  color     = "white"

}