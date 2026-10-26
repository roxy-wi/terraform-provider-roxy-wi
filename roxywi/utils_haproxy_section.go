package roxywi

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func parseServersList(configList []interface{}) []map[string]interface{} {
	var configs []map[string]interface{}
	for _, config := range configList {
		configDetails := config.(map[string]interface{})
		configs = append(configs, map[string]interface{}{
			EthField:    configDetails[EthField].(string),
			IDField:     intFromInterface(configDetails[IDField]),
			MasterField: boolToInt(configDetails[MasterField].(bool)),
		})
	}
	return configs
}

func parseServersResult(config []map[string]interface{}) []interface{} {
	var configList []interface{}
	for _, c := range config {
		configList = append(configList, map[string]interface{}{
			EthField:    c[EthField].(string),
			IDField:     intFromInterface(c[IDField]),
			MasterField: c[MasterField],
		})
	}
	return configList
}
func parsePeersConfigList(configList []interface{}) []map[string]interface{} {
	var configs []map[string]interface{}
	for _, config := range configList {
		configDetails := config.(map[string]interface{})
		configs = append(configs, map[string]interface{}{
			IPField:       configDetails[IPField].(string),
			PeerNameField: configDetails[PeerNameField].(string),
			PortField:     intFromInterface(configDetails[PortField]),
		})
	}
	return configs
}

func parsePeersConfigListResult(config []map[string]interface{}) []interface{} {
	var configList []interface{}
	for _, c := range config {
		configList = append(configList, map[string]interface{}{
			IPField:       c[IPField].(string),
			PeerNameField: c[PeerNameField].(string),
			PortField:     intFromInterface(c[PortField]),
		})
	}
	return configList
}

func parseUserListConfigList(configList []interface{}) []map[string]interface{} {
	var configs []map[string]interface{}
	for _, config := range configList {
		configDetails := config.(map[string]interface{})
		configs = append(configs, map[string]interface{}{
			UserFiled:      configDetails[UserFiled].(string),
			PasswordField:  configDetails[PasswordField].(string),
			GroupNameField: configDetails[GroupNameField].(string),
		})
	}
	return configs
}

func parseUserBindsList(configList []interface{}) []map[string]interface{} {
	var configs []map[string]interface{}
	for _, config := range configList {
		configDetails := config.(map[string]interface{})
		configs = append(configs, map[string]interface{}{
			IPField:   configDetails[IPField].(string),
			PortField: configDetails[PortField].(int),
		})
	}
	return configs
}

func parseBackendsServerList(configList []interface{}) []map[string]interface{} {
	var configs []map[string]interface{}
	for _, config := range configList {
		configDetails := config.(map[string]interface{})
		configs = append(configs, map[string]interface{}{
			ServerTimeoutField:           configDetails[ServerTimeoutField].(string),
			BackendPortField:             configDetails[BackendPortField].(int),
			BackendServersPortCheckField: configDetails[BackendServersPortCheckField].(int),
			MaxconnFiled:                 configDetails[MaxconnFiled].(int),
			BackendServersSendProxyField: configDetails[BackendServersSendProxyField].(bool),
			BackendServersBackupField:    configDetails[BackendServersBackupField].(bool),
		})
	}
	return configs
}

func parseAclsList(configList []interface{}) []map[string]interface{} {
	var configs []map[string]interface{}
	for _, config := range configList {
		configDetails := config.(map[string]interface{})
		configs = append(configs, map[string]interface{}{
			AclIfField:        configDetails[AclIfField].(int),
			AclValueField:     configDetails[AclValueField].(string),
			AclThenField:      configDetails[AclThenField].(int),
			AclThenValueField: configDetails[AclThenValueField].(string),
		})
	}
	return configs
}

func parseHeaderList(configList []interface{}) []map[string]interface{} {
	var configs []map[string]interface{}
	for _, config := range configList {
		configDetails := config.(map[string]interface{})
		configs = append(configs, map[string]interface{}{
			PathField:       configDetails[PathField].(string),
			MethodField:     configDetails[MethodField].(string),
			HeaderNameField: configDetails[HeaderNameField].(string),
			ValueField:      configDetails[ValueField].(string),
		})
	}
	return configs
}

func parseBindsResult(config []map[string]interface{}) []interface{} {
	var configList []interface{}
	for _, c := range config {
		configList = append(configList, map[string]interface{}{
			IPField:   c[IPField].(string),
			PortField: c[PortField].(float64),
		})
	}
	return configList
}

func parseAclsServerResult(config []map[string]interface{}) []interface{} {
	var configList []interface{}
	for _, c := range config {
		configList = append(configList, map[string]interface{}{
			AclIfField:        c[AclIfField].(float64),
			AclValueField:     c[AclValueField].(string),
			AclThenField:      c[AclThenField].(float64),
			AclThenValueField: c[AclThenValueField].(string),
		})
	}
	if len(configList) == 0 {
		return nil
	}
	return configList
}

