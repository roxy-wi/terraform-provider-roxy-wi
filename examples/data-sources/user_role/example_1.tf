provider "roxywi" {
  base_url = "https://..."
  login    = "test"
  password = "testpass"
}

data "roxywi_user_role" "example" {}

output "test" {
  value = data.roxywi_user_role.example.roles
}
