package roxywi

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intToBool(i float64) bool {
	return i == 1
}

func resourceParseId(fullId string, delimiter string) (string, string, error) {
	parts := strings.Split(fullId, delimiter)
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid ID format: %s", fullId)
	}
	return parts[0], parts[1], nil
}

func readDiagnostics(d *schema.ResourceData, err error) diag.Diagnostics {
	if isNotFound(err) {
		d.SetId("")
		return nil
	}
	return diag.FromErr(err)
}
