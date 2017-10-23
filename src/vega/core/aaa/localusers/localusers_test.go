// user_test
package localusers
/*
import (
	"os/user"
	"testing"
)

const (
	username     = "admin"
	username_bad = "vglasdf"

	username_test = "test"
	password_test = "123456"
)

func TestGetPrivilege(t *testing.T) {

	t.Log("[case] Test get privilege")
	previlige, err := GetPrivilege(username)
	if err == nil {
		t.Log("[info] Get privilege:", previlige)
	} else {
		t.Error("[err] Get privilege:", err)
	}

	t.Log("[case] Test get privilege(bad user)")
	previlige, err = GetPrivilege(username_bad)
	if err != nil {
		t.Log("[info] Get privilege:", err)
	} else {
		t.Error("[err] No error is reported for bad user", username_bad)
	}

	return
}

func TestGetPassword(t *testing.T) {

	t.Log("[case] Test get password")
	password, err := GetPassword(username)
	if err == nil {
		t.Log("[info] Get password:", password)
	} else {
		t.Error("[err] Get password:", err)
	}

	t.Log("[case] Test get password(bad user)")
	password, err = GetPassword(username_bad)
	if err != nil {
		t.Log("[info] Get password:", err)
	} else {
		t.Error("[err] No error is reported for bad user", username_bad)
	}

	return
}

func TestUser(t *testing.T) {

	t.Log("[case] Test add user")
	err := AddUser(username_test, password_test, PRIVILEGE_ADMIN)
	if err == nil {

		user_get, err := user.Lookup(username_test)
		if err == nil {
			t.Log("[info] Add user successful", user_get.Username, user_get.Uid)
		} else {
			t.Error("[err] Add user failed:", err)
		}

	} else {
		t.Error("[err] AddUser:", err)
	}

	t.Log("[case] Test modify user")
	err = SetPrivilege(username_test, PRIVILEGE_USER)
	if err == nil {
		privilege, err := GetPrivilege(username_test)
		if err == nil && privilege == PRIVILEGE_USER {
			t.Log("[info] Modify user successful", privilege)
		} else {
			t.Error("[err] Modify user failed:", privilege, err)
		}

	} else {
		t.Error("[err] ModifyUser:", err)
	}

	t.Log("[case] Test remove user")
	err = RemoveUser(username_test)
	if err == nil {
		user_get, err := user.Lookup(username_test)
		if user_get == nil {
			t.Log("[info] Remove user successful:", err)
		} else {
			t.Error("[err] Remove user failed, user still exists")
		}

	} else {
		t.Error("[err] RemoveUser:", err)
	}

	return
}

func TestConfig(t *testing.T) {

	var cfg Config

	t.Log("[case] Load then add")
	errs := cfg.Load()
	if len(errs) == 0 {
		t.Log("[info] Load successful")
		for _, user := range cfg.Users {
			t.Log("[info]", user)
		}
	} else {
		t.Error("[err] Load:", errs)
	}

	cfg.Users = append(cfg.Users, User{username_test, "$6$test", 2, 0})

	err := cfg.Verify()
	if err == nil {
		t.Log("[info] Verify successful")
	} else {
		t.Error("[err] Verify:", err)
	}

	err = cfg.Save()
	if err == nil {
		t.Log("[info] Save successful")
	} else {
		t.Error("[err] Save:", err)
	}

	t.Log("[case] Load then delete")
	errs = cfg.Load()
	if len(errs) == 0 {
		t.Log("[info] Load successful")
		for _, user := range cfg.Users {
			t.Log("[info]", user)
		}
	} else {
		t.Error("[err] Load:", errs)
	}

	for idx, user := range cfg.Users {
		if user.Username == username_test {
			cfg.Users = append(cfg.Users[:idx], cfg.Users[idx+1:]...)
		}
	}

	t.Log("[case] Verify")
	err = cfg.Verify()
	if err == nil {
		t.Log("[info] Verify successful")
	} else {
		t.Error("[err] Verify:", err)
	}

	err = cfg.Save()
	if err == nil {
		t.Log("[info] Save successful")
	} else {
		t.Error("[err] Save:", err)
	}

	return
}*/
