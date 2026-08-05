package roxywi

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestUDPListenerConfigParsers(t *testing.T) {
	raw := []interface{}{map[string]interface{}{
		BackendIPField:     "192.0.2.1",
		BackendPortField:   53,
		BackendWeightField: 10,
	}}
	parsed := parseConfigList(raw)
	if len(parsed) != 1 || parsed[0][BackendPortField] != 53 {
		t.Fatalf("unexpected parsed config: %#v", parsed)
	}
	flattened := parseConfigResult([]map[string]interface{}{{
		BackendIPField:     "192.0.2.1",
		BackendPortField:   float64(53),
		BackendWeightField: int64(10),
	}})
	if len(flattened) != 1 || flattened[0].(map[string]interface{})[BackendWeightField] != 10 {
		t.Fatalf("unexpected flattened config: %#v", flattened)
	}

	for _, test := range []struct {
		value interface{}
		want  int
	}{
		{float64(1), 1},
		{2, 2},
		{int32(3), 3},
		{int64(4), 4},
		{"invalid", 0},
	} {
		if got := intFromInterface(test.value); got != test.want {
			t.Fatalf("intFromInterface(%T(%v)) = %d, want %d", test.value, test.value, got, test.want)
		}
	}

	if parsed, err := parseConfig(nil); err != nil || parsed != nil {
		t.Fatalf("parseConfig(nil) = %#v, %v", parsed, err)
	}
	if parsed, err := parseConfig("[{'backend_ip':'192.0.2.1','port':53,'weight':10}]"); err != nil || len(parsed) != 1 {
		t.Fatalf("parseConfig(string) = %#v, %v", parsed, err)
	}
	if _, err := parseConfig(42); err == nil {
		t.Fatal("expected unsupported config type to fail")
	}
}

func TestHAProxySectionParsers(t *testing.T) {
	servers := parseServersList([]interface{}{map[string]interface{}{
		EthField: "eth0", IDField: 7, MasterField: true,
	}})
	if len(servers) != 1 || servers[0][MasterField] != 1 {
		t.Fatalf("unexpected servers payload: %#v", servers)
	}
	serverState := parseServersResult([]map[string]interface{}{{
		EthField: "eth0", IDField: float64(7), MasterField: float64(1),
	}})
	if len(serverState) != 1 || serverState[0].(map[string]interface{})[IDField] != 7 {
		t.Fatalf("unexpected servers state: %#v", serverState)
	}

	peers := parsePeersConfigList([]interface{}{map[string]interface{}{
		IPField: "192.0.2.2", PeerNameField: "peer", PortField: 1024,
	}})
	if len(peers) != 1 || peers[0][PortField] != 1024 {
		t.Fatalf("unexpected peers payload: %#v", peers)
	}
	peerState := parsePeersConfigListResult([]map[string]interface{}{{
		IPField: "192.0.2.2", PeerNameField: "peer", PortField: float64(1024),
	}})
	if len(peerState) != 1 || peerState[0].(map[string]interface{})[PortField] != 1024 {
		t.Fatalf("unexpected peers state: %#v", peerState)
	}

	users := parseUserListConfigList([]interface{}{map[string]interface{}{
		UserFiled: "alice", PasswordField: "secret", GroupNameField: "admins",
	}})
	if len(users) != 1 || users[0][UserFiled] != "alice" {
		t.Fatalf("unexpected userlist payload: %#v", users)
	}
	userState := parseUserListConfigListResult([]map[string]interface{}{{
		UserFiled: "alice", PasswordField: "secret", GroupNameField: "admins",
	}})
	if len(userState) != 1 || userState[0].(map[string]interface{})[GroupNameField] != "admins" {
		t.Fatalf("unexpected userlist state: %#v", userState)
	}

	binds := parseUserBindsList([]interface{}{map[string]interface{}{IPField: "0.0.0.0", PortField: 443}})
	if len(binds) != 1 || binds[0][PortField] != 443 {
		t.Fatalf("unexpected binds payload: %#v", binds)
	}
	bindState := parseBindsResult([]map[string]interface{}{{IPField: "0.0.0.0", PortField: float64(443)}})
	if len(bindState) != 1 || bindState[0].(map[string]interface{})[PortField] != float64(443) {
		t.Fatalf("unexpected binds state: %#v", bindState)
	}

	backend := map[string]interface{}{
		ServerTimeoutField:           "192.0.2.3",
		BackendPortField:             8080,
		BackendServersPortCheckField: 8081,
		MaxconnFiled:                 100,
		BackendServersSendProxyField: true,
		BackendServersBackupField:    false,
	}
	backendPayload := parseBackendsServerList([]interface{}{backend})
	if len(backendPayload) != 1 || backendPayload[0][BackendPortField] != 8080 {
		t.Fatalf("unexpected backend payload: %#v", backendPayload)
	}
	backendState := parseBackendServerResult([]map[string]interface{}{{
		ServerTimeoutField:           "192.0.2.3",
		BackendPortField:             float64(8080),
		BackendServersPortCheckField: float64(8081),
		MaxconnFiled:                 float64(100),
		BackendServersSendProxyField: true,
		BackendServersBackupField:    false,
	}})
	if len(backendState) != 1 {
		t.Fatalf("unexpected backend state: %#v", backendState)
	}
	if parseBackendServerResult(nil) != nil {
		t.Fatal("expected empty backend state to be nil")
	}

	aclPayload := parseAclsList([]interface{}{map[string]interface{}{
		AclIfField: 1, AclValueField: "example.com", AclThenField: 5, AclThenValueField: "api",
	}})
	if len(aclPayload) != 1 || aclPayload[0][AclThenField] != 5 {
		t.Fatalf("unexpected ACL payload: %#v", aclPayload)
	}
	aclState := parseAclsServerResult([]map[string]interface{}{{
		AclIfField: float64(1), AclValueField: "example.com", AclThenField: float64(5), AclThenValueField: "api",
	}})
	if len(aclState) != 1 || parseAclsServerResult(nil) != nil {
		t.Fatalf("unexpected ACL state: %#v", aclState)
	}

	headerPayload := parseHeaderList([]interface{}{map[string]interface{}{
		PathField: "http-request", MethodField: "set-header", HeaderNameField: "X-Test", ValueField: "yes",
	}})
	if len(headerPayload) != 1 || headerPayload[0][HeaderNameField] != "X-Test" {
		t.Fatalf("unexpected header payload: %#v", headerPayload)
	}
	headerState := parseHeadersResult([]map[string]interface{}{{
		PathField: "http-request", MethodField: "set-header", HeaderNameField: "X-Test", ValueField: "yes",
	}})
	if len(headerState) != 1 || parseHeadersResult(nil) != nil {
		t.Fatalf("unexpected header state: %#v", headerState)
	}
}

