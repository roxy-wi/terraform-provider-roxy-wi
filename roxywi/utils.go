package roxywi

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
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

func apiInt(result map[string]interface{}, field string) (int, error) {
	value, ok := result[field]
	if !ok || value == nil {
		return 0, fmt.Errorf("API response is missing field %q", field)
	}
	maxInt := int64(^uint(0) >> 1)
	minInt := -maxInt - 1

	switch value := value.(type) {
	case int:
		return value, nil
	case int32:
		return int(value), nil
	case int64:
		if value < minInt || value > maxInt {
			return 0, fmt.Errorf("API response field %q is outside the supported integer range", field)
		}
		return int(value), nil
	case float64:
		if math.Trunc(value) != value {
			return 0, fmt.Errorf("API response field %q must be an integer, got %v", field, value)
		}
		const maxExactJSONInteger = float64(1<<53 - 1)
		if value < float64(minInt) || value > float64(maxInt) || math.Abs(value) > maxExactJSONInteger {
			return 0, fmt.Errorf("API response field %q is outside the supported integer range", field)
		}
		return int(value), nil
	case json.Number:
		parsed, err := strconv.Atoi(value.String())
		if err != nil {
			return 0, fmt.Errorf("API response field %q must be an integer: %w", field, err)
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("API response field %q must be a number, got %T", field, value)
	}
}

func apiBool(result map[string]interface{}, field string) (bool, error) {
	if value, ok := result[field].(bool); ok {
		return value, nil
	}

	value, err := apiInt(result, field)
	if err != nil {
		return false, err
	}
	if value != 0 && value != 1 {
		return false, fmt.Errorf("API response field %q must be 0 or 1, got %d", field, value)
	}
	return value == 1, nil
}

func apiString(result map[string]interface{}, field string) (string, error) {
	value, ok := result[field]
	if !ok || value == nil {
		return "", fmt.Errorf("API response is missing field %q", field)
	}

	stringValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("API response field %q must be a string, got %T", field, value)
	}
	return stringValue, nil
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
