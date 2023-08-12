package client

import (
	"archive/zip"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	auth "github.com/nettica-com/nettica-admin/auth"
	core "github.com/nettica-com/nettica-admin/core"
	model "github.com/nettica-com/nettica-admin/model"
	util "github.com/nettica-com/nettica-admin/util"
	log "github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
	"golang.org/x/oauth2"
)

//var statusCache *cache.Cache

// ApplyRoutes applies router to gin Router
func ApplyRoutes(r *gin.RouterGroup) {
	g := r.Group("/vpn")
	{

		g.POST("", createVPN)
		g.GET("/:id", readVPN)
		g.PATCH("/:id", updateVPN)
		g.DELETE("/:id", deleteVPN)
		g.GET("", readVPNs)
		g.GET("/:id/config", configVPN)
		g.GET("/:id/status", statusVPN)
	}

	//	statusCache = cache.New(1*time.Minute, 10*time.Minute)
}

func createVPN(c *gin.Context) {
	var data model.VPN

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
	if data.AccountId == "" {
		data.AccountId = user.AccountId
	}

	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to generate state random string")
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	client, err := core.CreateVPN(&data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to create client")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, client)
}

func readVPN(c *gin.Context) {
	id := c.Param("id")

	client, err := core.ReadVPN(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read client")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, client)
}

func updateVPN(c *gin.Context) {
	var data model.VPN
	id := c.Param("id")
	if id == "" {
		log.Error("vpnid cannot be empty")
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

		device, err := core.ReadDevice(data.DeviceId)
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

	client, err := core.UpdateVPN(id, &data, false)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to update vpn")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, client)
}

func deleteVPN(c *gin.Context) {
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

	log.Infof("User %s deleted vpn %s", user.Name, id)

	err = core.DeleteVPN(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to remove client")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func readVPNs(c *gin.Context) {
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
	clients, err := core.ReadVPNsForUser(user.Email)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to list clients")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, clients)
}

func statusVPN(c *gin.Context) {

	//	id := c.Param("id")
	if c.Param("id") == "" {
		log.Error("DeviceId cannot be empty")
		c.AbortWithStatus(http.StatusInternalServerError)
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
				for _, net := range config.VPNs {
					if net.VPNGroup == id && net.APIKey == apikey {
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
		}).Error("failed to get device")
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

	nets, err := core.ReadVPN2("deviceId", deviceId)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read client config")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var msg model.Message
	hconfig := make([]model.VPNConfig, len(nets))

	msg.Id = c.Param("id")
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
					if client.DeviceId == deviceId {
						isIngress = true
					}
				}
				if client.Role == "Egress" {
					egress = client
					if client.DeviceId == deviceId {
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
			// If this is the egress vpn, only return the ingress and egress vpns
			// and remove the 0.0.0.0/0 from allowedIPs on the ingress vpn
			allowed := ingress.Current.AllowedIPs
			// Remove 0.0.0.0/0 from the allowed IPs
			for x, ip := range allowed {
				if ip == "0.0.0.0/0" {
					allowed = append(allowed[:x], allowed[x+1:]...)
					break
				}
			}
			msg.Config[i].Vpns = make([]model.VPN, 2)
			msg.Config[i].Vpns[0] = *ingress
			msg.Config[i].Vpns[0].Current.AllowedIPs = allowed
			msg.Config[i].Vpns[1] = *egress
		}

		for _, client := range clients {
			// If this config isn't explicitly for this vpn, remove the private key
			// from the results
			if client.DeviceId != deviceId {
				client.Current.PrivateKey = ""
			} else {
				//				client.LastSeen = time.Now()
				//				client2 := *client
				// update vpn from id with new last seen
				//				go func() {
				//					_, err = core.UpdateVPN(client2.Id, &client2, true)
				//					if err != nil {
				//						log.Error(err)
				//					}
				//				}()
			}
			//			client.LastSeen = time.Time{}

			if isEgress {
				// If this is the egress vpn, only return the ingress and egress vpns
				// (which was done above)
				continue
			}

			if client.Role == "Egress" && hasIngress && !isIngress {
				// VPNs pointing to ingress do not see the egress vpn
				// If it's the ingress itself, it needs to see the egress vpn
				continue
			}

			// If this isn't the egress vpn, or there is
			// only an egress vpn, or if it's neither,
			// include this client in the results

			msg.Config[i].Vpns = append(msg.Config[i].Vpns, *client)
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

func configVPN(c *gin.Context) {

	formatQr := c.DefaultQuery("qrcode", "false")
	zipcode := c.DefaultQuery("zip", "false")

	data, net, err := core.ReadVPNConfig(c.Param("id"))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read vpn config")
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