func parseHeadersResult(config []map[string]interface{}) []interface{} {
	var configList []interface{}
	for _, c := range config {
		configList = append(configList, map[string]interface{}{
			PathField:       c[PathField].(string),
			MethodField:     c[MethodField].(string),
			HeaderNameField: c[HeaderNameField].(string),
			ValueField:      c[ValueField].(string),
		})
	}
	if len(configList) == 0 {
		return nil
	}
	return configList
}

func parseBackendServerResult(config []map[string]interface{}) []interface{} {
	var configList []interface{}
	for _, c := range config {
		configList = append(configList, map[string]interface{}{
			ServerTimeoutField:           c[ServerTimeoutField].(string),
			BackendPortField:             c[BackendPortField].(float64),
			BackendServersPortCheckField: c[BackendServersPortCheckField].(float64),
			MaxconnFiled:                 c[MaxconnFiled].(float64),
			BackendServersSendProxyField: c[BackendServersSendProxyField].(bool),
			BackendServersBackupField:    c[BackendServersBackupField].(bool),
		})
	}
	if len(configList) == 0 {
		return nil
	}
	return configList
}

func parseUserListConfigListResult(config []map[string]interface{}) []interface{} {
	var configList []interface{}
	for _, c := range config {
		configList = append(configList, map[string]interface{}{
			UserFiled:      c[UserFiled].(string),
			PasswordField:  c[PasswordField].(string),
			GroupNameField: c[GroupNameField].(string),
		})
	}
	return configList
}

func hashMapStringInterface(v interface{}) int {
	m, ok := v.(map[string]interface{})
	if !ok {
		return 0
	}

	hash := 0
	for k, val := range m {
		hash += schema.HashString(k)
		hash += schema.HashString(fmt.Sprintf("%v", val))
	}
	return hash
}

func setTimeoutField(d *schema.ResourceData, fieldName string, value interface{}) error {
	if value == nil {
		items := make([]interface{}, 0)
		return d.Set(fieldName, schema.NewSet(hashMapStringInterface, items))
	}
	switch v := value.(type) {
	case map[string]interface{}:
		items := []interface{}{v}
		timeoutSet := schema.NewSet(hashMapStringInterface, items)
		return d.Set(fieldName, timeoutSet)
	default:
		return fmt.Errorf("unexpected type for %s: %T", fieldName, value)
	}
}

func getTimeoutMap(d *schema.ResourceData, fieldName string) (map[string]interface{}, error) {
	v := d.Get(fieldName)

	set, ok := v.(*schema.Set)
	if !ok || set.Len() == 0 {
		return nil, fmt.Errorf("field %s is not a valid Set", fieldName)
	}

	if timeoutMap, ok := set.List()[0].(map[string]interface{}); ok {
		return timeoutMap, nil
	}
	return nil, fmt.Errorf("unexpected type in the set for field %s", fieldName)
}

func getSetMap(d *schema.ResourceData, fieldName string) (map[string]interface{}, error) {
	v := d.Get(fieldName)

	set, ok := v.(*schema.Set)
	if !ok {
		return nil, fmt.Errorf("field %s is not a valid Set", fieldName)
	}
	if set.Len() == 0 {
		return nil, nil
	}

	if timeoutMap, ok := set.List()[0].(map[string]interface{}); ok {
		return timeoutMap, nil
	}
	return nil, fmt.Errorf("unexpected type in the set for field %s", fieldName)
}

func validateModeAndOptions(d *schema.ResourceDiff) error {
	modeInterface := d.Get(ModeField)
	mode, ok := modeInterface.(string)
	if !ok {
		return fmt.Errorf("field %s should be of type string", ModeField)
	}

	onlyWithHttpMode := []string{AntiBotField, CompressionField, CacheField, CookieField, SlowAttackField, SslOffloadingField, WafField}
	if mode == "tcp" {
		for _, field := range onlyWithHttpMode {
			fieldValueInterface := d.Get(field)
			fieldValue, ok := fieldValueInterface.(bool)
			if ok && fieldValue {
				return fmt.Errorf("field %s is not allowed in tcp mode", field)
			}
		}
	}
	return nil
}

func resourceSectionParseId(fullId string) (string, string, error) {
	parts := strings.Split(fullId, "-")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("expected ID in the format 'server_id-section_name', got: %s", fullId)
	}
	serverId := parts[0]
	sectionName := strings.Join(parts[1:], "-")
	return serverId, sectionName, nil

}

const haproxySectionIDPrefix = "haproxy-section-v2?"

