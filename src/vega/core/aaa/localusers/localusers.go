// Copyright (c) 2016, Virtual Gateway Labs. All rights reserved.

// Package user provide APIs for showing/adding/deleting system users
package localusers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"vega/syslogger"
)

const (
	PRIVILEGE_USER  = 1
	PRIVILEGE_ADMIN = 2

	RADIUS_USER = "radius"

	user_file  = "/etc/passwd"
	pass_file  = "/etc/shadow"
	group_file = "/etc/group"

	home_dir   = "/tmp"
	shell_path = "/bin/vega-shell"

	default_user     = "admin"
	default_password = `$6$CZcv3EIg$7Q2f2vXhYgPupQb0oY58Xd90jk8vcZctKKa5IHY/FI1hlX1upHBYbNv2e1szC7/IbZDanw3LSewxI7buULTPP.`

	default_group = "users"
	radius_group  = "radius"
	admin_group   = "wheel"

	GID_ROOT  = 0
	GID_WHEEL = 10
	GID_SUDO  = 27
	GID_USER  = 100

	uid_local       = 1000
	uid_local_start = 1002

	PASSWD_LOCKED  = "!"
	PASSWD_SHADOW  = "x"
	PASSWD_NOLOGIN = "*"
	PASSWD_NIC     = "*NP*"
)

var (
	validUsernameRegexp = regexp.MustCompile("^[a-z]([a-z0-9]{0,31})$")
)

type Config []User

type (
	User struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		Privilege int    `json:"privilege"`
		uID       uint16
	}
)

func (config *Config) Legacy(legacyRoot string) {
	users, err := legacyGetUsers(legacyRoot)
	if err != nil {
		syslogger.Err("Legacy[aaa/localusers]: ", err)
	} else {
		*config = make([]User, len(users))
		copy(*config, users)
	}
}

func (config *Config) CopyFrom(otherConfig Config) {
	*config = make([]User, len(otherConfig))
	copy(*config, otherConfig)
}

func (config *Config) CopyFromInterface(data interface{}) bool {
	otherConfig, ok := data.(*Config)
	if !ok {
		return false
	}

	config.CopyFrom(*otherConfig)
	return true
}

func (config *Config) CloneInterface() interface{} {
	return config.Clone()
}

func (config *Config) Clone() *Config {
	newConfig := new(Config)
	newConfig.CopyFrom(*config)

	return newConfig
}

//func (cfg *Config) Load() (errs []error) {

//	var err error

//	if cfg == nil {
//		err = errors.New("FATAL: Bad user config instance")
//		errs = append(errs, err)
//		return
//	}

//	cfg.Users, err = GetUsers()
//	if err != nil {
//		errs = append(errs, err)
//		return
//	}

//	return
//}

func (config *Config) SaveInterface(data interface{}) (bool, []error) {
	oldConfig, ok := data.(*Config)
	if !ok {
		return false, nil
	}

	return true, config.Save(*oldConfig)
}

func (config *Config) Save(oldConfig Config) (errs []error) {
	for idx, user := range *config {
		if user.Privilege == 0 {
			(*config)[idx].Privilege = PRIVILEGE_USER
		}
	}

	has_admin := false
	for _, user := range *config {
		if user.Username == default_user {
			has_admin = true
			break
		}
	}

	if !has_admin {
		for _, user := range oldConfig {
			if user.Username == default_user {
				*config = append(*config, user)
				break
			}
		}
	}

	var err error

	users, err := parse_passwd("")
	if err != nil {
		errs = append(errs, err)
		return
	}

	for _, user := range users {
		if user.Username == RADIUS_USER || user.Username == default_user {
			continue
		}

		err = RemoveUser(user.Username)
		if err != nil {
			errs = append(errs, err)
			return
		}
	}

	err = add_to_passwd(*config)
	if err != nil {
		errs = append(errs, err)
		return
	}

	err = add_to_shadow(*config)
	if err != nil {
		errs = append(errs, err)
		return
	}

	err = add_to_group(*config)
	if err != nil {
		errs = append(errs, err)
		return
	}

	return
}

func (config *Config) Verify() (errs []error) {
	errs = VerifyUsers(*config)

	return errs
}

