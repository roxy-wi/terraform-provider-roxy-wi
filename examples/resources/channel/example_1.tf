provider "roxywi" {
  base_url = "https://..."
  login    = "testlog"
  password = "testpass"
}

resource "roxywi_channel" "example" {
  receiver = "pd"
  channel  = "test_my_channel"
  group_id = 1
  token    = "some_token"
}