func haproxyFilePathSchema(forceNew bool) *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     forceNew,
		ValidateFunc: validation.StringIsNotWhiteSpace,
		Description:  "Absolute HAProxy configuration file managed by this resource. Requires Roxy-WI 9.2 or later; omit it to use the main configuration file.",
	}
}

func haproxySectionURL(endpoint, filePath string) string {
	if filePath == "" {
		return endpoint
	}
	query := url.Values{}
	query.Set(FilePathField, filePath)
	return endpoint + "?" + query.Encode()
}

func haproxySectionFilePath(d *schema.ResourceData) string {
	filePath, _ := d.Get(FilePathField).(string)
	return filePath
}

func syncHaproxySectionFilePath(d *schema.ResourceData, filePath string) (string, error) {
	if filePath == "" {
		return haproxySectionFilePath(d), nil
	}
	if err := d.Set(FilePathField, filePath); err != nil {
		return "", fmt.Errorf("set file_path: %w", err)
	}
	return filePath, nil
}

func haproxySectionID(serverID, sectionName, filePath string) string {
	if filePath == "" {
		return serverID + "-" + sectionName
	}
	query := url.Values{}
	query.Set(ServerIdField, serverID)
	query.Set(NameField, sectionName)
	query.Set(FilePathField, filePath)
	return haproxySectionIDPrefix + query.Encode()
}

func resourceHaproxySectionParseID(fullID string) (string, string, string, error) {
	if strings.HasPrefix(fullID, haproxySectionIDPrefix) {
		query, err := url.ParseQuery(strings.TrimPrefix(fullID, haproxySectionIDPrefix))
		if err != nil {
			return "", "", "", fmt.Errorf("parse HAProxy section ID: %w", err)
		}
		serverID := query.Get(ServerIdField)
		sectionName := query.Get(NameField)
		filePath := query.Get(FilePathField)
		if serverID == "" || sectionName == "" || filePath == "" {
			return "", "", "", fmt.Errorf("HAProxy section ID is missing server_id, name, or file_path")
		}
		if _, err := strconv.Atoi(serverID); err != nil {
			return "", "", "", fmt.Errorf("HAProxy section ID has invalid server_id %q", serverID)
		}
		return serverID, sectionName, filePath, nil
	}

	if strings.Contains(fullID, "|") {
		parts := strings.SplitN(fullID, "|", 3)
		if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
			return "", "", "", fmt.Errorf("expected multi-file import ID in the format 'server_id|file_path|section_name', got: %s", fullID)
		}
		if _, err := strconv.Atoi(parts[0]); err != nil {
			return "", "", "", fmt.Errorf("HAProxy section ID has invalid server_id %q", parts[0])
		}
		return parts[0], parts[2], parts[1], nil
	}

	serverID, sectionName, err := resourceSectionParseId(fullID)
	return serverID, sectionName, "", err
}

func setHaproxySectionIDFromAPI(d *schema.ResourceData, apiID string) error {
	serverID, sectionName, err := resourceSectionParseId(apiID)
	if err != nil {
		return err
	}
	d.SetId(haproxySectionID(serverID, sectionName, haproxySectionFilePath(d)))
	return nil
}

func setSingletonHaproxySectionID(d *schema.ResourceData, sectionName string) error {
	serverID, ok := d.Get(ServerIdField).(int)
	if !ok || serverID < 1 {
		return fmt.Errorf("invalid server_id in Terraform state")
	}
	d.SetId(haproxySectionID(strconv.Itoa(serverID), sectionName, haproxySectionFilePath(d)))
	return nil
}

func haproxySectionImport(expectedSectionName string) schema.StateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
		serverID, sectionName, filePath, err := resourceHaproxySectionParseID(d.Id())
		if err != nil {
			return nil, err
		}
		if expectedSectionName != "" && sectionName != expectedSectionName {
			return nil, fmt.Errorf("expected %s section in import ID, got %q", expectedSectionName, sectionName)
		}
		serverIDValue, err := strconv.Atoi(serverID)
		if err != nil {
			return nil, fmt.Errorf("HAProxy section ID has invalid server_id %q", serverID)
		}
		if err := d.Set(ServerIdField, serverIDValue); err != nil {
			return nil, fmt.Errorf("set imported server_id: %w", err)
		}
		if expectedSectionName == "" {
			if err := d.Set(NameField, sectionName); err != nil {
				return nil, fmt.Errorf("set imported section name: %w", err)
			}
		}
		if filePath != "" {
			if err := d.Set(FilePathField, filePath); err != nil {
				return nil, fmt.Errorf("set imported file_path: %w", err)
			}
		}
		d.SetId(haproxySectionID(serverID, sectionName, filePath))
		return []*schema.ResourceData{d}, nil
	}
}
