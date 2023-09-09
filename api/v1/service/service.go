package service

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	core "github.com/nettica-com/nettica-admin/core"
	model "github.com/nettica-com/nettica-admin/model"
	log "github.com/sirupsen/logrus"
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	account, _, err := core.AuthFromContext(c, data.AccountID)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account")
		return
	}

	data.CreatedBy = account.Email

	if data.AccountID == "" {
		data.AccountID = account.Id
	}

	if account.Role != "Admin" && account.Role != "Owner" {
		log.Errorf("createService: You must be an admin with credits to create a service")
		c.JSON(http.StatusForbidden, gin.H{"error": "You must be an admin with credits to create a service"})
		return
	}

	client, err := core.CreateService(&data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to create net")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, client)
}

func readService(c *gin.Context) {
	id := c.Param("id")

	account, service, err := core.AuthFromContext(c, id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account")
		return
	}

	if account.Status == "Suspended" {
		log.Errorf("readService: Account %s is suspended", account.Email)
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is suspended"})
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
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	account, v, err := core.AuthFromContext(c, id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account")
		return
	}

	service := v.(*model.Service)

	apikey := c.Request.Header.Get("X-API-KEY")

	if apikey != "" {

		authorized := false

		if service.ApiKey == apikey {
			authorized = true
		}

		if !authorized {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, client)
}

func deleteService(c *gin.Context) {
	id := c.Param("id")

	account, _, err := core.AuthFromContext(c, id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account")
		return
	}

	if account != nil && account.Role != "Admin" && account.Role != "Owner" {
		log.Errorf("deleteService: You must be an admin to delete a service")
		c.JSON(http.StatusForbidden, gin.H{"error": "You must be an admin to delete a service"})
		return
	}

	err = core.DeleteService(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to remove client")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func readServices(c *gin.Context) {

	account, _, err := core.AuthFromContext(c, "")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account")
		return
	}

	services, err := core.ReadServices(account.Email)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to list clients")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, services)
}
