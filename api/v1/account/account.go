package account

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nettica-com/nettica-admin/auth"
	core "github.com/nettica-com/nettica-admin/core"
	model "github.com/nettica-com/nettica-admin/model"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
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
		g.PATCH("/:id", updateAccount)
		g.DELETE("/:id", deleteAccount)
	}
}

func activateAccount(c *gin.Context) {
	id := c.Param("id")

	v, err := core.ActivateAccount(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to create account")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, v)
}

func emailAccount(c *gin.Context) {
	id := c.Param("id")

	oauth2Token := c.MustGet("oauth2Token").(*oauth2.Token)
	oauth2Client := c.MustGet("oauth2Client").(auth.Auth)

	user, err := oauth2Client.UserInfo(oauth2Token)
	if err != nil {
		log.WithFields(log.Fields{
			"oauth2Token": oauth2Token,
			"err":         err,
		}).Error("failed to get user with oauth token")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if user.Email == "" {
		log.WithFields(log.Fields{
			"oauth2Token": oauth2Token,
			"err":         err,
		}).Error("failed to get user with oauth token")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	account, err := core.ReadAccount(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read account")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	account.From = user.Email

	err = core.EmailUser(account.Email, account.Id, account.NetId)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to send email to client")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func createAccount(c *gin.Context) {
	var account model.Account

	if err := c.ShouldBindJSON(&account); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to bind")
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	// get creation user from token and add to client infos
	oauth2Token := c.MustGet("oauth2Token").(*oauth2.Token)
	oauth2Client := c.MustGet("oauth2Client").(auth.Auth)
	user, err := oauth2Client.UserInfo(oauth2Token)
	if err != nil {
		log.WithFields(log.Fields{
			"oauth2Token": oauth2Token,
			"err":         err,
		}).Error("failed to get user with oauth token")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	account.From = user.Email

	v, err := core.CreateAccount(&account)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to create account")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, v)
}

func readAllAccounts(c *gin.Context) {
	email := c.Param("id")

	// get creation user from token and add to client infos
	oauth2Token := c.MustGet("oauth2Token").(*oauth2.Token)
	oauth2Client := c.MustGet("oauth2Client").(auth.Auth)
	user, err := oauth2Client.UserInfo(oauth2Token)
	if err != nil {
		log.WithFields(log.Fields{
			"oauth2Token": oauth2Token,
			"err":         err,
		}).Error("failed to get user with oauth token")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if user.Email == "" {
		log.WithFields(log.Fields{
			"oauth2Token": oauth2Token,
			"err":         err,
		}).Error("failed to get email address from valid oauth token")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if email == "" {
		email = user.Email
	}

	accounts, err := core.ReadAllAccounts(email)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read accounts")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, accounts)
}

func updateAccount(c *gin.Context) {
	var data model.Account
	id := c.Param("id")

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
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}
	client, err := core.UpdateAccount(id, &data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to update client")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, client)
}

func deleteAccount(c *gin.Context) {
	id := c.Param("id")

	err := core.DeleteAccount(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to remove client")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
