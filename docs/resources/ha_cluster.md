---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "roxywi_ha_cluster Resource - roxywi"
subcategory: ""
description: |-
  Managing HA cluster resources.
---

# roxywi_ha_cluster (Resource)

Managing HA cluster resources.

## Example Usage

```terraform
provider "roxywi" {
  base_url = "https://..."
  login    = "your_login"
  password = "your_password"
}

resource "roxywi_ha_cluster" "example" {
  description = "Example HA"
  virt_server = true
  use_src     = false
  name        = "example listener"
  syn_flood   = false
  vip         = "10.0.0.127"

  servers {
    id     = 1
    eth    = "eth0"
    master = true
  }

  servers {
    id     = 29
    eth    = "eth0"
    master = false
  }

  services {
    name    = "haproxy"
    docker  = false
    enabled = true
  }

  services {
    name    = "nginx"
    docker  = false
    enabled = false
  }

  services {
    name    = "apache"
    docker  = false
    enabled = false
  }
}
```


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `description` (String) Description of the HA Cluster.
- `name` (String) Name of the HA Cluster.
- `servers` (Block List, Min: 1) List of servers in the HA Cluster. (see [below for nested schema](#nestedblock--servers))
- `vip` (String) Virtual IP address for the HA Cluster.

### Optional

- `return_master` (Boolean) Return to master setting for the HA Cluster.
- `services` (Block List) Services configuration for the HA Cluster. (see [below for nested schema](#nestedblock--services))
- `syn_flood` (Boolean) SYN flood protection setting for the HA Cluster.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `use_src` (Boolean) Use source setting for the HA Cluster.
- `virt_server` (Boolean) Virtual server setting for the HA Cluster.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--servers"></a>
### Nested Schema for `servers`

Required:

- `eth` (String) Ethernet interface for the server.
- `id` (Number) Server ID.
- `master` (Boolean) Master setting for the server.


<a id="nestedblock--services"></a>
### Nested Schema for `services`

Required:

- `docker` (Boolean) Docker setting for the service.
- `enabled` (Boolean) Enabled status for the service.
- `name` (String) Name of the service.


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
  to = roxywi_ha_cluster.example
  id = "6"
}
```

Using terraform import, import Group can be imported using the `id`, e.g. For example:

```shell
% terraform import roxywi_ha_cluster.example 1
```
