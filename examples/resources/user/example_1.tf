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
