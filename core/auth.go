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
func AuthFromContext(c *gin.Context, id string) (*model.Account, interface{}, error) {

	var accounts []*model.Account
	var device *model.Device
	var account *model.Account
	var service *model.Service
	var err error

	apikey := c.Request.Header.Get("X-API-KEY")

	if strings.HasPrefix(apikey, "nettica-api-") {

		account, err = GetAccountFromApiKey(apikey)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return nil, nil, err
		}
		if account.Status == "Suspended" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return nil, nil, errors.New("account is suspended")
		}

	} else if strings.HasPrefix(apikey, "device-api-") {

		device, err = ReadDevice(id)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return nil, nil, err
		}

	} else if strings.HasPrefix(apikey, "service-api-") {

		service, err = ReadService(id)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return nil, nil, err
		}

	} else {

		value, exists := c.Get("oauth2Token")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return nil, nil, errors.New("failed to get oauth2Token from context")
		}
		oauth2Token := value.(*oauth2.Token)
		oauth2Client := c.MustGet("oauth2Client").(model.Authentication)
		user, err := oauth2Client.UserInfo(oauth2Token)
		if err != nil {
			log.WithFields(log.Fields{
				"oauth2Token": oauth2Token,
				"err":         err,
			}).Error("failed to get user with oauth token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return nil, nil, err
		}

		if user.Email == "" {
			log.WithFields(log.Fields{
				"oauth2Token": oauth2Token,
				"err":         err,
			}).Error("SECURITY ALERT: failed to get user email with valid oauth token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return nil, nil, errors.New("SECURITY ALERT: failed to get user email with valid oauth token")
		}

		accounts, err = ReadAllAccounts(user.Email)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return nil, nil, err
		}

		if len(accounts) > 0 {
			account = accounts[0]
		}

		if len(accounts) == 0 {
			log.Errorf("ERROR: Failed to get account for user %s", user.Email)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return nil, nil, errors.New("failed to get account for user")
		}

	}

	if strings.HasPrefix(id, "device-") {

		// if we already have the device use it
		if device != nil {
			if device.Id == id {
				// do nothing
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "device id mismatch"})
				return nil, nil, errors.New("device id mismatch")
			}
		} else {
			device, err = ReadDevice(id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
				return nil, nil, err
			}
		}

		for _, a := range accounts {
			if a.Id == device.AccountID {
				account = a
				break
			}
		}

		if account.Status == "Suspended" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return nil, nil, errors.New("account is suspended")
		}

		return account, device, nil
	}

	if strings.HasPrefix(id, "account-") {

		a, err := ReadAccount(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
			return nil, nil, err
		}

		if account.Status == "Suspended" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return nil, nil, errors.New("account is suspended")
		}

		return account, a, nil
	}

	if strings.HasPrefix(id, "net-") {

		net, err := ReadNet(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
			return nil, nil, err
		}

		for _, a := range accounts {
			if a.Id == net.AccountID {
				account = a
				break
			}
		}

		if account.Status == "Suspended" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return nil, nil, errors.New("account is suspended")
		}

		return account, net, nil
	}

	if strings.HasPrefix(id, "vpn-") {

		vpn, err := ReadVPN(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
			return nil, nil, err
		}

		for _, a := range accounts {
			if a.Id == vpn.AccountID {
				account = a
				break
			}
		}

		if account.Status == "Suspended" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return nil, nil, errors.New("account is suspended")
		}

		return account, vpn, nil
	}

	if strings.HasPrefix(id, "service-") {

		if service != nil {
			if service.Id == id {
				// do nothing
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "service id mismatch"})
				return nil, nil, errors.New("service id mismatch")
			}
		} else {

			service, err = ReadService(id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
				return nil, nil, err
			}
		}

		for _, a := range accounts {
			if a.Id == service.AccountID {
				account = a
				break
			}
		}

		if account.Status == "Suspended" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return nil, nil, errors.New("account is suspended")
		}

		return account, service, nil
	}

	return account, nil, nil

}
