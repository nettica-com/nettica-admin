package shadow

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/amoghe/go-crypt"
)

const secsInDay = 86400

func (e *Entry) IsAccountValid() bool {
	if e.AcctExpiry == -1 {
		return true
	}

	nowDays := int(time.Now().Unix() / secsInDay)
	return nowDays < e.AcctExpiry
}

func (e *Entry) IsPasswordValid() bool {
	if e.LastChange == -1 || e.MaxPassAge == -1 || e.InactivityPeriod == -1 {
		return true
	}

	nowDays := int(time.Now().Unix() / secsInDay)
	return nowDays < e.LastChange+e.MaxPassAge+e.InactivityPeriod
}

func (e *Entry) VerifyPassword(pass string) (err error) {
	// Do not permit null and locked passwords.
	if e.Pass == "" {
		return errors.New("verify: null password")
	}
	if e.Pass[0] == '!' {
		return errors.New("verify: locked password")
	}

	// Get the salt from the password.
	salt := ""
	parts := strings.SplitN(e.Pass, "$", 5)
	switch len(parts) {
	case 4:
		salt = "$" + parts[1] + "$" + parts[2] + "$"
	case 5:
		salt = "$" + parts[1] + "$" + parts[2] + "$" + parts[3] + "$"
	default:
		return errors.New("verify: malformed password")
	}

	// crypt.NewFromHash may panic on unknown hash function.
	defer func() {
		if rcvr := recover(); rcvr != nil {
			err = fmt.Errorf("%v", rcvr)
		}
	}()

	hash, err := crypt.Crypt(pass, salt)
	if err != nil {
		return err
	}
	if hash != e.Pass {
		return ErrWrongPassword
	}
	return nil
}
