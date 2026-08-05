package roxywi

import "testing"

func TestFlattenHAServices(t *testing.T) {
	services, err := flattenHAServices(map[string]interface{}{
		"nginx":   map[string]interface{}{DockerField: float64(0), EnabledField: float64(1)},
		"haproxy": map[string]interface{}{DockerField: float64(1), EnabledField: float64(1)},
	})
	if err != nil {
		t.Fatalf("flattenHAServices: %v", err)
	}
	if len(services) != 2 {
		t.Fatalf("expected two services, got %d", len(services))
	}
	if services[0][NameField] != "haproxy" || services[1][NameField] != "nginx" {
		t.Fatalf("services are not sorted deterministically: %#v", services)
	}
	if services[0][DockerField] != true || services[1][DockerField] != false {
		t.Fatalf("service flags were not converted: %#v", services)
	}
}

func TestFlattenHAServicesRejectsUnexpectedShape(t *testing.T) {
	if _, err := flattenHAServices([]interface{}{}); err == nil {
		t.Fatal("expected unexpected services shape to fail")
	}
}
