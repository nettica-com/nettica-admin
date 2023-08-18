package client

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	auth "github.com/nettica-com/nettica-admin/auth"
	core "github.com/nettica-com/nettica-admin/core"
	model "github.com/nettica-com/nettica-admin/model"
	util "github.com/nettica-com/nettica-admin/util"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

//var statusCache *cache.Cache

// ApplyRoutes applies router to gin Router
func ApplyRoutes(r *gin.RouterGroup) {
	g := r.Group("/device")
	{

		g.POST("", createDevice)
		g.GET("/:id", readDevice)
		g.PATCH("/:id", updateDevice)
		g.DELETE("/:id", deleteDevice)
		g.GET("", readDevices)
		//		g.GET("/:id/config", configDevice)
		g.GET("/:id/status", statusDevice)
		//		g.GET("/:id/email", emailDevice)
	}

	//	statusCache = cache.New(1*time.Minute, 10*time.Minute)
}

func createDevice(c *gin.Context) {
	var data model.Device

	if err := c.ShouldBindJSON(&data); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to bind")
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	a := util.GetCleanAuthToken(c)
	log.Infof("%v", a)
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
	if data.AccountID == "" {
		data.AccountID = user.AccountID
	}

	client, err := core.CreateDevice(&data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to create client")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, client)
}

func readDevice(c *gin.Context) {
	id := c.Param("id")

	client, err := core.ReadDevice(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read client")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, client)
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

	apikey := c.Request.Header.Get("X-API-KEY")

	if apikey != "" {

		device, err := core.ReadDevice(id)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to read client config")
			c.AbortWithStatus(http.StatusInternalServerError)
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

	client, err := core.UpdateDevice(id, &data, false)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to update device")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, client)
}

func deleteDevice(c *gin.Context) {
	id := c.Param("id")
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

	log.Infof("User %s deleted device %s", user.Name, id)

	err = core.DeleteDevice(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to remove client")
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

	/*
		m, _ := statusCache.Get(id)
		if m != nil {
			msg := m.(model.Message)
			authorized := false

			for _, config := range msg.Config {
				for _, net := range config.Devices {
					if net.DeviceGroup == id && net.APIKey == apikey {
						authorized = true
						break
					}
				}
			}
			if !authorized {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			c.JSON(http.StatusOK, m)
			return
		}
	*/

	device, err := core.ReadDevice(deviceId)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read client config")
		c.AbortWithStatus(http.StatusInternalServerError)
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
				device.LastSeen = time.Now()
				device2 := device
				// update device from id with new last seen
				go func() {
					_, err = core.UpdateDevice(device.Id, device2, true)
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

	//	statusCache.Set(id, msg, 0)
}

/*
func configDevice(c *gin.Context) {

	formatQr := c.DefaultQuery("qrcode", "false")
	zipcode := c.DefaultQuery("zip", "false")

	data, net, err := core.ReadDeviceConfig(c.Param("id"))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read device config")
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	sdata := string(data)

	if zipcode == "false" {
		c.Writer.Header().Set("Content-Type", "application/zip")
		c.Writer.Header().Set("Content-Disposition", "attachment; filename="+*net+".zip")
		w := zip.NewWriter(c.Writer)
		defer w.Close()
		// Make a zip file with the config file
		f, err := w.Create(*net + ".conf")
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to create zip file")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		_, err = f.Write(data)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to write zip file")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		return
	}

	if formatQr == "false" {
		// return config as txt file
		c.Header("Content-Disposition", "attachment; filename=nettica.conf")
		c.Data(http.StatusOK, "application/config", data)
		return
	}
	// return config as png qrcode
	png, err := qrcode.Encode(sdata, qrcode.Medium, 250)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to create qrcode")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "image/png", png)

	return
}

func emailDevice(c *gin.Context) {
	id := c.Param("id")

	err := core.EmailDevice(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to send email to client")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
*/
