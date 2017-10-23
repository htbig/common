package radius
/*
import (
	"testing"
)

const (
	test_server_ip    = "192.168.165.30"
	test_server_ip_v6 = "fe80::e54:a5ff:fe57:9641"
	test_secret       = "testing123"

	test_username = "test"
	test_password = "pass"

	test_username_bad = "@#@&$()^&%%$"
)

func TestServerList(t *testing.T) {

	t.Log("[case] Test write radius server config")
	servers, err := read_server_list()
	if err == nil {
		for idx, server := range servers {
			t.Log("[info] server:", idx, server.IPaddr, server.Secret, server.Port)
		}
	} else {
		t.Error("[err] write radius server list:", err)
	}

	t.Log("[case] Test write radius server list")
	err = write_server_list(servers)
	if err == nil {
		t.Log("[info] Write radius server list successful")
	} else {
		t.Error("[err] Write radius server list:", err)
	}

	err = AddServer(test_server_ip_v6, test_secret, authserver_port)
	if err == nil {
		t.Log("[info] Add radius server successful")
	} else {
		t.Error("[err] Add radius server:", err)
	}
}

func TestAuth(t *testing.T) {

	t.Log("[case] Test auth")

	t.Log("[info] add server")
	err := AddServer(test_server_ip, test_secret, authserver_port)
	if err != nil {
		t.Error("[err] Add server:", err)
	}

	privilege, _, err := RadiusAuthenticate(test_username, test_password)
	if err == nil {
		t.Log("[info] Get radius user privilege =", privilege)
	} else {
		t.Error("[err] Test auth:", err)
	}

	t.Log("[case] Test auth(bad username)")
	privilege, ok, err := RadiusAuthenticate(test_username_bad, test_password)
	if err != nil {
		t.Log("[info] Test auth:", err)
	} else {
		t.Error("[err] No error is reported", ok)
	}
}
*/
