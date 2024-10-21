provider "roxywi" {
  base_url = "https://..."
  login    = "test"
  password = "testpass"
}

resource "roxywi_haproxy_section_defaults" "example" {
  server_id = 1
  action    = "restart"
  timeout {
    check = 11
  }
}