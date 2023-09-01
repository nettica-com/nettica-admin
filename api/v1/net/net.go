package net

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	auth "github.com/nettica-com/nettica-admin/auth"
	core "github.com/nettica-com/nettica-admin/core"
	model "github.com/nettica-com/nettica-admin/model"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// ApplyRoutes applies router to gin Router
func ApplyRoutes(r *gin.RouterGroup) {
	g := r.Group("/net")
	{

		g.POST("", createNet)
		g.GET("/:id", readNet)
		g.PATCH("/:id", updateNet)
		g.DELETE("/:id", deleteNet)
		g.GET("", readNetworks)
	}
}

func createNet(c *gin.Context) {
	var data model.Network

	if err := c.ShouldBindJSON(&data); err != nil {
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
	data.CreatedBy = user.Email
	data.UpdatedBy = user.Email

	if data.AccountID == "" {
		data.AccountID = user.AccountID
	}

	client, err := core.CreateNet(&data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to create net")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, client)
}

func readNet(c *gin.Context) {
	id := c.Param("id")

	client, err := core.ReadNet(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read client")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, client)
}

func updateNet(c *gin.Context) {
	var data model.Network
	id := c.Param("id")

	if err := c.ShouldBindJSON(&data); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to bind")
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	// get update user from token and add to client infos
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

	net, err := core.ReadNet(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read net")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	account, err := core.GetAccount(user.Email, net.AccountID)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read account")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if account == nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read account")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	authorized := false

	if net.CreatedBy == user.Email || account.Role == "Admin" || account.Role == "Owner" {
		authorized = true
	}

	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	data.UpdatedBy = user.Email

	client, err := core.UpdateNet(id, &data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to update client")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, client)
}

func deleteNet(c *gin.Context) {
	id := c.Param("id")

	err := core.DeleteNet(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to remove client")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func readNetworks(c *gin.Context) {
	value, exists := c.Get("oauth2Token")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	oauth2Token := value.(*oauth2.Token)
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

	if user.Email == "" && os.Getenv("OAUTH2_PROVIDER_NAME") != "fake" {
		log.Error("security alert: Email empty on authenticated token")
		c.AbortWithStatus(http.StatusForbidden)
	}

	nets, err := core.ReadNetworks(user.Email)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to list nets")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, nets)
}
