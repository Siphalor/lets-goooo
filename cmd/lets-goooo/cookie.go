package main

import (
	"fmt"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/journal"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/util"
	"strings"
)

func Validate(string2 string) (journal.User, error) {

	userData := strings.Split(string2, ";")

	if util.Base64Encode(util.HashString(userData[0]+"\t"+secret)) == userData[1] {
		userData0, _ := util.Base64Decode(userData[0])
		userData = strings.Split(string(userData0), "\t")
		user := journal.User{
			Name:    userData[0],
			Address: userData[1],
		}

		return user, nil
	}

	return journal.User{
		Name:    "",
		Address: "",
	}, fmt.Errorf("user Data in Cookie did not match the hash")
}
