// Copyright (c) 2016, Virtual Gateway Labs. All rights reserved.

// Package radius provide APIs for getting or setting the radius authentication
package radius

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"vega/core/util"
	"vega/core/util/cfg"
	"vega/syslogger"
)

const (
	radius_server     = "/etc/raddb/server"
	pam_radius        = "/etc/pam.d/pam_radius"
	nss_conf          = "/etc/nsswitch.conf"
	pam_radius_legacy = "/etc/pam.d/vbg_radius"

	pam_option_required   = "auth required  pam_radius_auth.so debug"
	pam_option_sufficient = "auth sufficient  pam_radius_auth.so debug"

	nss_key = "passwd:"

	authserver_port = 1812
)

type (
	Config struct {
		Fallback bool     `json:"fallback"`
		Enabled  bool     `json:"enable"`
		Servers  []Server `json:"servers"`
	}

	Server struct {
		IPaddr string `json:"ip"`
		Secret string `json:"secret"`
		Port   uint16 `json:"port"`
	}
)

func (cfg *Config) Legacy(legacyRoot string) {
	const errString string = "Legacy[aaa/radius]:"

	if servers, err := read_server_list(legacyRoot); err != nil {
		syslogger.Err(errString, err)
	} else {
		cfg.Servers = make([]Server, len(servers))
		copy(cfg.Servers, servers)
	}

	if len(cfg.Servers) == 0 {
		cfg.Enabled = false
	} else {
		enable, err := StatusLegacy(legacyRoot + pam_radius_legacy)
		if err != nil {
			syslogger.Err(errString, err)
		} else {
			cfg.Enabled = enable
		}
	}

	// fallback was always true
	cfg.Fallback = true

	return
}