func (config *Config) Factory() {

	var admin User

	admin.Username = "admin"
	admin.Password = "$6$CZcv3EIg$7Q2f2vXhYgPupQb0oY58Xd90jk8vcZctKKa5IHY/FI1hlX1upHBYbNv2e1szC7/IbZDanw3LSewxI7buULTPP."
	admin.Privilege = 2
	admin.uID = uid_local
	*config = []User{admin}

	return
}

// Get users from the config
func (config *Config) LoadUsers() (err error) {

	users, err := GetUsers()
	if err != nil {
		return
	}

	*config = users

	return
}

func validUsername(username string) bool {
	return validUsernameRegexp.MatchString(username)
}

func VerifyUsers(users []User) (errs []error) {
	for _, user := range users {
		if user.Username == "" {
			err := errors.New("Empty username")
			errs = append(errs, err)
		} else if user.Username == default_user {
			if user.Privilege != PRIVILEGE_ADMIN {
				err := fmt.Errorf("Can not change reserved user's privilege level: %s", user.Username)
				errs = append(errs, err)
			}
		} else if user.Username == RADIUS_USER {
			err := fmt.Errorf("Can not add reserved user: %s", user.Username)
			errs = append(errs, err)
		}

		if !validUsername(user.Username) {
			err := fmt.Errorf("Invalid username: %s", user.Username)
			errs = append(errs, err)
		}

		if user.Password == "" {
			err := fmt.Errorf("Empty user password of the user: %s", user.Username)
			errs = append(errs, err)
		}

		if user.Privilege == 0 {
			err := fmt.Errorf("User privilege level not specified for user: %s", user.Username)
			errs = append(errs, err)
		} else if user.Privilege != PRIVILEGE_USER && user.Privilege != PRIVILEGE_ADMIN {
			err := fmt.Errorf("Bad user privilege level %d of user: %s", user.Privilege, user.Username)
			errs = append(errs, err)
		}
	}

	return errs
}

// Get groups of user
func getGroups(rootPath, username string) ([]string, error) {
	confPath := rootPath + group_file
	fileData, err := ioutil.ReadFile(confPath)
	if err != nil {
		return nil, err
	}

	groups := []string{}
	lines := strings.Split(strings.TrimSpace(string(fileData)), "\n")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 4 {
			continue
		}
		group, usernames := fields[0], fields[3]
		for _, name := range strings.Split(usernames, ",") {
			if username == name {
				groups = append(groups, group)
				break
			}
		}
	}

	return groups, nil
}

// legacy GetPrivilege
func legacyGetPrivilege(rootPath, username string) (int, error) {
	groups, err := getGroups(rootPath, username)
	if err != nil {
		return PRIVILEGE_USER, err
	}

	for _, group := range groups {
		if group == `wheel` {
			return PRIVILEGE_ADMIN, nil
		}
	}

	return PRIVILEGE_USER, nil
}

// legacy GetUsers
func legacyGetUsers(rootPath string) ([]User, error) {
	users, err := parse_passwd(rootPath)
	if err != nil {
		return nil, err
	}

	for idx, user := range users {
		users[idx].Privilege, err = legacyGetPrivilege(rootPath, user.Username)
		if err != nil {
			return nil, err
		}
	}

	return users, nil
}

// Get users from the system
func GetUsers() (users []User, err error) {

	users, err = parse_passwd("")
	if err != nil {
		return
	}

	for idx, user := range users {
		users[idx].Privilege, err = GetPrivilege(user.Username)
		if err != nil {
			return
		}
	}

	return
}

func AddUser(username string, password string, level int) (err error) {
	groups := []string{"users"}

	switch level {
	case PRIVILEGE_USER:
	case PRIVILEGE_ADMIN:
		groups = append(groups, admin_group)
	default:
		err = errors.New("Bad privilege level")
	}

	group_line := strings.Join(groups, ",")

	cmd := exec.Command("useradd", "-M", "-g", default_group, "-G", group_line,
		"-d", "/tmp", "-s", shell_path, username)

	output, err := cmd.CombinedOutput()
	if err != nil {
		err = errors.New(fmt.Sprint(err, "\n", string(output)))
		return
	}

	err = SetPassword(username, password)
	if err != nil {
		return
	}

	return
}

