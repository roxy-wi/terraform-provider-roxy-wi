package roxywi

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAPIInt(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    int
		wantErr bool
	}{
		{name: "JSON number", value: float64(42), want: 42},
		{name: "native integer", value: 7, want: 7},
		{name: "number token", value: json.Number("9"), want: 9},
		{name: "fraction", value: 1.5, wantErr: true},
		{name: "string", value: "42", wantErr: true},
		{name: "nil", value: nil, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := apiInt(map[string]interface{}{"field": test.value}, "field")
			if test.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got %d", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("got %d, want %d", got, test.want)
			}
		})
	}
}

func TestAPIBool(t *testing.T) {
	for _, test := range []struct {
		value   interface{}
		want    bool
		wantErr bool
	}{
		{value: true, want: true},
		{value: float64(1), want: true},
		{value: float64(0), want: false},
		{value: float64(2), wantErr: true},
		{value: "true", wantErr: true},
	} {
		got, err := apiBool(map[string]interface{}{"field": test.value}, "field")
		if test.wantErr != (err != nil) {
			t.Fatalf("apiBool(%#v) error = %v", test.value, err)
		}
		if err == nil && got != test.want {
			t.Fatalf("apiBool(%#v) = %v, want %v", test.value, got, test.want)
		}
	}
}

func TestAPIString(t *testing.T) {
	if got, err := apiString(map[string]interface{}{"field": "value"}, "field"); err != nil || got != "value" {
		t.Fatalf("apiString returned %q, %v", got, err)
	}
	if _, err := apiString(map[string]interface{}{"field": 1}, "field"); err == nil {
		t.Fatal("expected a type error")
	}
	if _, err := apiString(map[string]interface{}{}, "field"); err == nil {
		t.Fatal("expected a missing-field error")
	}
}

func TestParseConfigRejectsInvalidListItem(t *testing.T) {
	_, err := parseConfig([]interface{}{map[string]interface{}{}, "invalid"})
	if err == nil || !strings.Contains(err.Error(), "index 1") {
		t.Fatalf("expected an indexed type error, got %v", err)
	}
}

func TestGetSetMapRejectsNonSetWithoutPanic(t *testing.T) {
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"field": {Type: schema.TypeString, Optional: true},
	}, map[string]interface{}{"field": "value"})

	if _, err := getSetMap(d, "field"); err == nil {
		t.Fatal("expected a non-Set value to fail")
	}
}

func TestConvertToIntSupportsIntegerWidths(t *testing.T) {
	for _, test := range []struct {
		value interface{}
		want  int
	}{
		{value: float64(1), want: 1},
		{value: 2, want: 2},
		{value: int32(3), want: 3},
		{value: int64(4), want: 4},
	} {
		if got := convertToInt(test.value); got != test.want {
			t.Fatalf("convertToInt(%T(%v)) = %d, want %d", test.value, test.value, got, test.want)
		}
	}
}
