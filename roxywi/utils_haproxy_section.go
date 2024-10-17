package roxywi

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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
	if len(configs) == 0 {
		return nil
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
	if len(configs) == 0 {
		return nil
	}
	return configs
}

func parseBindsResult(config []map[string]interface{}) []interface{} {
	var configList []interface{}
	for _, c := range config {
		configList = append(configList, map[string]interface{}{
			IPField:   c[IPField].(string),
			PortField: c[PortField].(int),
		})
	}
	return configList
}

func parseBackendServerResult(config []map[string]interface{}) []interface{} {
	var configList []interface{}
	for _, c := range config {
		configList = append(configList, map[string]interface{}{
			ServerTimeoutField:           c[ServerTimeoutField].(string),
			BackendPortField:             c[BackendPortField].(int),
			BackendServersPortCheckField: c[BackendServersPortCheckField].(int),
			MaxconnFiled:                 c[MaxconnFiled].(int),
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
	if set.Len() == 0 {
		return nil, nil
	}
	if !ok {
		return nil, fmt.Errorf("field %s is not a valid Set", fieldName)
	}

	if timeoutMap, ok := set.List()[0].(map[string]interface{}); ok {
		return timeoutMap, nil
	}
	return nil, fmt.Errorf("unexpected type in the set for field %s", fieldName)
}
