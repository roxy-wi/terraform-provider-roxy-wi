package roxywi

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func udpListenerAPIResult() map[string]interface{} {
	return map[string]interface{}{
		ListenerIdField:       float64(9),
		CheckEnabledField:     float64(1),
		ClusterIdField:        float64(2),
		DelayBeforeRetryField: float64(3),
		DelayLoopField:        float64(4),
		DescriptionField:      "listener",
		RetryField:            float64(5),
		ServerIdField:         float64(6),
		VIPField:              "192.0.2.10",
		LbAlgorithmField:      "roundrobin",
		NameField:             "'edge'",
		PortField:             float64(53),
		GroupIdField:          float64(7),
		ConfigField:           "[{'backend_ip':'192.0.2.20','port':5353,'weight':10}]",
	}
}

func TestSetResourceDataFromUDPListenerResult(t *testing.T) {
	d := schema.TestResourceDataRaw(t, dataSourceUdpListener().Schema, nil)
	if err := setResourceDataFromResult(d, udpListenerAPIResult()); err != nil {
		t.Fatal(err)
	}

	if d.Id() != "9" || d.Get(NameField) != "edge" || d.Get(PortField) != 53 {
		t.Fatalf("unexpected listener state: id=%q name=%#v port=%#v", d.Id(), d.Get(NameField), d.Get(PortField))
	}
	config, ok := d.Get(ConfigField).(*schema.Set)
	if !ok || config.Len() != 1 {
		t.Fatalf("unexpected listener config: %#v", d.Get(ConfigField))
	}
	backend := config.List()[0].(map[string]interface{})
	if backend[BackendIPField] != "192.0.2.20" || backend[BackendPortField] != 5353 || backend[BackendWeightField] != 10 {
		t.Fatalf("unexpected backend state: %#v", backend)
	}
}

func TestSetResourceDataFromUDPListenerResultRejectsMalformedData(t *testing.T) {
	d := schema.TestResourceDataRaw(t, dataSourceUdpListener().Schema, nil)
	if err := setResourceDataFromResult(d, map[string]interface{}{}); err == nil {
		t.Fatal("expected a missing ID error")
	}

	result := udpListenerAPIResult()
	result[ConfigField] = "not-json"
	if err := setResourceDataFromResult(d, result); err == nil {
		t.Fatal("expected malformed config to fail")
	}
}

func TestDataSourceUDPListenerReadByIDAndName(t *testing.T) {
	result := udpListenerAPIResult()
	config := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/udp/listener/9":
			writeTestJSON(t, w, result)
		case "/api/udp/listeners":
			writeTestJSON(t, w, []map[string]interface{}{result})
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.Error(w, "unexpected request", http.StatusBadRequest)
		}
	})

	byID := schema.TestResourceDataRaw(t, dataSourceUdpListener().Schema, map[string]interface{}{ListenerIdField: "9"})
	if diagnostics := dataSourceUdpListenerRead(context.Background(), byID, config); diagnostics.HasError() {
		t.Fatalf("read by ID returned diagnostics: %v", diagnostics)
	}

	byName := schema.TestResourceDataRaw(t, dataSourceUdpListener().Schema, map[string]interface{}{NameField: "edge"})
	if diagnostics := dataSourceUdpListenerRead(context.Background(), byName, config); diagnostics.HasError() {
		t.Fatalf("read by name returned diagnostics: %v", diagnostics)
	}

	missing := schema.TestResourceDataRaw(t, dataSourceUdpListener().Schema, map[string]interface{}{NameField: "missing"})
	if diagnostics := dataSourceUdpListenerRead(context.Background(), missing, config); !diagnostics.HasError() {
		t.Fatal("expected missing listener name to fail")
	}
}
