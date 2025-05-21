package roxywi

func parseNginxBackendsServerList(configList []interface{}) []map[string]interface{} {
	var configs []map[string]interface{}
	for _, config := range configList {
		configDetails := config.(map[string]interface{})
		configs = append(configs, map[string]interface{}{
			ServerTimeoutField: configDetails[ServerTimeoutField].(string),
			BackendPortField:   configDetails[BackendPortField].(int),
			MaxFails:           configDetails[MaxFails].(int),
			FailTimeout:        configDetails[FailTimeout].(int),
		})
	}
	return configs
}

func parseNginxBackendServerResult(config []map[string]interface{}) []interface{} {
	var configList []interface{}
	for _, c := range config {
		configList = append(configList, map[string]interface{}{
			ServerTimeoutField: c[ServerTimeoutField].(string),
			BackendPortField:   c[BackendPortField].(float64),
			MaxFails:           c[MaxFails].(float64),
			FailTimeout:        c[FailTimeout].(float64),
		})
	}
	if len(configList) == 0 {
		return nil
	}
	return configList
}
