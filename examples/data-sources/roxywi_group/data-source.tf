provider "roxywi" {
  base_url = "https://..."
  login    = "test"
  password = "testpass"
}

data "roxywi_group" "example" {
  id = "4"
}

output "view" {
  value = data.roxywi_group.example
}