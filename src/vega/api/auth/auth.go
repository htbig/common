// Copyright (c) 2016, Virtual Gateway Labs. All rights reserved.

package auth

import (
	"errors"
	"os/exec"
	"strings"

	"vega/core/aaa/radius"

	"github.com/msteinert/pam"
)

//func authenticateAdminRADIUS(username, password string) (bool, bool, []error) {
//	privileged, ok, errs := radius.RadiusAuthenticate(username, password)
//	return ok, privileged, errs
//}

//func authenticateAdminPAM(username, password string) (bool, bool, []error) {
//	const serviceName = "login"
//	transaction, err := pam.StartFunc(
//		serviceName,
//		username,
//		func(style pam.Style, message string) (string, error) {
//			switch style {
//			case pam.PromptEchoOff:
//				return password, nil
//			case pam.PromptEchoOn, pam.ErrorMsg, pam.TextInfo:
//				return "", nil
//			}
//			return "", errors.New("Invalid PAM message style")
//		},
//	)

//	if err != nil {
//		return false, false, []error{err}
//	}

//	if err = transaction.Authenticate(pam.Flags(0)); err != nil {
//		return false, false, nil
//	}

//	//
//	// check current user against privilege group
//	const privilegedGroup = "wheel"

//	bytes, err := exec.Command("groups", username).Output()
//	if err != nil {
//		// fail to check groups
//		return false, false, []error{err}
//	}

//	groups := strings.Fields(string(bytes))
//	for _, group := range groups {
//		if group == privilegedGroup {
//			// in privileged group
//			return true, true, nil
//		}
//	}

//	// not in privileged group
//	return true, false, nil
//}

//func authenticateAdmin(radius bool, username, password string) (bool, bool, []error) {
//	if radius {
//		return authenticateAdminRADIUS(username, password)
//	} else {
//		return authenticateAdminPAM(username, password)
//	}
//}

//func authenticateUserRADIUS(username, password string) (bool, []error) {
//	_, ok, errs := radius.RadiusAuthenticate(username, password)

//	if len(errs) > 0 {
//		return false, errs
//	}

//	return ok, nil
//}

//func authenticateUserPAM(username, password string) (bool, []error) {
//	const serviceName = "login"
//	transaction, err := pam.StartFunc(
//		serviceName,
//		username,
//		func(style pam.Style, message string) (string, error) {
//			switch style {
//			case pam.PromptEchoOff:
//				return password, nil
//			case pam.PromptEchoOn, pam.ErrorMsg, pam.TextInfo:
//				return "", nil
//			}
//			return "", errors.New("Invalid PAM message style")
//		},
//	)

//	if err != nil {
//		return false, []error{err}
//	}

//	if err = transaction.Authenticate(pam.Flags(0)); err != nil {
//		return false, nil
//	}

//	return true, nil
//}

//func authenticateUser(radius bool, username, password string) (bool, []error) {
//	if radius {
//		return authenticateUserRADIUS(username, password)
//	} else {
//		return authenticateUserPAM(username, password)
//	}
//}

func authenticatePAM(checkPrivilege bool, username, password string) (bool, bool, []error) {
	const serviceName = "login"
	transaction, err := pam.StartFunc(
		serviceName,
		username,
		func(style pam.Style, message string) (string, error) {
			switch style {
			case pam.PromptEchoOff:
				return password, nil
			case pam.PromptEchoOn, pam.ErrorMsg, pam.TextInfo:
				return "", nil
			}
			return "", errors.New("Invalid PAM message style")
		},
	)

	if err != nil {
		return false, false, []error{err}
	}

	if err = transaction.Authenticate(pam.Flags(0)); err != nil {
		return false, false, nil
	}

	if !checkPrivilege {
		return true, true, nil
	}

	//
	// check current user against privilege group
	const privilegedGroup = "wheel"

	bytes, err := exec.Command("groups", username).Output()
	if err != nil {
		// fail to check groups
		return false, false, []error{err}
	}

	groups := strings.Fields(string(bytes))
	for _, group := range groups {
		if group == privilegedGroup {
			// in privileged group
			return true, true, nil
		}
	}

	// not in privileged group
	return true, false, nil
}

func authenticateRADIUS(checkPrivilege bool, username, password string) (bool, bool, []error) {
	// reverse returned values
	privileged, ok, errs := radius.RadiusAuthenticate(username, password) // authorized, authenticated, errors

	if len(errs) > 0 {
		return false, false, errs
	} else {
		if checkPrivilege {
			return ok, privileged, nil
		} else {
			return ok, true, nil
		}
	}
}

func AuthenticateAPI(radius, fallback, checkPrivilege bool, username, password string) (bool, bool, []error) {
	var authenticated, authorized bool
	var errs []error = []error{}

	if radius {
		authenticated, authorized, errs = authenticateRADIUS(checkPrivilege, username, password)
	}

	if !radius || (fallback && !authenticated) {
		authenticated, authorized, errs = authenticatePAM(checkPrivilege, username, password)

	}

	return authenticated, authorized, errs
}
