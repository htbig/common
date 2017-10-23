// Copyright (c) 2016, Virtual Gateway Labs. All rights reserved.

// Package util provide helpers and tools
package util

import (
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/vaughan0/ini"
)

// Determine whether a string a IP address
func IsIPaddress(str string) bool {

	ip := net.ParseIP(str)
	if ip != nil {
		return true
	} else {
		return false
	}
}

// Determine whether a string a IP v4 address
func IsIPv4Address(str string) bool {

	ip := net.ParseIP(str)
	if ip == nil {
		return false
	}

	if strings.Contains(str, ".") {
		return true
	} else {
		return false
	}
}

// Determine whether a string a IP v6 address
func IsIPv6Address(str string) bool {

	ip := net.ParseIP(str)
	if ip == nil {
		return false
	}

	if strings.Contains(str, ":") {
		return true
	} else {
		return false
	}
}

// Get IPv4 address in a String
func GetIPInString(input string) string {
	numBlock := "(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])"
	regexPattern := numBlock + "\\." + numBlock + "\\." + numBlock + "\\." + numBlock

	regEx := regexp.MustCompile(regexPattern)
	return regEx.FindString(input)
}

func GetMACAddr(interfaceName string) (mac net.HardwareAddr, err error) {
	net_interface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return
	}

	mac = net_interface.HardwareAddr

	return
}

func IsServiceActive(service string) (enabled bool, err error) {
	stdout, err := exec.Command("systemctl", "show", service).Output()
	if err != nil {
		return
	}

	var svc_info ini.File
	svc_info, err = ini.Load(strings.NewReader(string(stdout)))
	if err != nil {
		return
	}

	var is_active, ok = svc_info.Get("", "ActiveState")
	if ok && is_active == "active" {
		enabled = true
	}

	return
}

func StartService(service string) (err error) {

	stdout, err := exec.Command("systemctl", "start", service).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%s:%s:%s", service, err.Error(), string(stdout))
		err = ErrorWithInfo(err)
		return
	}

	return
}

func StopService(service string) (err error) {

	stdout, err := exec.Command("systemctl", "stop", service).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%s:%s:%s", service, err.Error(), string(stdout))
		err = ErrorWithInfo(err)
		return
	}

	return
}

func RestartService(service string) (err error) {

	if err := StopService(service); err != nil {
		return err
	} else {
		return StartService(service)
	}

	//	stdout, err := exec.Command("systemctl", "restart", service).CombinedOutput()
	//	if err != nil {
	//		err = fmt.Errorf("%s:%s:%s", service, err.Error(), string(stdout))
	//		err = ErrorWithInfo(err)
	//		return
	//	}

	//	return
}

const path_sysfs = "/sys/class/"

func ReadSysFS(path string) (output string, err error) {
	full_path := fmt.Sprint(path_sysfs + path)

	data, err := ioutil.ReadFile(full_path)
	if err != nil {
		return
	}

	output = strings.Trim(string(data), "\n")

	return
}

func ErrorWithInfo(errin error) (err error) {

	defer recover()

	pc, _, _, _ := runtime.Caller(1)
	fields := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	func_name := fields[len(fields)-1]

	err = fmt.Errorf("%s:%s", func_name, errin.Error())

	return
}
