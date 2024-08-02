package roxywi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
	"net/mail"
)

const (
	UserEmailField    = "email"
	UserEnabledField  = "enabled"
	UserPasswordField = "password"
	UserUsernameField = "username"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Schema: map[string]*schema.Schema{
			UserEmailField: {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The email of the user.",
				ValidateFunc: validateEmail,
			},
			UserEnabledField: {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Whether the user is enabled (true for enabled, false for disabled).",
			},
			UserPasswordField: {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The password of the user.",
			},
			UserUsernameField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The username of the user.",
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	user := map[string]interface{}{
		UserEmailField:    d.Get(UserEmailField).(string),
		UserEnabledField:  boolToInt(d.Get(UserEnabledField).(bool)),
		UserPasswordField: d.Get(UserPasswordField).(string),
		UserUsernameField: d.Get(UserUsernameField).(string),
	}

	resp, err := client.doRequest("POST", "/api/user", user)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	id, ok := result["id"]
	if !ok {
		return diag.Errorf("unable to extract user ID from response: %v", result)
	}

	// Handle both string and numeric ID
	switch v := id.(type) {
	case string:
		d.SetId(v)
	case float64:
		d.SetId(fmt.Sprintf("%.0f", v))
	default:
		return diag.Errorf("unexpected type for user ID: %T", id)
	}

	return resourceUserRead(ctx, d, m)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	resp, err := client.doRequest("GET", fmt.Sprintf("/api/user/%s", d.Id()), nil)
	if err != nil {
		if isNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(UserEmailField, result[UserEmailField]); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(UserEnabledField, intToBool(result[UserEnabledField].(float64))); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(UserUsernameField, result[UserUsernameField]); err != nil {
		return diag.FromErr(err)
	}
	// Note: Password is not set here for security reasons

	return nil
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	user := map[string]interface{}{
		UserEmailField:    d.Get(UserEmailField).(string),
		UserEnabledField:  boolToInt(d.Get(UserEnabledField).(bool)),
		UserPasswordField: d.Get(UserPasswordField).(string),
		UserUsernameField: d.Get(UserUsernameField).(string),
	}

	_, err := client.doRequest("PUT", fmt.Sprintf("/api/user/%s", d.Id()), user)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceUserRead(ctx, d, m)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	_, err := client.doRequest("DELETE", fmt.Sprintf("/api/user/%s", d.Id()), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

// Utility function to check if the error is a 404 not found error
func isNotFound(err error) bool {
	if httpErr, ok := err.(*httpError); ok {
		return httpErr.StatusCode == http.StatusNotFound
	}
	return false
}

// Define the HTTPError struct and methods
type httpError struct {
	StatusCode int
	Err        error
}

func (e *httpError) Error() string {
	return e.Err.Error()
}

// Utility function to convert bool to int
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Utility function to convert int to bool
func intToBool(i float64) bool {
	return i == 1
}

// Utility function to validate email format
func validateEmail(val interface{}, key string) (warns []string, errs []error) {
	_, err := mail.ParseAddress(val.(string))
	if err != nil {
		errs = append(errs, fmt.Errorf("%q must be a valid email address: %v", key, err))
	}
	return
}
