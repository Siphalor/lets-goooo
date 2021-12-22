package main

import (
	"fmt"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/internal/journal"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/internal/util"
	"strings"
)

func Validate(string2 string) (journal.User, error) {

	user := journal.User{
		Name:    "",
		Address: "",
	}
	userData := strings.Split(string2, ":")
	if len(userData) < 2 {
		return user, fmt.Errorf("cookie did not contain separator ':'")
	}

	if util.Base64Encode(util.HashString(userData[0]+"\t"+cookieSecret)) == userData[1] {
		userData0, _ := util.Base64Decode(userData[0])
		userData = strings.Split(string(userData0), "\t")
		user := journal.User{
			Name:    userData[0],
			Address: userData[1],
		}

		return user, nil
	}

	return user, fmt.Errorf("user Data in Cookie did not match the hash")
}
