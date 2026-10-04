package roxywi

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func backupScheduleSchema() *schema.Schema {
	return &schema.Schema{Type: schema.TypeList, Computed: true, Description: "Scheduler state in Roxy-WI 9.1+; empty if no schedule exists.", Elem: &schema.Resource{Schema: map[string]*schema.Schema{
		"timezone":           {Type: schema.TypeString, Computed: true},
		"next_run_at":        {Type: schema.TypeString, Computed: true},
		"retry_at":           {Type: schema.TypeString, Computed: true},
		"migration_required": {Type: schema.TypeBool, Computed: true},
		"last_task_id":       {Type: schema.TypeInt, Computed: true},
		"last_status":        {Type: schema.TypeString, Computed: true},
	}}}
}

func readBackupSchedule(d *schema.ResourceData, value interface{}) diag.Diagnostics {
	schedule := []interface{}{}
	if value != nil {
		state, ok := value.(map[string]interface{})
		if !ok {
			return diag.Errorf("unexpected backup schedule format")
		}
		schedule = append(schedule, state)
	}
	if err := d.Set("schedule", schedule); err != nil {
		return diag.Errorf("set backup schedule: %v", err)
	}
	return nil
}
