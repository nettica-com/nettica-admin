package account

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	core "github.com/nettica-com/nettica-admin/core"
	model "github.com/nettica-com/nettica-admin/model"
	log "github.com/sirupsen/logrus"

	"github.com/auth0/go-auth0/management"
)

// ApplyRoutes applies router to gin Router
func ApplyRoutes(r *gin.RouterGroup) {
	g := r.Group("/accounts")
	{

		g.GET("/", readAllAccounts)
		g.POST("/", createAccount)
		g.POST("/:id/activate", activateAccount)
		g.GET("/:id/invite", emailAccount)
		g.GET("/:id", readAllAccounts)
		g.GET("/:id/users", readUsers)
		g.GET("/:id/limits", getLimits)
		g.PATCH("/:id", updateAccount)
		g.DELETE("/:id", deleteAccount)
		g.DELETE("/:id/soft", softDeleteAccount)
	}
}

// ActivateAccount activates an account from pending to active
// @Summary Activate an account
// @Description Set an account to "active"
// @Tags accounts
// @Security none
// @Success 200 {object} model.Account
// @Failure 400 {object} error
// @Router /accounts/{id}/activate [post]
// @Param id path string true "Account ID"
func activateAccount(c *gin.Context) {
	id := c.Param("id")

	v, err := core.ActivateAccount(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to create account")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, v)
}

// EmailAccount sends an email invitation to join an account
// @Summary Email an account invitation
// @Description Send an email invitation to join an account
// @Tags accounts
// @Security apiKey
// @Success 200 {object} string "OK"
// @Failure 400 {object} error
// @Router /accounts/{id}/invite [get]
// @Param id path string true "Account ID"
func emailAccount(c *gin.Context) {
	id := c.Param("id")

	account, v, err := core.AuthFromContext(c, id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read account")
		return
	}
	a := v.(*model.Account)

	log.Infof("emailAccount: %s sending invite to %s", account.Email, a.Email)

	err = core.EmailUser(a.Email, a.Id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to send email to client")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// CreateAccount creates a new account
// @Summary Create a new account
// @Description Create a new account
// @Tags accounts
// @Security apiKey
// @Accept  json
// @Produce  json
// @Param account body model.Account true "Account"
// @Success 200 {object} model.Account
// @Failure 400 {object} error
// @Router /accounts [post]
func createAccount(c *gin.Context) {
	var account model.Account

	if err := c.ShouldBindJSON(&account); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to bind")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	id := ""
	if account.Parent != "" {
		id = account.Parent
	}

	acnt, _, err := core.AuthFromContext(c, id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read account from context")
		return
	}

	if acnt.Role == "User" || acnt.Role == "Guest" {
		log.Infof("createAccount: %s is not authorized to create an account", acnt.Email)
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to create an account"})
		return
	}

	if acnt.Parent != account.Parent {
		log.Infof("createAccount: %s is not authorized to create an account for %s", acnt.Email, account.Parent)
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to create an account for that parent account"})
		return
	}

	if core.EnforceLimits() {
		// check if the account has reached the limits
		members, err := core.ReadAllAccounts(id)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to read accounts")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		limits, err := core.ReadLimits(id)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to read limits")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if limits.MembersLimitReached(len(members)) {
			log.Infof("createAccount: %s has reached the members limit", acnt.Email)
			c.JSON(http.StatusForbidden, gin.H{"error": "User limit reached"})
			return
		}
	}

	account.Email = strings.ToLower(account.Email)

	account.CreatedBy = acnt.Email
	account.UpdatedBy = acnt.Email

	v, err := core.CreateAccount(&account)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to create account")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, v)
}

// ReadAllAccounts reads all accounts for a user
// @Summary Read all accounts for a user
// @Description Read all accounts for a user
// @Tags accounts
// @Security apiKey
// @Success 200 {array} model.Account
// @Failure 400 {object} error
// @Router /accounts [get]
// @Router /accounts/{id} [get]
// @Param id path string false "Account ID"
func readAllAccounts(c *gin.Context) {
	email := c.Param("id")

	account, _, err := core.AuthFromContext(c, "")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read account from context")
		return
	}

	if email == "" {
		email = account.Email
	}

	accounts, err := core.ReadAllAccounts(email)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read accounts")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, accounts)
}

// ReadUsers reads all users for an account
// @Summary Read all users for an account
// @Description Read all users for an account
// @Tags accounts
// @Security apiKey
// @Success 200 {array} model.Account
// @Failure 400 {object} error
// @Router /accounts/{id}/users [get]
// @Param id path string true "Account ID"
func readUsers(c *gin.Context) {
	id := c.Param("id")

	account, _, err := core.AuthFromContext(c, "")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read account from context")
		return
	}

	accounts, err := core.ReadAllAccounts(account.Email)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read accounts")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	role := "None"
	acnt := &model.Account{}
	for _, a := range accounts {
		if a.Parent == id && a.Email == account.Email {
			role = a.Role
			acnt = a
			break
		}
	}

	if role == "None" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if role == "User" || role == "Guest" {
		result := []*model.Account{}
		result = append(result, acnt)
		c.JSON(http.StatusOK, result)
		return
	}

	if role == "Admin" || role == "Owner" {
		accounts, err = core.ReadAllAccounts(id)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to read accounts")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, accounts)
}

// UpdateAccount updates an account
// @Summary Update an account
// @Description Update an account
// @Tags accounts
// @Security apiKey
// @Accept  json
// @Produce  json
// @Param id path string true "Account ID"
// @Param account body model.Account true "Account"
// @Success 200 {object} model.Account
// @Failure 400 {object} error
// @Router /accounts/{id} [patch]
func updateAccount(c *gin.Context) {
	var data model.Account
	id := c.Param("id")

	account, v, err := core.AuthFromContext(c, id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read account from context")
		return
	}
	update := v.(*model.Account)

	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
		log.Infof("updateAccount - %s", string(bodyBytes))
	}

	// Restore the io.ReadCloser to its original state
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	log.Infof("updateAccount - %s", string(bodyBytes))

	if err := c.ShouldBindJSON(&data); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to bind")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// The account owner may want to test something as a user,
	// and then set itself back to being the owner.
	if account.Id == account.Parent {
		account.Role = "Owner"
	}

	// check if the account is authorized to update this account

	if account.Role == "Admin" || account.Role == "Owner" {
		update = &data
	} else if account.Id == id {
		if update.NetName != data.NetName {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update that field"})
			return
		}
		update.Name = data.Name
		update.Picture = data.Picture
		update.Email = strings.ToLower(data.Email)
		update.ApiKey = data.ApiKey
	} else {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this account"})
		return
	}

	data.UpdatedBy = account.Email

	result, err := core.UpdateAccount(id, update)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to update account")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteAccount deletes an account
// @Summary Delete an account
// @Description Delete an account
// @Tags accounts
// @Security apiKey
// @Success 200 {object} string "OK"
// @Failure 400 {object} error
// @Router /accounts/{id} [delete]
// @Param id path string true "Account ID"
func deleteAccount(c *gin.Context) {
	id := c.Param("id")

	account, _, err := core.AuthFromContext(c, id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read account from context")
		return
	}

	if account.Role == "User" || account.Role == "Guest" {
		if account.Id != id {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this account"})
			return
		}
	}

	err = core.DeleteAccount(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to remove client")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

// SoftDeleteAccount soft deletes an account
// @Summary Soft delete an account
// @Description Soft delete an account.  All devices, networks, and services must be deleted first.
// @Tags accounts
// @Security apiKey
// @Success 200 {object} string "OK"
// @Failure 400 {object} error
// @Router /accounts/{id}/soft [delete]
// @Param id path string true "Account ID"
func softDeleteAccount(c *gin.Context) {
	id := c.Param("id")

	account, _, err := core.AuthFromContext(c, id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read account from context")
		return
	}

	if account.Role == "User" || account.Role == "Guest" {
		if account.Id != id {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this account"})
			return
		}
	}

	// for now at least, only the account principal can soft delete their account.  Admins can hard delete.
	if account.Id != id {
		c.JSON(http.StatusForbidden, gin.H{"error": "You cannot delete this account"})
		return
	}

	var devices []*model.Device
	devices, err = core.ReadDevicesForAccount(id)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// only direct account devices must be deleted
	if len(devices) > 0 {
		for _, d := range devices {
			if d.AccountID == id {
				c.JSON(http.StatusForbidden, gin.H{"error": "You must delete all devices before deleting this account"})
				return
			}
		}
	}

	var networks []*model.Network
	networks, err = core.ReadNetworksForAccount(id)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// don't require parent account networks to be deleted, only direct account networks
	if len(networks) > 0 {
		for _, n := range networks {
			if n.AccountID == id {
				c.JSON(http.StatusForbidden, gin.H{"error": "You must delete all networks before deleting this account"})
				return
			}
		}
	}

	var services []*model.Service
	services, err = core.ReadServicesForAccount(id)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// don't require parent account services to be deleted
	if len(services) > 0 {
		for _, s := range services {
			if s.AccountID == id {
				c.JSON(http.StatusForbidden, gin.H{"error": "You must delete all services before deleting this account"})
				return
			}
		}
	}

	// Delete the account
	log.Infof("softDeleteAccount: %s deleting account %s", account.Email, id)

	err = core.DeleteAccount(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to soft remove client")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// So much for a soft delete, Apple wants the account gone - hard delete
	auth0 := os.Getenv("USE_AUTH0")
	if auth0 == "true" && strings.Contains(account.Sub, "apple") {

		domain := os.Getenv("OAUTH2_PROVIDER_URL")
		clientid := os.Getenv("OAUTH2_CLIENT_ID")
		secret := os.Getenv("OAUTH2_CLIENT_SECRET")

		m, err := management.New(domain, management.WithClientCredentials(context.TODO(), clientid, secret))
		if err == nil {
			log.Infof("Deleting user %s (%s)from auth0", account.Email, account.Sub)
			err = m.User.Delete(context.TODO(), account.Sub)
			if err != nil {
				log.Errorf("Error deleting auth0 apple user %s %v", account.Sub, err)
			}
		} else {
			log.Errorf("Error talking to auth0: %v", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

// GetLimits gets the limits for an account
// @Summary Get the limits for an account
// @Description Get the limits for an account
// @Tags accounts
// @Security apiKey
// @Success 200 {object} model.Limits
// @Failure 400 {object} error
// @Router /accounts/{id}/limits [get]
// @Param id path string true "Account ID"
func getLimits(c *gin.Context) {
	id := c.Param("id")

	account, _, err := core.AuthFromContext(c, id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read account from context")
		return
	}

	if account.Role == "User" || account.Role == "Guest" {
		if account.Id != id {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to read this account"})
			return
		}
	}

	limits, err := core.ReadLimits(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read limits")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	accounts, err := core.ReadAllAccounts(id)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	limits.Members = len(accounts)

	devices, err := core.ReadDevicesForAccount(id)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	limits.Devices = len(devices)

	networks, err := core.ReadNetworksForAccount(id)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	limits.Networks = len(networks)

	services, err := core.ReadServicesForAccount(id)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	limits.Services = len(services)

	c.JSON(http.StatusOK, limits)
}
