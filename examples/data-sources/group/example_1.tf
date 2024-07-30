provider "roxywi" {
  base_url = "https://..."
  login    = "test"
  password = "testpass"
}

data "roxywi_group" "example_id" {
  id = "4"
}

output "view" {
  value = data.roxywi_group.example_id
}

// ------------------------------------

data "roxywi_group" "example_name" {
  name = "test"
}

output data {
  value = data.roxywi_group.example_name
}