func RemoveUser(username string) (err error) {
	exist, err := check_exist(username)
	if err != nil {
		return
	}

	if !exist {
		err = errors.New("User not found")
	} else if username == default_user {
		err = fmt.Errorf("Cannot delete the default user: %s", default_user)
	} else if username == RADIUS_USER {
		err = errors.New("Cannot delete the radius user")
	} else if in_use(username) {
		err = fmt.Errorf("Cannot delete because user %s is logged in", username)
	} else {
		output, err := exec.Command("userdel", "-f", username).CombinedOutput()
		if err != nil {
			err = errors.New(fmt.Sprint(err, "\n", string(output)))
			return err
		}
	}

	return
}

// Change user's privilege level.
func SetPrivilege(username string, level int) (err error) {
	exist, err := check_exist(username)
	if err != nil {
		return
	}

	var privilege int

	if !exist {
		err = errors.New("User not found")
	} else if username == default_user {
		err = errors.New(fmt.Sprint("Cannot modify the default user:", default_user))
	} else if username == RADIUS_USER {
		err = errors.New("Cannot modify the radius user")
	} else {
		privilege, err = GetPrivilege(username)
		if err != nil {
			return
		}

		if privilege == level {
			return
		}
	}

	// change group that user belongs to
	switch level {
	case PRIVILEGE_USER:
		if privilege == PRIVILEGE_ADMIN {
			output, errexec := exec.Command("gpasswd", "-d", username, "wheel").CombinedOutput()
			if errexec != nil {
				err = errors.New(fmt.Sprint(errexec, "\n", string(output)))
				return
			}
		}
	case PRIVILEGE_ADMIN:
		if privilege == PRIVILEGE_USER {
			output, errexec := exec.Command("gpasswd", "-a", username, "wheel").CombinedOutput()
			if errexec != nil {
				err = errors.New(fmt.Sprint(errexec, "\n", string(output)))
				return
			}
		}
	default:
		err = errors.New("Bad privilege level")
	}

	return
}

func GetPrivilege(username string) (level int, err error) {
	user, err := user.Lookup(username)
	if err != nil {
		return
	}

	gid, err := strconv.Atoi(user.Gid)
	if err != nil {
		return
	}

	if gid != GID_USER {
		err = fmt.Errorf("Internal error: User %s is not in 'users' group", user)
	}

	output, err := exec.Command("id", username).CombinedOutput()
	if err != nil {
		err = errors.New(fmt.Sprint(err, "\n", string(output)))
		return
	}

	fields := strings.Fields(string(output))
	for _, field := range fields {
		if strings.Contains(field, "groups=") {
			if strings.Contains(field, "10(wheel)") {
				level = PRIVILEGE_ADMIN
			} else {
				level = PRIVILEGE_USER
			}
		}
	}

	if level == 0 {
		err = errors.New("Failed to get privilege level")
	}

	return
}

func SetPassword(username string, password string) (err error) {
	pair := username + ":" + password
	reader := strings.NewReader(pair)

	cmd := exec.Command("chpasswd")
	cmd.Stdin = reader

	output, err := cmd.CombinedOutput()
	if err != nil {
		err = errors.New(fmt.Sprint(err, "\n", string(output)))
		return
	}

	return
}

// Get encrypted password of a user from '/etc/shadow'
func GetPassword(rootPath, username string) (password string, err error) {
	if username == "radius" {
		err = errors.New("Cannot get radius user password")
		return
	}

	data, err := ioutil.ReadFile(rootPath + pass_file)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")

	var found bool

	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}

		if username == fields[0] {
			password = fields[1]
			found = true
			break
		}
	}

	if !found {
		err = errors.New("No User found")
	}

	return
}

func in_use(username string) bool {
	err := exec.Command("pstree", username).Run()
	if err != nil {
		return false
	} else {
		return true
	}
}

func check_exist(username string) (exist bool, err error) {
	users, err := parse_passwd("")
	if err != nil {
		return
	}

	for _, user := range users {
		if username == user.Username {
			exist = true
			break
		}
	}

	return
}

// passwd format
// username:password indicator:uid:gid:gecos:home dir:shell
// example
// vgl:x:1000:100:VG-Labs:/home/vgl:/bin/bash

