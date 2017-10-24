package radius

import (
	"errors"
	"fmt"
	"strconv"
	"github.com/htbig/common/src/vega/syslogger"

	"github.com/kirves/goradius"
)

const RadiusCfgError = "Radius: Failed to read radius config"
const RadiusNoServerError = "Radius: No Radius server is set"
const RadiusAuthError = "Radius: Failed to authenticate with any servers"
const GatewayTimeoutError = "Timed out while waiting for an answer"

func RadiusAuthenticate(username, password string) (privileged, ok bool, errs []error) {
	var err error

	servers, err := read_server_list("")
	if err != nil {
		errs = append(errs, errors.New(RadiusCfgError))
		errs = append(errs, err)
		return
	}

	if len(servers) == 0 {
		errs = append(errs, errors.New(RadiusNoServerError))
		return
	}

	var server_errs []error

	for _, server := range servers {
		address := server.IPaddr
		port := strconv.Itoa(int(server.Port))
		secret := server.Secret

		auth := goradius.Authenticator(address, port, secret)
		privilege, ok, err := auth.AuthenticateWithPrivilege(username, password)

		if err == nil {
			return privilege == 2, ok, nil
		} else {
			err = fmt.Errorf("%s: %s", address, err.Error())
			server_errs = append(server_errs, err)
			syslogger.Err("Radius auth:", err)
		}
	}

	if len(server_errs) > 0 {
		errs = append(errs, errors.New(RadiusAuthError))
		errs = append(errs, server_errs...)
	}

	return
}
