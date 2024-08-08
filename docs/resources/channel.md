---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "roxywi_channel Resource - roxywi"
subcategory: ""
description: |-
  Represents a communication channel such as Telegram, Slack, PagerDuty, or Mattermost.
---

# roxywi_channel (Resource)

Represents a communication channel such as Telegram, Slack, PagerDuty, or Mattermost.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `channel` (String) The channel identifier.
- `group_id` (Number) The ID of the group to which the channel belongs.
- `receiver` (String) The type of the receiver. Only `telegram`, `slack`, `pd`, `mm` are allowed.
- `token` (String) The token used for the channel.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)