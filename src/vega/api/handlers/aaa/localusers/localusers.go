// Copyright (c) 2016, Virtual Gateway Labs. All rights reserved.

package localusers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"vega/api/handlers"
	"vega/core/aaa/localusers"
)

func Get(ctx handlers.Context) {
	ctx.Encode(ctx.Config.AAA.LocalUsers)
}

func GetUser(ctx handlers.Context) {
	name := ctx.Params.ByName("username")
	if user, err := findUser(ctx.Config.AAA.LocalUsers, name); err != nil {
		ctx.NotFound()
	} else {
		ctx.Encode(user)
	}
}

func PutUser(ctx handlers.Context) {
	username := ctx.Params.ByName("username")
	cfg := ctx.Config.AAA.LocalUsers.Clone()

	var err error
	user, err := findUser(*cfg, username)
	if err != nil {
		ctx.NotFound()
		return
	}

	user.Password = ""

	if !ctx.Decode(&user) {
		ctx.EncodeBadRequests()
		return
	}

	user.Username = username

	errs := localusers.VerifyUsers([]localusers.User{user})
	if len(errs) > 0 {
		ctx.EncodeBadRequests(errs...)
		return
	}

	localusers.SetPassword(user.Username, user.Password)
	if err != nil {
		defer rollback(ctx)
		ctx.EncodeInternalServerErrors(err)
		return
	}

	localusers.SetPrivilege(user.Username, user.Privilege)
	if err != nil {
		defer rollback(ctx)
		ctx.EncodeInternalServerErrors(err)
		return
	}

	err = cfg.LoadUsers()
	if err != nil {
		defer rollback(ctx)
		ctx.EncodeInternalServerErrors(err)
		return
	} else {
		ctx.Config.AAA.LocalUsers.CopyFrom(*cfg)
	}
}

func PutPassByName(ctx handlers.Context) {
	username := ctx.Params.ByName("username")
	cfg := ctx.Config.AAA.LocalUsers.Clone()

	var err error
	user, err := findUser(*cfg, username)
	if err != nil {
		ctx.NotFound()
		return
	}

	user.Password = ""

	if !ctx.Decode(&user.Password) {
		ctx.EncodeBadRequests()
		return
	}

	user.Username = username

	errs := localusers.VerifyUsers([]localusers.User{user})
	if len(errs) > 0 {
		ctx.EncodeBadRequests(errs...)
		return
	}

	localusers.SetPassword(user.Username, user.Password)
	if err != nil {
		defer rollback(ctx)
		ctx.EncodeInternalServerErrors(err)
		return
	}

	err = cfg.LoadUsers()
	if err != nil {
		defer rollback(ctx)
		ctx.EncodeInternalServerErrors(err)
		return
	} else {
		ctx.Config.AAA.LocalUsers.CopyFrom(*cfg)
	}
}

func PutPrivilegeByName(ctx handlers.Context) {
	username := ctx.Params.ByName("username")
	cfg := ctx.Config.AAA.LocalUsers.Clone()

	var err error
	user, err := findUser(*cfg, username)
	if err != nil {
		ctx.NotFound()
		return
	}

	if !ctx.Decode(&user.Privilege) {
		ctx.EncodeBadRequests()
		return
	}

	user.Username = username

	errs := localusers.VerifyUsers([]localusers.User{user})
	if len(errs) > 0 {
		ctx.EncodeBadRequests(errs...)
		return
	}

	localusers.SetPrivilege(user.Username, user.Privilege)
	if err != nil {
		defer rollback(ctx)
		ctx.EncodeInternalServerErrors(err)
		return
	}

	err = cfg.LoadUsers()
	if err != nil {
		defer rollback(ctx)
		ctx.EncodeInternalServerErrors(err)
		return
	} else {
		ctx.Config.AAA.LocalUsers.CopyFrom(*cfg)
	}
}

func Patch(ctx handlers.Context) {
	users := new(localusers.Config)
	if !ctx.Decode(&users) {
		return
	}

	ctx.VerifySave(users, &ctx.Config.AAA.LocalUsers)
}

