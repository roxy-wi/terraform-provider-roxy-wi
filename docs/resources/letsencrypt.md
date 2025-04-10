---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "roxywi_letsencrypt Resource - roxywi"
subcategory: ""
description: |-
  Manage Let's Encrypt certificates.
---

# roxywi_letsencrypt (Resource)

Manage Let's Encrypt certificates.

## Example Usage

```terraform
provider "roxywi" {
  base_url = "https://..."
  login    = "testlog"
  password = "testpass"
}

resource "roxywi_letsencrypt" "example" {
  email     = "test23@gmail.com"
  domains   = ["exmaple.com", "example2.com"]
  type      = "route53"
  api_key   = "aws_access_key_id"
  api_token = "aws_secret_access_key"
  server_id = 1
}
```


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `domains` (List of String) A list of user letsencrypts.
- `server_id` (Number) The ID of the server to deploy to.
- `type` (String) What challenge should be used. Available: 'standalone', 'route53', 'digitalocean', 'cloudflare', 'linode'

### Optional

- `api_key` (String) Secret key to use for authentication in DNS API. For Route53.
- `api_token` (String) Token to use for authentication in DNS API. For Route53 it is the access key.
- `description` (String) Description of the certificate.
- `email` (String) Email address to use for registration with Let's Encrypt.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)

## Import

In Terraform v1.7.0 and later, use an import block to import Group. For example:

```terraform
import {
  to = roxywi_letsencrypt.example
  id = "1"
}
```

Using terraform import, import Group can be imported using the `id`, e.g. For example:

```shell
% terraform import roxywi_letsencrypt.example 1
```
