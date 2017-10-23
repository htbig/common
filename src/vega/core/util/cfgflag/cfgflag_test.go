// cfgflag_test
package cfgflag

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

type VpnUser struct {
	Username string `json:"name"`
	Password string `json:"pass"`
}

type Config struct {
	Enabled     bool               `json:"enable"`
	Ipv4network string             `json:"ipv4-network"`
	Dns         string             `json:"dns"`
	Auth        string             `json:"auth"`
	User        map[string]VpnUser `json:"user"`
}

func TestCompareConfig(t *testing.T) {
	t.Log("[case] Test compare config")

	var user VpnUser
	user.Username = "user"
	user.Password = "hello"

	var user1 VpnUser
	user1.Username = "user1"
	user1.Password = "hello1"

	var user2 VpnUser
	user2.Username = "user2"
	user2.Password = "hello2"

	var user3 VpnUser
	user3.Username = "user3"
	user3.Password = "hello3"

	cfg1 := new(Config)
	cfg2 := new(Config)
	cfg1.User = make(map[string]VpnUser)
	cfg2.User = make(map[string]VpnUser)

	cfg1.Enabled = true
	cfg1.Ipv4network = "aaa"
	cfg1.Dns = "8.8.8.8"
	cfg1.Auth = "plain"

	cfg1.User[user1.Username] = user1
	cfg1.User[user3.Username] = user3

	cfg2.Enabled = true
	cfg2.Ipv4network = "aaa"
	cfg2.Dns = "8.8.8.4"
	cfg2.Auth = "plain"

	user.Username = "user1"
	user.Password = "hellopassword"

	cfg2.User[user.Username] = user
	cfg2.User[user2.Username] = user2

	var cfgflag ConfigFlag
	cfgflag.Init()
	cfgflag.UpdateFlag(cfg1, cfg2)

	assert.True(t, cfgflag.GetFlag("Enabled") == OP_NORMAL, "Enabled should be NORMAL")
	assert.True(t, cfgflag.GetFlag("Ipv4network") == OP_NORMAL, "Ipv3network should be NORMAL")
	assert.True(t, cfgflag.GetFlag("Dns") == OP_UPDATE, "Dns should be OPDATE")
	assert.True(t, cfgflag.GetFlag("Auth") == OP_NORMAL, "Auth should be NORMAL")

	assert.True(t, cfgflag.GetListFlag("User")["user1"] == OP_UPDATE, "User user1 should be UPDATE")
	assert.True(t, cfgflag.GetListFlag("User")["user2"] == OP_ADD, "User user2 should be ADD")
	assert.True(t, cfgflag.GetListFlag("User")["user3"] == OP_DEL, "User user3 should be DEL")

	return
}
