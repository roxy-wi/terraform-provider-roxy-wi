---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "roxywi_user Resource - roxywi"
subcategory: ""
description: |-
  This resource manages user entries in Roxy-WI. It allows you to define users with specific email addresses, usernames, passwords, and enabled statuses.
---

# roxywi_user (Resource)

This resource manages user entries in Roxy-WI. It allows you to define users with specific email addresses, usernames, passwords, and enabled statuses.

## Example Usage

```terraform
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
```

## Schema

### Required

- `email` (String) The email of the user.
- `enabled` (Boolean) Whether the user is enabled (true for enabled, false for disabled).
- `password` (String, Sensitive) The password of the user.
- `username` (String) The username of the user.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--timeouts"></a>

### Nested Schema for `timeouts`

This resource supports the following timeouts:

Optional:

* `create` - Default is 10 minutes.
* `update` - Default is 10 minutes.
* `delete` - Default is 30 minutes.

## Import

In Terraform v1.7.0 and later, use an import block to import User. For example:

```terraform
import {
  to = roxywi_user.example
  id = "6"
}
```

Using terraform import, import User can be imported using the `id`, e.g. For example:

```shell
% terraform import roxywi_user.example 1
```