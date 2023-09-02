package service

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	core "github.com/nettica-com/nettica-admin/core"
	model "github.com/nettica-com/nettica-admin/model"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// ApplyRoutes applies router to gin Router
func ApplyRoutes(r *gin.RouterGroup) {
	g := r.Group("/service")
	{

		g.POST("", createService)
		g.GET("/:id", readService)
		g.PATCH("/:id", updateService)
		g.DELETE("/:id", deleteService)
		g.GET("/:id/status", statusService)
		g.GET("", readServices)
	}
}

func statusService(c *gin.Context) {

	if c.Param("id") == "" {
		log.Error("servicegroup cannot be empty")
		c.AbortWithStatus(http.StatusForbidden)
	}
	serviceGroup := c.Param("id")

	apikey := c.Request.Header.Get("X-API-KEY")
	etag := c.Request.Header.Get("If-None-Match")

	server, err := core.ReadServer2(serviceGroup)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read server config")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	authorized := false

	if server.ServiceApiKey == apikey {
		authorized = true
	}

	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	services, err := core.ReadServiceVPN(serviceGroup)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read services config")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	var msg model.ServiceMessage
	sConfig := make([]model.Service, len(services))

	msg.Id = serviceGroup

	for i, s := range services {
		sConfig[i] = *s
	}
	msg.Config = sConfig

	bytes, err := json.Marshal(msg)
	if err != nil {
		log.Errorf("cannot marshal msg %v", err)
	}
	md5 := fmt.Sprintf("%x", md5.Sum(bytes))
	if md5 == etag {
		c.AbortWithStatus(http.StatusNotModified)
	} else {
		c.Header("ETag", md5)
		c.JSON(http.StatusOK, msg)
	}

	//	StatusCache.Set(id, msg, 0)
}

func createService(c *gin.Context) {
	var data model.Service

	if err := c.ShouldBindJSON(&data); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to bind")
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	account, _, err := core.GetFromContext(c, data.AccountID)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	data.CreatedBy = account.Email

	if data.AccountID == "" {
		data.AccountID = account.Id
	}

	if account.Role != "Admin" && account.Role != "Owner" {
		log.Errorf("createService: You must be an admin with credits to create a service")
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	client, err := core.CreateService(&data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to create net")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, client)
}

func readService(c *gin.Context) {
	id := c.Param("id")

	account, service, err := core.GetFromContext(c, id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if account.Status == "Suspended" {
		log.Errorf("readService: Account %s is suspended", account.Email)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	c.JSON(http.StatusOK, service)
}

func updateService(c *gin.Context) {
	var data model.Service
	id := c.Param("id")

	if err := c.ShouldBindJSON(&data); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to bind")
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	account, v, err := core.GetFromContext(c, id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	service := v.(*model.Service)

	if account.Role != "Admin" && account.Role != "Owner" {
		log.Errorf("updateService: You must be an admin to update a service")
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	apikey := c.Request.Header.Get("X-API-KEY")

	if apikey != "" {

		authorized := false

		if service.ApiKey == apikey {
			authorized = true
		}

		if !authorized {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		data.UpdatedBy = "API"

	} else {

		data.UpdatedBy = account.Email
	}
	client, err := core.UpdateService(id, &data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to update client")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, client)
}

func deleteService(c *gin.Context) {
	id := c.Param("id")

	account, _, err := core.GetFromContext(c, id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if account.Status == "Suspended" {
		log.Errorf("deleteService: Account %s is suspended", account.Email)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if account.Role != "Admin" && account.Role != "Owner" {
		log.Errorf("deleteService: You must be an admin to delete a service")
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	err = core.DeleteService(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to remove client")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func readServices(c *gin.Context) {
	value, exists := c.Get("oauth2Token")
	if !exists {
		c.AbortWithStatus(401)
		return
	}
	oauth2Token := value.(*oauth2.Token)
	oauth2Client := c.MustGet("oauth2Client").(model.Authentication)
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

	services, err := core.ReadServices(user.Email)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to list clients")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, services)
}
