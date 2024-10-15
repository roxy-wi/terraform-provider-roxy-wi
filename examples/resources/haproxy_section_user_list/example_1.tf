provider "roxywi" {
  base_url = "https://..."
  login    = "test"
  password = "testpass"
}

resource "roxywi_haproxy_section_user_list" "example" {
  userlist_groups = ["group1", "group2"]
  userlist_users {
    user     = "user1"
    password = "password1"
    group    = "group2"
  }
  userlist_users {
    user     = "user2"
    password = "password2"
  }

  name      = "user_list"
  server_id = 1
}