package service

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
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

	services, err := core.ReadServiceHost(serviceGroup)
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

	//	statusCache.Set(id, msg, 0)
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

	service, err := core.ReadService(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read client")
		c.AbortWithStatus(http.StatusInternalServerError)
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

	apikey := c.Request.Header.Get("X-API-KEY")

	if apikey != "" {

		service, err := core.ReadService(id)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to read client config")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		authorized := false

		if service.ApiKey == apikey {
			authorized = true
		}

		if !authorized {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	} else {
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
		data.UpdatedBy = user.Email
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

	err := core.DeleteService(id)
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
