package roxywi

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceChannelImport(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceChannel().Schema, nil)
	d.SetId("telegram:42")

	resources, err := resourceChannelImport(context.Background(), d, nil)
	if err != nil {
		t.Fatalf("resourceChannelImport: %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("expected one imported resource, got %d", len(resources))
	}
	if got := d.Id(); got != "42" {
		t.Errorf("unexpected imported ID %q", got)
	}
	if got := d.Get(ReceiverField); got != ReceiverTypeTelegram {
		t.Errorf("unexpected receiver %q", got)
	}
}

func TestResourceChannelImportRejectsInvalidID(t *testing.T) {
	for _, id := range []string{"1", "unknown:1", "telegram:", ":1", "telegram:not-a-number", "telegram:0"} {
		d := schema.TestResourceDataRaw(t, resourceChannel().Schema, nil)
		d.SetId(id)
		if _, err := resourceChannelImport(context.Background(), d, nil); err == nil {
			t.Errorf("expected import ID %q to fail", id)
		}
	}
}