// Get all created user(UID >= 1000) from '/etc/passwd'
func parse_passwd(rootPath string) (users []User, err error) {
	data, err := ioutil.ReadFile(rootPath + user_file)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Split(line, ":")
		if len(fields) < 3 {
			continue
		}

		uid, _ := strconv.Atoi(fields[2])

		var user User

		// exclude reserved users
		if fields[0] == RADIUS_USER || uid < uid_local {
			continue
		}

		user.Username = fields[0]

		switch fields[1] {
		case PASSWD_LOCKED:
		case PASSWD_NOLOGIN:
			continue
		case PASSWD_NIC:
			continue
		case PASSWD_SHADOW:
			user.Password, err = GetPassword(rootPath, user.Username)
		default:
		}

		user.uID = uint16(uid)

		users = append(users, user)
	}

	return
}

func add_to_passwd(users []User) (err error) {

	if len(users) == 0 {
		err = errors.New("No users inputed")
	}

	data, err := ioutil.ReadFile(user_file)
	if err != nil {
		return
	}

	var found bool

	lines := strings.Split(string(data), "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	for idx, user := range users {
		if user.Username == default_user || user.Username == "" {
			continue
		} else if user.Username == RADIUS_USER {
			err = errors.New("Could not modify reserved user information")
			return
		}

		// check if there is already a user
		for _, line := range lines {
			fields := strings.Split(line, ":")
			if len(fields) < 3 {
				continue
			}

			if user.Username == fields[0] {
				uid, _ := strconv.Atoi(fields[2])
				// exclude reserved users
				if uid < uid_local {
					err = errors.New("Could not modify reserved user information")
					return
				} else {
					found = true
				}
			}
		}

		if !found {
			new_line := user.Username + ":" + PASSWD_SHADOW + ":" + strconv.Itoa(uid_local_start+idx) + ":" +
				strconv.Itoa(GID_USER) + "::" + home_dir + ":" + shell_path
			lines = append(lines, new_line)
		} else {
			found = false
		}
	}

	err = ioutil.WriteFile(user_file, []byte(strings.Join(lines, "\n")), 600)
	if err != nil {
		return
	}

	return
}

// add password to /etc/shadow
func add_to_shadow(users []User) (err error) {

	if len(users) == 0 {
		err = errors.New("No users inputed")
	}

	data, err := ioutil.ReadFile(pass_file)
	if err != nil {
		return
	}

	var found bool

	lines := strings.Split(string(data), "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	for _, user := range users {
		if user.Username == RADIUS_USER {
			err = errors.New("Could not modify reserved user password")
			return
		}

		// check if there is already a user
		for idx, line := range lines {
			fields := strings.Split(line, ":")
			if len(fields) < 3 {
				continue
			}

			if user.Username == fields[0] {
				fields[1] = user.Password
				found = true
				lines[idx] = strings.Join(fields, ":")
			}
		}

		if !found {
			new_line := user.Username + ":" + user.Password + ":99999:0:99999:7:::"
			lines = append(lines, new_line)
		} else {
			found = false
		}

	}

	err = ioutil.WriteFile(pass_file, []byte(strings.Join(lines, "\n")), 600)
	if err != nil {
		return
	}

	return
}

// group format
// users:x:100:vgl,radius

// add privilege to /etc/group
func add_to_group(users []User) (err error) {

	if len(users) == 0 {
		err = errors.New("No users inputed")
	}

	data, err := ioutil.ReadFile(group_file)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	var users_normal []string
	var users_admin []string
	var admin_with_root []string

	users_admin = append(users_admin, default_user)

	for _, user := range users {
		if user.Username == default_user || user.Username == RADIUS_USER {
			continue
		}

		// reserved user, should be always in the wheel group

		if user.Privilege == PRIVILEGE_ADMIN {
			users_admin = append(users_admin, user.Username)
		}
	}

	// check if there is already a user
	for idx, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 3 {
			continue
		}

		if fields[0] == default_group {
			users_normal = append(users_normal, users_admin...)
			fields[3] = strings.Join(users_normal, ",")
			lines[idx] = strings.Join(fields, ":")
		} else if fields[0] == admin_group {
			admin_with_root = append(admin_with_root, "root")
			admin_with_root = append(admin_with_root, users_admin...)
			fields[3] = strings.Join(admin_with_root, ",")
			lines[idx] = strings.Join(fields, ":")
		}
	}

	err = ioutil.WriteFile(group_file, []byte(strings.Join(lines, "\n")), 600)
	if err != nil {
		return
	}

	return
}
