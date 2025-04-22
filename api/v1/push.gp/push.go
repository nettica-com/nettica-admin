package push

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nettica-com/nettica-admin/core"
	model "github.com/nettica-com/nettica-admin/model"
	"github.com/nettica-com/nettica-admin/template"
	"github.com/nettica-com/nettica-admin/util"
	log "github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

// ApplyRoutes applies router to gin Router
func ApplyRoutes(r *gin.RouterGroup) {
	g := r.Group("/push")
	{
		g.POST("", registerPush)
		g.GET("/:id", readPush)
		g.POST("/:id", sendPush)
		g.GET("", reloadPush)
		g.DELETE("/:id", unregisterPush)
	}
}

func registerPush(c *gin.Context) {
	var pusher model.Pusher
	if err := c.ShouldBindJSON(&pusher); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("push: register: failed to bind json")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to bind json"})
		return
	}

	if pusher.Server == "" {
		log.Error("push: register: server is empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "server is required"})
		return
	}

	if pusher.Host == "" {
		log.Error("push: register: host is empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "host is required"})
		return
	}

	if pusher.Id == "" {
		id, err := util.RandomString(12)
		if err != nil {
			log.Error("push: register: failed to generate id")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate id"})
			return
		}
		pusher.Id = "push-" + id
	}

	if pusher.ApiKey == "" {
		apiKey, err := util.RandomString(32)
		if err != nil {
			log.Error("push: register: failed to generate api key")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate api key"})
			return
		}
		pusher.ApiKey = "push-api-" + apiKey
	}

	now := time.Now()
	pusher.Created = &now
	pusher.Updated = &now
	pusher.Version = "1.0"

	if err := pusher.IsValid(); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("push: register: failed to validate pusher")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to validate pusher"})
		return
	}

	err := core.PM.Add(&pusher)
	if err != nil {
		log.Errorf("push: register: failed to add pusher: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add pusher"})
		return
	}

	err = pushEmail(&pusher)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("push: register: failed to send email")
	}

	c.JSON(http.StatusOK, pusher)
}

func reloadPush(c *gin.Context) {

	// verify the call came from 127.0.0.1
	if c.ClientIP() != "127.0.0.1" {
		log.Errorf("push: reload push called from %s", c.ClientIP())
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	log.Info("push: reload push")
	core.PM.Load()

	c.JSON(http.StatusOK, gin.H{"message": "push reloaded"})
}

func readPush(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		log.Error("push: read: id is empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if !validate(id) {
		log.Error("push: read: id is not valid")
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is not valid"})
		return
	}

	apiKey := c.Request.Header.Get("X-API-KEY")
	if apiKey == "" {
		log.Error("push: read: api key is empty")
		c.JSON(http.StatusForbidden, gin.H{"error": "api key is required"})
		return
	}

	pp, err := core.PM.Get(id)
	if err != nil {
		log.Errorf("push: read: failed to get pusher: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "failed to get pusher"})
		return
	}
	if pp == nil {
		log.Errorf("push: read: pusher %s not found", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "pusher not found"})
		return
	}

	if pp.ApiKey != apiKey {
		log.Errorf("push: read: api key mismatch %s != %s", pp.ApiKey, apiKey)
		c.JSON(http.StatusForbidden, gin.H{"error": "api key mismatch"})
		return
	}

	c.JSON(http.StatusOK, pp)
}

func sendPush(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		log.Error("push: send: id is empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var push model.Push
	if err := c.ShouldBindJSON(&push); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("push: send: failed to bind json")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to bind json"})
		return
	}

	if err := push.IsValid(); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("push: send: failed to validate push")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to validate push"})
		return
	}

	if !validate(id) {
		log.Error("push: send: id is not valid")
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is not valid"})
		return
	}

	if push.Id != id {
		log.Error("push: send: id mismatch")
		c.JSON(http.StatusBadRequest, gin.H{"error": "id mismatch"})
		return
	}

	apiKey := c.Request.Header.Get("X-API-KEY")
	if apiKey == "" {
		log.Error("push: send: api key is empty")
		c.JSON(http.StatusForbidden, gin.H{"error": "api key is required"})
		return
	}

	if push.ApiKey != apiKey {
		log.Error("push: send: api key mismatch")
		c.JSON(http.StatusForbidden, gin.H{"error": "api key mismatch"})
		return
	}

	pp, err := core.PM.Get(id)
	if err != nil {
		log.Errorf("push: send: failed to get pusher: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "failed to get pusher"})
		return
	}

	if pp == nil {
		log.Errorf("push: send: pusher %s not found", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "pusher not found"})
		return
	}

	if pp.ApiKey != push.ApiKey {
		log.Errorf("push: send: api key mismatch %s != %s", pp.ApiKey, push.ApiKey)
		c.JSON(http.StatusForbidden, gin.H{"error": "api key mismatch"})
		return
	}

	err = core.Push.SendPushNotification(push.Token, push.Title, push.Message)
	if err != nil {
		log.Errorf("push: send: failed to send push: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "failed to send push"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func unregisterPush(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		log.Error("push: unregister: id is empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if !validate(id) {
		log.Error("push: unregister: id is not valid")
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is not valid"})
		return
	}

	apiKey := c.Request.Header.Get("X-API-KEY")
	if apiKey == "" {
		log.Error("push: unregister: api key is empty")
		c.JSON(http.StatusForbidden, gin.H{"error": "api key is required"})
		return
	}

	pp, err := core.PM.Get(id)
	if err != nil {
		log.Errorf("push: unregister: failed to get pusher: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "failed to get pusher"})
		return
	}
	if pp == nil {
		log.Errorf("push: unregister: pusher %s not found", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "pusher not found"})
		return
	}

	if pp.ApiKey != apiKey {
		log.Errorf("push: unregister: api key mismatch %s != %s", pp.ApiKey, apiKey)
		c.JSON(http.StatusForbidden, gin.H{"error": "api key mismatch"})
		return
	}

	err = core.PM.Remove(id)
	if err != nil {
		log.Errorf("push: unregister: failed to remove pusher: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove pusher"})
	}

	c.JSON(http.StatusOK, gin.H{"message": "push deleted"})
}

func validate(s string) bool {

	return !strings.ContainsAny(s, "${}()\"")

}

func pushEmail(p *model.Pusher) error {
	// get email body
	emailBody, err := template.PushEmail(p)
	if err != nil {
		return err
	}

	// port to int
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return err
	}

	d := gomail.NewDialer(os.Getenv("SMTP_HOST"), port, os.Getenv("SMTP_USERNAME"), os.Getenv("SMTP_PASSWORD"))
	s, err := d.Dial()
	if err != nil {
		return err
	}
	m := gomail.NewMessage()

	m.SetHeader("From", os.Getenv("SMTP_FROM"))
	m.SetAddressHeader("To", "info@nettica.com", "Nettica")
	m.SetHeader("Subject", "Nettica Push Registration: "+p.Server)
	m.SetBody("text/html", string(emailBody))

	err = gomail.Send(s, m)
	if err != nil {
		return err
	}

	return nil
}
