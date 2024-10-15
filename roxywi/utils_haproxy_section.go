package roxywi

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
