package client

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	core "github.com/nettica-com/nettica-admin/core"
	model "github.com/nettica-com/nettica-admin/model"
	util "github.com/nettica-com/nettica-admin/util"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// ApplyRoutes applies router to gin Router
func ApplyRoutes(r *gin.RouterGroup) {
	g := r.Group("/device")
	{

		g.POST("", createDevice)
		g.GET("/:id", readDevice)
		g.PATCH("/:id", updateDevice)
		g.DELETE("/:id", deleteDevice)
		g.GET("", readDevices)
		g.GET("/:id/status", statusDevice)
	}

}

// CreateDevice creates a device
// @Summary Create a device
// @Description Create a device
// @Tags devices
// @Security ApiKeyAuth true "X-API-KEY" "device-api-<apikey>"
// @Security OAuth2
// @Accept  json
// @Produce  json
// @Param device body Device true "Device"
// @Success 200 {object} Device
// @Failure 400 {object} Error
// @Failure 422 {object} Error
// @Router /device [post]
func createDevice(c *gin.Context) {
	var data model.Device

	if err := c.ShouldBindJSON(&data); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to bind")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	a := util.GetCleanAuthToken(c)
	log.Infof("%v", a)

	account, _, err := core.GetFromContext(c, "")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account from context")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data.CreatedBy = account.Email
	data.UpdatedBy = account.Email

	if data.AccountID == "" {
		data.AccountID = account.Parent
	}

	client, err := core.CreateDevice(&data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to create client")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, client)
}

// ReadDevice reads a device
// @Summary Read a device
// @Description Read a device
// @Tags devices
// @Security ApiKeyAuth true "X-API-KEY" "device-api-<apikey>"
// @Security OAuth2
// @Produce  json
// @Param id path string true "Device ID"
// @Success 200 {object} Device
// @Failure 400 {object} Error
// @Failure 403 {object} Error
func readDevice(c *gin.Context) {
	id := c.Param("id")

	account, device, err := core.GetFromContext(c, id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account from context")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if account.Status == "Suspended" {
		log.Infof("Account %s is suspended", account.Email)
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is suspended"})
		return
	}

	c.JSON(http.StatusOK, device)
}

func updateDevice(c *gin.Context) {
	var data model.Device
	id := c.Param("id")
	if id == "" {
		log.Error("deviceid cannot be empty")
		c.AbortWithStatus(http.StatusInternalServerError)
	}

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
		}).Error("failed to get account from context")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	device := v.(*model.Device)

	apikey := c.Request.Header.Get("X-API-KEY")

	if apikey != "" && strings.HasPrefix(apikey, "device-api-") {

		authorized := false

		if device.ApiKey == apikey {
			authorized = true
		}
		data.UpdatedBy = device.Name

		if !authorized {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	} else {

		account, err := core.GetAccount(account.Email, device.AccountID)
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
			}).Error("account not found")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		authorized := false

		if device.CreatedBy == account.Email || account.Role == "Admin" || account.Role == "Owner" {
			authorized = true
		}

		if !authorized {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		data.UpdatedBy = account.Email
	}

	client, err := core.UpdateDevice(id, &data, false)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to update device")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	core.FlushCache(id)
	c.JSON(http.StatusOK, client)
}

func deleteDevice(c *gin.Context) {
	id := c.Param("id")

	account, v, err := core.GetFromContext(c, id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account from context")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	device := v.(*model.Device)

	apikey := c.Request.Header.Get("X-API-KEY")

	if apikey != "" {

		log.Infof("Device %s deleted VPN %s", device.Name, id)

	} else {
		log.Infof("User %s deleted device %s", account.Email, id)
	}

	err = core.DeleteDevice(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to delete device")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func readDevices(c *gin.Context) {
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
	clients, err := core.ReadDevicesForUser(user.Email)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to list clients")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, clients)
}

func statusDevice(c *gin.Context) {

	//	id := c.Param("id")
	if c.Param("id") == "" {
		log.Error("deviceid cannot be empty")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	deviceId := c.Param("id")

	apikey := c.Request.Header.Get("X-API-KEY")
	etag := c.Request.Header.Get("If-None-Match")

	device, err := core.ReadDevice(deviceId)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read client config")
		if err.Error() == "mongo: no documents in result" {
			c.AbortWithStatus(http.StatusNotFound)
		} else {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		return
	}

	authorized := false

	if device.ApiKey == apikey {
		authorized = true
	}

	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	m, _ := core.GetCache(deviceId)
	if m != nil {
		e := m.(string)
		if e == etag {
			c.AbortWithStatus(http.StatusNotModified)
			go func() {
				device.LastSeen = time.Now()
				_, err = core.UpdateDevice(device.Id, device, true)
				if err != nil {
					log.Error(err)
				}
			}()

			return
		}
	}

	nets, err := core.ReadVPN2("deviceid", deviceId)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read client config")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var msg model.Message
	hconfig := make([]model.VPNConfig, len(nets))

	msg.Id = deviceId
	msg.Device = device
	msg.Config = hconfig

	for i, net := range nets {
		clients, err := core.ReadVPN2("netid", net.NetId)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to list clients")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		msg.Config[i] = model.VPNConfig{}
		msg.Config[i].NetName = net.NetName
		msg.Config[i].NetId = net.NetId

		hasIngress := false
		ingress := &model.VPN{}
		egress := &model.VPN{}
		isIngress := false
		isEgress := false

		// Check the net to see if it has ingress and egress roles
		for _, client := range clients {
			// They should all match
			if client.NetId == msg.Config[i].NetId {
				if client.Role == "Ingress" {
					hasIngress = true
					ingress = client
					if client.DeviceID == deviceId {
						isIngress = true
					}
				}
				if client.Role == "Egress" {
					egress = client
					if client.DeviceID == deviceId {
						isEgress = true
					}
				}
			} else {
				log.Errorf("internal error")
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		if isEgress && hasIngress {
			// If this is the egress device, only return the ingress and egress devices
			// and remove the 0.0.0.0/0 from allowedIPs on the ingress device
			allowed := ingress.Current.AllowedIPs
			// Remove 0.0.0.0/0 from the allowed IPs
			for x, ip := range allowed {
				if ip == "0.0.0.0/0" {
					allowed = append(allowed[:x], allowed[x+1:]...)
					break
				}
			}
			msg.Config[i].VPNs = make([]model.VPN, 2)
			msg.Config[i].VPNs[0] = *ingress
			msg.Config[i].VPNs[0].Current.AllowedIPs = allowed
			msg.Config[i].VPNs[1] = *egress
		}

		for _, client := range clients {
			// If this config isn't explicitly for this device, remove the private key
			// from the results
			if client.DeviceID != deviceId {
				client.Current.PrivateKey = ""
			} else {
				device2 := *device
				// update device from id with new last seen
				go func() {
					device2.LastSeen = time.Now()
					_, err = core.UpdateDevice(device2.Id, &device2, true)
					if err != nil {
						log.Error(err)
					}
				}()
			}
			device.LastSeen = time.Time{}

			if isEgress {
				// If this is the egress device, only return the ingress and egress devices
				// (which was done above)
				continue
			}

			if client.Role == "Egress" && hasIngress && !isIngress {
				// Devices pointing to ingress do not see the egress device
				// If it's the ingress itself, it needs to see the egress device
				continue
			}

			// If this isn't the egress device, or there is
			// only an egress device, or if it's neither,
			// include this client in the results

			msg.Config[i].VPNs = append(msg.Config[i].VPNs, *client)
		}
	}
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

	core.SetCache(deviceId, md5)
}
