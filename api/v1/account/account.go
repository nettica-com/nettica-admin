package account

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	core "github.com/nettica-com/nettica-admin/core"
	model "github.com/nettica-com/nettica-admin/model"
	log "github.com/sirupsen/logrus"
)

// ApplyRoutes applies router to gin Router
func ApplyRoutes(r *gin.RouterGroup) {
	g := r.Group("/accounts")
	{

		g.GET("/", readAllAccounts)
		g.POST("/", createAccount)
		g.POST("/:id/activate", activateAccount)
		g.PATCH("/:id/activate", activateAccount)
		g.GET("/:id/invite", emailAccount)
		g.GET("/:id", readAllAccounts)
		g.GET("/:id/users", readUsers)
		g.PATCH("/:id", updateAccount)
		g.DELETE("/:id", deleteAccount)
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
// @Router /accounts/{id}/activate [patch]
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

	a.From = account.Email

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

	acnt, _, err := core.AuthFromContext(c, "")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read account from context")
		return
	}

	account.From = acnt.Email

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

	account, err = core.GetAccount(account.Email, update.Parent)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read account")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if account == nil || account.Role == "User" || account.Role == "Guest" {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	client, err := core.UpdateAccount(id, &data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to update client")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, client)
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
