provider "roxywi" {
  base_url = "https://..."
  login    = "testlog"
  password = "testpass"
}

resource "roxywi_user" "example" {
  email    = "test23@gmail.com"
  enabled  = true
  password = "testpassword"
  username = "testuser2"
}

resource "roxywi_user_role_binding" "example" {
  user_id  = roxywi_user.example.id
  role_id  = 1
  group_id = 1
}
