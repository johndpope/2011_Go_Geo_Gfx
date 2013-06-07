package usermanager

import (
	"fmt"
	"strings"
)

type UserManager struct {
	IsToLower bool
}

func (this *UserManager) ToUpperOrLower (val string, result *string) error {
	fmt.Printf("ToUpperOrLower: in=%s\n", val)
	if this.IsToLower {
		val = strings.ToLower(val)
	} else {
		val = strings.ToUpper(val)
	}
	*result = val
	fmt.Printf("ToUpperOrLower: out=%s\n", *result)
	return nil
}

var (
	UserMan = &UserManager {}
)