func Post(ctx handlers.Context) {
	cfg := ctx.Config.AAA.LocalUsers.Clone()

	var users localusers.Config

	var err error
	var errs []error

	if !ctx.Decode(&users) {
		ctx.EncodeBadRequests()
		return
	}

	defer func() {
		if ctx.Config.AAA.RADIUS.Enabled {
			if errs := ctx.Config.AAA.RADIUS.Save(ctx.Config.AAA.RADIUS); errs != nil {
				ctx.EncodeInternalServerErrors(errs...)
				return
			}
		}
	}()

	errs = users.Verify()
	if len(errs) > 0 {
		ctx.EncodeBadRequests(errs...)
		return
	}

	for _, user_cfg := range *cfg {
		for _, user := range users {
			if user_cfg.Username == user.Username {
				errs = append(errs, fmt.Errorf("User %s already exist on the system", user.Username))
			}
		}
	}

	radiusCopy := ctx.Config.AAA.RADIUS.Clone()
	if radiusCopy.Enabled {
		radiusCopy.Enabled = false
		if errs := radiusCopy.Save(ctx.Config.AAA.RADIUS); errs != nil {
			// rollback
			if errs := ctx.Config.AAA.RADIUS.Save(ctx.Config.AAA.RADIUS); errs != nil {
				ctx.EncodeInternalServerErrors(errs...)
				return
			}

			ctx.EncodeInternalServerErrors(errs...)
			return
		}
	}

	if len(errs) == 0 {
		for _, user := range users {
			err = localusers.AddUser(user.Username, user.Password, user.Privilege)
			if err != nil {
				ctx.EncodeInternalServerErrors(err)
				rollback(ctx)
				return
			}
		}

		err = cfg.LoadUsers()
		if err != nil {
			ctx.EncodeInternalServerErrors(err)
			rollback(ctx)
			return
		} else {
			ctx.Config.AAA.LocalUsers.CopyFrom(*cfg)
		}
	} else {
		ctx.EncodeBadRequests(errs...)
	}
}

func rollback(ctx handlers.Context) {
	errs := ctx.Config.AAA.LocalUsers.Save(ctx.Config.AAA.LocalUsers)
	if len(errs) > 0 {
		ctx.EncodeInternalServerErrors(errs...)
	}
}

func Delete(ctx handlers.Context) {
	userList := ctx.Request.URL.Query().Get("users")
	usersToDelete := strings.FieldsFunc(userList, func(c rune) bool { return c == ',' })

	newUsers := localusers.Config{}
	removedSet := make(map[string]bool)

	if len(usersToDelete) > 0 {
	userLoop:
		for _, user := range ctx.Config.AAA.LocalUsers {
			for _, userToDelete := range usersToDelete {
				if userToDelete == user.Username {
					removedSet[userToDelete] = true
					continue userLoop
				}
			}

			newUsers = append(newUsers, user)
		}
	}

	errFlag := false
	errs := []error{}
	for _, remove := range usersToDelete {
		if !removedSet[remove] {
			errFlag = true
			errs = append(errs, errors.New("["+remove+"] is not configured"))
		}
	}

	if errFlag {
		ctx.EncodeErrors(http.StatusNotFound, errs...)
	}

	ctx.VerifySave(&newUsers, &ctx.Config.AAA.LocalUsers)

	//	cfg := ctx.Config.AAA.LocalUsers.Clone()

	//	var users localusers.Config
	//	var users_del localusers.Config

	//	var err error

	//	if !ctx.Decode(&users) {
	//		ctx.EncodeBadRequests()
	//		return
	//	}

	//	for _, user_cfg := range *cfg {
	//		for idx, user := range users {
	//			if user_cfg.Username == user.Username {
	//				users_del = append(users_del, user)

	//				if idx == 0 {
	//					users = append(localusers.Config{}, users[1:]...)
	//				} else {
	//					users = append(users[:idx], users[idx+1:]...)
	//				}
	//			}
	//		}
	//	}

	//	if len(users) > 0 {
	//		var user_str string
	//		for _, user := range users {
	//			user_str += "'" + user.Username + "' "
	//		}

	//		err = fmt.Errorf("User %s does not exist", user_str)
	//		ctx.EncodeBadRequests(err)
	//		return
	//	}

	//	defer func() {
	//		err = cfg.LoadUsers()
	//		if err != nil {
	//			errs := ctx.Config.AAA.LocalUsers.Save(ctx.Config.AAA.LocalUsers)
	//			if len(errs) > 0 {
	//				ctx.EncodeInternalServerErrors(errs...)
	//				return
	//			}

	//			ctx.EncodeInternalServerErrors(err)
	//			return
	//		} else {
	//			ctx.Config.AAA.LocalUsers.CopyFrom(*cfg)
	//		}
	//	}()

	//	for _, user := range users_del {
	//		err = localusers.RemoveUser(user.Username)
	//		if err != nil {
	//			ctx.EncodeInternalServerErrors(err)
	//			return
	//		}
	//	}
}

func findUser(users []localusers.User, query string) (localusers.User, error) {
	for _, usr := range users {
		if query == usr.Username {
			return usr, nil
		}
	}
	return localusers.User{}, errors.New("[" + query + "] is not configured")
}
