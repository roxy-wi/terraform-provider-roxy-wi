provider "roxywi" {
  base_url = "https://..."
  login    = "test"
  password = "testpass"
}

data "roxywi_udp_listener" "example" {
  id = 1
}

output "test" {
  value = data.roxywi_udp_listener.example
}