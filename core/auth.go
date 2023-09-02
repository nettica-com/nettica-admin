package core

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	model "github.com/nettica-com/nettica-admin/model"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// AuthFromContext takes the gin context, the id from the request and the account and
// object used to find the account, along with any error.
//
//	Example: account, device, err := GetFromContext(c, id)
func GetFromContext(c *gin.Context, id string) (*model.Account, interface{}, error) {

	var accounts []*model.Account
	var device *model.Device
	var account *model.Account
	var err error

	apikey := c.Request.Header.Get("X-API-KEY")

	if strings.HasPrefix(apikey, "nettica-api-") {

		account, err = GetAccountFromApiKey(apikey)

		if err != nil {
			return nil, nil, err
		}

	} else if strings.HasPrefix(apikey, "device-api-") {

		device, err = ReadDevice(id)
		if err != nil {
			return nil, nil, err
		}

	} else {

		oauth2Token := c.MustGet("oauth2Token").(*oauth2.Token)
		oauth2Client := c.MustGet("oauth2Client").(model.Authentication)
		user, err := oauth2Client.UserInfo(oauth2Token)
		if err != nil {
			log.WithFields(log.Fields{
				"oauth2Token": oauth2Token,
				"err":         err,
			}).Error("failed to get user with oauth token")
			c.AbortWithStatus(http.StatusUnauthorized)
			return nil, nil, err
		}

		if user.Email == "" {
			log.WithFields(log.Fields{
				"oauth2Token": oauth2Token,
				"err":         err,
			}).Error("SECURITY ALERT: failed to get user email with valid oauth token")
			c.AbortWithStatus(http.StatusUnauthorized)
			return nil, nil, errors.New("SECURITY ALERT: failed to get user email with valid oauth token")
		}

		accounts, err = ReadAllAccounts(user.Email)
		if err != nil {
			return nil, nil, err
		}

	}

	if strings.HasPrefix(id, "device-") {

		// if we already have the device use it
		if device != nil {
			if device.Id == id {
				// do nothing
			} else {
				return nil, nil, errors.New("device id mismatch")
			}
		} else {
			device, err = ReadDevice(id)
			if err != nil {
				return nil, nil, err
			}
		}

		if accounts != nil {
			for _, a := range accounts {
				if a.Id == device.AccountID {
					account = a
					break
				}
			}
		}

		return account, device, nil
	}

	if strings.HasPrefix(id, "account-") {

		a, err := ReadAccount(id)
		if err != nil {
			return nil, nil, err
		}

		return account, a, nil
	}

	if strings.HasPrefix(id, "net-") {

		net, err := ReadNet(id)
		if err != nil {
			return nil, nil, err
		}

		if accounts != nil {
			for _, a := range accounts {
				if a.Id == net.AccountID {
					account = a
					break
				}
			}
		}

		return account, net, nil
	}

	if strings.HasPrefix(id, "vpn-") {

		vpn, err := ReadVPN(id)
		if err != nil {
			return nil, nil, err
		}

		if accounts != nil {
			for _, a := range accounts {
				if a.Id == vpn.AccountID {
					account = a
					break
				}
			}
		}

		return account, vpn, nil
	}

	return account, nil, nil

}
