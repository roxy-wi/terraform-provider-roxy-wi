provider "roxywi" {
  base_url = "https://demo.roxy-wi.org"
  login    = "testlog"
  password = "testpass"
}

resource "roxywi_user" "example" {
  email    = "test23@yandex.ru"
  enabled  = true
  password = "testpassword"
  username = "testuser2"
}

resource "roxywi_user_role_binding" "example" {
  user_id = roxywi_user.example.id
  role_id = 1
  group_id = 1
}