func TestNginxSectionParsers(t *testing.T) {
	payload := parseNginxBackendsServerList([]interface{}{map[string]interface{}{
		ServerTimeoutField: "192.0.2.4",
		BackendPortField:   8080,
		MaxFails:           3,
		FailTimeout:        30,
	}})
	if len(payload) != 1 || payload[0][MaxFails] != 3 {
		t.Fatalf("unexpected nginx payload: %#v", payload)
	}
	state := parseNginxBackendServerResult([]map[string]interface{}{{
		ServerTimeoutField: "192.0.2.4",
		BackendPortField:   float64(8080),
		MaxFails:           float64(3),
		FailTimeout:        float64(30),
	}})
	if len(state) != 1 || parseNginxBackendServerResult(nil) != nil {
		t.Fatalf("unexpected nginx state: %#v", state)
	}
}

func TestTimeoutSetHelpers(t *testing.T) {
	resourceSchema := map[string]*schema.Schema{
		"timeout": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"value": {Type: schema.TypeString, Optional: true},
			}},
		},
	}
	d := schema.TestResourceDataRaw(t, resourceSchema, nil)
	if err := setTimeoutField(d, "timeout", map[string]interface{}{"value": "5s"}); err != nil {
		t.Fatal(err)
	}
	timeout, err := getTimeoutMap(d, "timeout")
	if err != nil || timeout["value"] != "5s" {
		t.Fatalf("getTimeoutMap returned %#v, %v", timeout, err)
	}
	setValue, err := getSetMap(d, "timeout")
	if err != nil || setValue["value"] != "5s" {
		t.Fatalf("getSetMap returned %#v, %v", setValue, err)
	}
	if err := setTimeoutField(d, "timeout", "invalid"); err == nil {
		t.Fatal("expected invalid timeout value to fail")
	}
	if hashMapStringInterface("invalid") != 0 || hashMapStringInterface(map[string]interface{}{"value": "5s"}) == 0 {
		t.Fatal("unexpected timeout hash result")
	}
}
