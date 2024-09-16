provider "roxywi" {
  base_url = "https://demo.roxy-wi.org"
  login    = "your_login"
  password = "your_password"
}

resource "roxywi_udp_listener" "example" {
  config {
    backend_ip = "192.168.1.100"
    port = 9997
    weight = 50
  }
  сonfig {
    backend_ip = "192.168.2.100"
    port = 443
    weight = 50
  }
  description = "Example UDP listener"
  group_id = 2
  lb_algo = "rr"
  name = "example listener"
  port = 1234
  server_id = 1
  vip = "192.168.1.100"
}
