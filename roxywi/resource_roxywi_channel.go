package roxywi

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	ReceiverField          = "receiver"
	ChannelField           = "channel"
	TokenField             = "token"
	ReceiverTypeTelegram   = "telegram"
	ReceiverTypeSlack      = "slack"
	ReceiverTypePagerDuty  = "pd"
	ReceiverTypeMattermost = "mm"
)

func resourceChannel() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceChannelCreate,
		ReadContext:   resourceChannelRead,
		UpdateContext: resourceChannelUpdate,
		DeleteContext: resourceChannelDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceChannelImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Description: "Represents a communication channel such as Telegram, Slack, PagerDuty, or Mattermost.",

		Schema: map[string]*schema.Schema{
			ReceiverField: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  fmt.Sprintf("The type of the receiver. Only `%s`, `%s`, `%s`, `%s` are allowed.", ReceiverTypeTelegram, ReceiverTypeSlack, ReceiverTypePagerDuty, ReceiverTypeMattermost),
				ValidateFunc: validation.StringInSlice([]string{ReceiverTypeTelegram, ReceiverTypeSlack, ReceiverTypePagerDuty, ReceiverTypeMattermost}, true),
			},
			ChannelField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The channel identifier.",
			},
			GroupIDField: {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The ID of the group to which the channel belongs.",
			},
			TokenField: {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The token used for the channel.",
			},
		},
	}
}

func resourceChannelImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("expected channel import ID in the format 'receiver:id', got %q", d.Id())
	}

	switch parts[0] {
	case ReceiverTypeTelegram, ReceiverTypeSlack, ReceiverTypePagerDuty, ReceiverTypeMattermost:
	default:
		return nil, fmt.Errorf("unsupported channel receiver %q", parts[0])
	}
	if id, err := strconv.Atoi(parts[1]); err != nil || id <= 0 {
		return nil, fmt.Errorf("channel ID must be a positive integer, got %q", parts[1])
	}

	if err := d.Set(ReceiverField, parts[0]); err != nil {
		return nil, fmt.Errorf("set imported channel receiver: %w", err)
	}
	d.SetId(parts[1])
	return []*schema.ResourceData{d}, nil
}

func resourceChannelCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	receiver := d.Get(ReceiverField).(string)

	channel := map[string]interface{}{
		ReceiverField: receiver,
		ChannelField:  strings.ReplaceAll(d.Get(ChannelField).(string), "'", ""),
		GroupIDField:  d.Get(GroupIDField).(int),
		TokenField:    d.Get(TokenField).(string),
	}

	resp, err := client.doRequest(ctx, "POST", fmt.Sprintf("/api/channel/%s", receiver), channel)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	id, ok := result["id"].(float64)
	if !ok {
		return diag.Errorf("unable to find ID in response: %v", result)
	}

	d.SetId(fmt.Sprintf("%d", int(id)))
	return resourceChannelRead(ctx, d, m)
}

func resourceChannelRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id := d.Id()
	receiver := d.Get(ReceiverField).(string)

	resp, err := client.doRequest(ctx, "GET", fmt.Sprintf("/api/channel/%s/%s", receiver, id), nil)
	if err != nil {
		return readDiagnostics(d, err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	if receiverValue, ok := result[ReceiverField].(string); ok && receiverValue != "" {
		if err := d.Set(ReceiverField, receiverValue); err != nil {
			return diag.Errorf("set Terraform state: %v", err)
		}
	}

	if channelValue, ok := result[ChannelField].(string); ok && channelValue != "" {
		channel := strings.ReplaceAll(channelValue, "'", "")
		if err := d.Set(ChannelField, channel); err != nil {
			return diag.Errorf("set Terraform state: %v", err)
		}
	}

	if groupIDValue, ok := result[GroupIDField].(float64); ok {
		if err := d.Set(GroupIDField, int(groupIDValue)); err != nil {
			return diag.Errorf("set Terraform state: %v", err)
		}
	}

	if tokenValue, ok := result[TokenField].(string); ok && tokenValue != "" {
		if err := d.Set(TokenField, tokenValue); err != nil {
			return diag.Errorf("set Terraform state: %v", err)
		}
	}

	return nil
}

func resourceChannelUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id := d.Id()
	receiver := d.Get(ReceiverField).(string)

	channel := map[string]interface{}{
		ReceiverField: d.Get(ReceiverField).(string),
		ChannelField:  strings.ReplaceAll(d.Get(ChannelField).(string), "'", ""),
		GroupIDField:  d.Get(GroupIDField).(int),
		TokenField:    d.Get(TokenField).(string),
	}

	_, err := client.doRequest(ctx, "PUT", fmt.Sprintf("/api/channel/%s/%s", receiver, id), channel)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceChannelRead(ctx, d, m)
}

func resourceChannelDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id := d.Id()
	receiver := d.Get(ReceiverField).(string)

	_, err := client.doRequest(ctx, "DELETE", fmt.Sprintf("/api/channel/%s/%s", receiver, id), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
