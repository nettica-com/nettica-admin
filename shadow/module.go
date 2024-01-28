package shadow

import (
	"errors"
	"fmt"
	"os"
)

func ShadowInit() error {

	f, err := os.Open("/etc/shadow")
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("shadow: can't read /etc/shadow due to permission error, use helper binary or run maddy as a privileged user")
		}
		return fmt.Errorf("shadow: can't read /etc/shadow: %v", err)
	}
	f.Close()

	return nil
}

func ShadowLookup(username string) (string, bool, error) {

	ent, err := Lookup(username)
	if err != nil {
		if errors.Is(err, ErrNoSuchUser) {
			return "", false, nil
		}
		return "", false, err
	}

	if !ent.IsAccountValid() {
		return "", false, nil
	}

	return "", true, nil
}

func ShadowAuthPlain(username, password string) error {

	ent, err := Lookup(username)
	if err != nil {
		return err
	}

	if !ent.IsAccountValid() {
		return fmt.Errorf("shadow: account is expired")
	}

	if !ent.IsPasswordValid() {
		return fmt.Errorf("shadow: password is expired")
	}

	if err := ent.VerifyPassword(password); err != nil {
		if errors.Is(err, ErrWrongPassword) {
			return ErrWrongPassword
		}
		return err
	}

	return nil
}