func (config *Config) CopyFrom(otherConfig Config) {
	config.Enabled = otherConfig.Enabled
	config.Fallback = otherConfig.Fallback

	config.Servers = make([]Server, len(otherConfig.Servers))
	copy(config.Servers, otherConfig.Servers)
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

//	_, err = os.Stat(radius_server)
//	if err != nil {
//		file, err := os.Create(radius_server)
//		if err != nil {
//			errs = append(errs, err)
//		} else {
//			file.Close()
//		}
//	}

//	_, err = os.Stat(pam_radius)
//	if err != nil {
//		file, err := os.Create(pam_radius)
//		if err != nil {
//			errs = append(errs, err)
//		} else {
//			file.Close()
//		}
//	}

//	cfg.Servers, err = read_server_list()
//	if err != nil {
//		errs = append(errs, err)
//		return
//	}

//	if len(cfg.Servers) == 0 {
//		cfg.Enabled = false
//		err = Disable()
//		if err != nil {
//			errs = append(errs, err)
//		}
//	} else {
//		cfg.Enabled, err = Status()
//		if err != nil {
//			errs = append(errs, err)
//		}
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

func (cfg *Config) Save(oldConfig Config) (errors []error) {

	var err error

	if cfg.Enabled {
		err = Enable(cfg.Fallback)
	} else {
		err = Disable()
	}

	if err != nil {
		errors = append(errors, err)
	}

	for idx, server := range cfg.Servers {
		if server.Port == 0 {
			cfg.Servers[idx].Port = authserver_port
		}
	}

	err = write_server_list(cfg.Servers)
	if err != nil {
		errors = append(errors, err)
	}

	return
}

func (cfg *Config) Verify() (errs []error) {

	if len(cfg.Servers) == 0 && cfg.Enabled {
		err := fmt.Errorf("Enable radius service requires at least 1 server")
		errs = append(errs, err)
		return
	}

	for idx, server := range cfg.Servers {
		if !util.IsIPaddress(server.IPaddr) {
			err := fmt.Errorf("Bad Radius server IP: %s", server.IPaddr)
			errs = append(errs, err)
		}

		for i := idx + 1; i < len(cfg.Servers); i++ {
			if server.IPaddr == cfg.Servers[i].IPaddr {
				err := fmt.Errorf("Duplicate Radius server:  %s", server.IPaddr)
				errs = append(errs, err)
			}
		}

		if server.Secret == "" {
			err := fmt.Errorf("Can not have empty server secret")
			errs = append(errs, err)
		}
	}

	return
}

func (cfg *Config) Factory() {
	cfg.Servers = []Server{}
	cfg.Enabled = false
	cfg.Fallback = false

	return
}

func AddServer(ipAddr string, secret string, port uint16) (err error) {

	if !util.IsIPaddress(ipAddr) {
		err = errors.New(fmt.Sprintf("Bad server address:", ipAddr))
		return
	}

	if secret == "" {
		err = errors.New("Unable to have empty secret")
		return
	}

	cfg_file, err := cfg.LoadConfig(radius_server)
	if err != nil {
		return
	}
	defer cfg_file.Close()

	if !util.IsIPv4Address(ipAddr) {
		ipAddr = "[" + ipAddr + "]"
	}

	if port == 0 {
		port = authserver_port
	}

	port_str := strconv.Itoa(int(port))

	err = cfg_file.AddStrings(ipAddr+":"+port_str, secret)
	if err != nil {
		return
	}

	return
}

func RemoveServer(ipAddr string, port uint16) (err error) {

	if !util.IsIPaddress(ipAddr) {
		err = fmt.Errorf("Bad server address:", ipAddr)
		return
	}

	cfg_file, err := cfg.LoadConfig(radius_server)
	if err != nil {
		return
	}
	defer cfg_file.Close()

	port_str := strconv.Itoa(int(port))
	kv_pairs, err := cfg_file.GetKVPair(ipAddr + ":" + port_str)
	if err != nil {
		return
	}

	if len(kv_pairs) == 0 {
		err = fmt.Errorf("Server %s:%d is not in the list:", ipAddr, port)
		return
	}

	err = cfg_file.Delete(kv_pairs[0])
	if err != nil {
		return
	}

	return
}

func read_server_list(root_path string) (servers []Server, err error) {
	servers = []Server{}

	cfg_file, err := cfg.LoadConfig(root_path + radius_server)
	if err != nil {
		return
	}
	defer cfg_file.Close()

	server_list, err := cfg_file.GetAll()
	if err != nil {
		return
	}

	for _, kv_pair := range server_list {
		var server Server
		var ip []string

		if strings.HasPrefix(kv_pair.Key, "[") &&
			strings.Contains(kv_pair.Key, "]") { //IPv6 address
			ip = strings.SplitAfter(kv_pair.Key, "]:")
			ip[0] = strings.TrimRight(ip[0], ":")
		} else {
			ip = strings.SplitN(kv_pair.Key, ":", 2)
		}

		if util.IsIPaddress(ip[0]) {
			server.IPaddr = ip[0]
		} else {
			syslogger.Err("Radius: Bad IP address in config file %s", radius_server)
			break
		}

		if len(ip) < 2 {
			server.Port = authserver_port
		} else {
			port, _ := strconv.Atoi(ip[1])
			server.Port = uint16(port)
		}

		// ignore line that doesn't have secret
		if len(kv_pair.Values) < 1 {
			syslogger.Err("Radius: Empty secret in config file %s", radius_server)
			break
		}

		server.Secret = kv_pair.Values[0]
		servers = append(servers, server)
	}

	return
}

func write_server_list(servers []Server) (err error) {

	cfg_file, err := cfg.LoadConfig(radius_server)
	if err != nil {
		return
	}
	defer cfg_file.Close()

	err = cfg_file.DeleteAll()
	if err != nil {
		return
	}

	for _, server := range servers {
		if server.Port == 0 {
			server.Port = authserver_port
		}

		ipAddr := server.IPaddr
		if !util.IsIPv4Address(ipAddr) {
			ipAddr = "[" + ipAddr + "]"
		}

		err = cfg_file.AddLine(ipAddr + ":" + strconv.Itoa(int(server.Port)) +
			" " + server.Secret)
		if err != nil {
			break
		}
	}

	return
}

func StatusLegacy(confPath string) (enabled bool, err error) {
	pam_file, err := cfg.LoadConfig(confPath)
	if err != nil {
		return
	}
	defer pam_file.Close()

	kv_pairs, err := pam_file.GetKVPair("auth")
	if err != nil {
		return
	}

	for _, kv_pair := range kv_pairs {
		if len(kv_pair.Values) >= 2 {
			if kv_pair.Values[0] == "sufficient" &&
				kv_pair.Values[1] == "pam_radius_auth.so" {
				enabled = true
			} else {
				enabled = false
			}
		} else {
			err = errors.New("PAM radius config file corrupted")
		}
	}

	return
}

func Enable(fallback bool) (err error) {

	pam_file, err := cfg.LoadConfig(pam_radius)
	if err != nil {
		return
	}
	defer pam_file.Close()

	nss_file, err := cfg.LoadConfig(nss_conf)
	if err != nil {
		return
	}
	defer nss_file.Close()

	err = pam_file.DeleteByKey("auth")
	if err != nil {
		return
	}

	pam_option := pam_option_required

	if fallback {
		pam_option = pam_option_sufficient
	}

	err = pam_file.AddLine(pam_option)
	if err != nil {
		return
	}

	err = nss_file.DeleteByKey(nss_key)
	if err != nil {
		return
	}
	err = nss_file.AddStrings(nss_key, "files", "vega")
	if err != nil {
		return
	}

	return
}

func Disable() (err error) {

	pam_file, err := cfg.LoadConfig(pam_radius)
	if err != nil {
		return
	}
	defer pam_file.Close()

	err = pam_file.DeleteByKey("auth")
	if err != nil {
		return
	}

	nss_file, err := cfg.LoadConfig(nss_conf)
	if err != nil {
		return
	}
	defer nss_file.Close()

	err = nss_file.DeleteByKey(nss_key)
	if err != nil {
		return
	}

	err = nss_file.AddStrings(nss_key, "files")
	if err != nil {
		return
	}

	return
}
