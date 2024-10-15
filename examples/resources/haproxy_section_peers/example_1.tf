provider "roxywi" {
  base_url = "https://..."
  login    = "test"
  password = "testpass"
}

resource "roxywi_haproxy_section_peers" "example" {
  name = "example"
  peers {
    name = "demo"
    ip   = "127.0.0.1"
    port = 887
  }
  peers {
    name = "test"
    ip   = "127.0.0.2"
    port = 887
  }
  server_id = 1

}