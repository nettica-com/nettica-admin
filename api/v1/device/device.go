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
	"github.com/nettica-com/nettica-admin/push"
	log "github.com/sirupsen/logrus"
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
// @Security apiKey
// @Accept  json
// @Produce  json
// @Success 200 {object} model.Device
// @Failure 400 {object} error
// @Failure 401 {object} error
// @Failure 403 {object} error
// @Failure 422 {object} error
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

	account, _, err := core.AuthFromContext(c, "")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account from context")
		return
	}

	data.CreatedBy = account.Email
	data.UpdatedBy = account.Email

	if !strings.HasPrefix(data.AccountID, "account-") {
		data.AccountID = account.Parent
	}

	if core.EnforceLimits() {
		limit, err := core.ReadLimits(data.AccountID)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to read limits")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		devices, err := core.ReadDevicesForAccount(data.AccountID)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to read devices")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if limit.DevicesLimitReached(len(devices)) {
			log.Infof("createDevice: %s has reached the device limit", account.Email)
			c.JSON(http.StatusForbidden, gin.H{"error": "Device limit reached"})
			return
		}
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
// @Security apiKey
// @Produce  json
// @Param id path string true "Device ID"
// @Success 200 {object} model.Device
// @Failure 400 {object} error
// @Failure 403 {object} error
// @Router /device/{id} [get]
func readDevice(c *gin.Context) {
	id := c.Param("id")

	account, device, err := core.AuthFromContext(c, id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account from context")
		return
	}

	if account != nil && account.Status == "Suspended" {
		log.Infof("Account %s is suspended", account.Email)
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is suspended"})
		return
	}

	c.JSON(http.StatusOK, device)
}

// UpdateDevice updates a device
// @Summary Update a device
// @Description Update a device
// @Tags devices
// @Security apiKey
// @Accept  json
// @Produce  json
// @Param id path string true "Device ID"
// @Param device body model.Device true "Device"
// @Success 200 {object} model.Device
// @Failure 400 {object} error
// @Failure 401 {object} error
// @Failure 403 {object} error
// @Failure 422 {object} error
// @Router /device/{id} [patch]
func updateDevice(c *gin.Context) {
	var data model.Device
	id := c.Param("id")
	if id == "" {
		log.Error("deviceid cannot be empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "devceid cannot be empty"})
	}

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
		}).Error("failed to get account from context")
		return
	}
	device := v.(*model.Device)

	apikey := c.Request.Header.Get("X-API-KEY")

	if apikey != "" && strings.HasPrefix(apikey, "device-api-") {

		authorized := false

		if device.ApiKey == apikey {
			authorized = true
		}

		if !authorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "You cannot update this device"})
			return
		}
		data.UpdatedBy = device.Name

	} else {

		account, err := core.GetAccount(account.Email, device.AccountID)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to read account")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			c.JSON(http.StatusForbidden, gin.H{"error": "You cannot update this device"})
			return
		}

		data.UpdatedBy = account.Email
	}

	client, err := core.UpdateDevice(id, &data, false)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to update device")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	core.FlushCache(id)
	if data.Push != "" {
		push.SendPushNotification(device.Push, "Device Updated", "Device "+device.Name+" has been updated")
	}
	c.JSON(http.StatusOK, client)
}

// DeleteDevice deletes a device
// @Summary Delete a device
// @Description Delete a device
// @Tags devices
// @Security apiKey
// @Produce  json
// @Param id path string true "Device ID"
// @Success 200 {object} string "OK"
// @Failure 400 {object} error
// @Failure 401 {object} error
// @Failure 403 {object} error
// @Failure 404 {object} error
// @Router /device/{id} [delete]
func deleteDevice(c *gin.Context) {
	id := c.Param("id")

	account, v, err := core.AuthFromContext(c, id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account from context")
		return
	}
	device := v.(*model.Device)

	apikey := c.Request.Header.Get("X-API-KEY")

	if account != nil {
		if account.Role != "Admin" && account.Role != "Owner" && device.CreatedBy != account.Email {
			c.JSON(http.StatusForbidden, gin.H{"error": "You cannot delete this device"})
			return
		}
		log.Infof("User %s deleted device %s", account.Email, id)

	} else if device != nil && device.Id == id && device.ApiKey == apikey {
		log.Infof("Device %s deleted itself", device.Id)
	}

	err = core.DeleteDevice(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to delete device")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

// ReadDevices reads all devices
// @Summary Read all devices
// @Description Read all devices
// @Tags devices
// @Security apiKey
// @Produce  json
// @Success 200 {object} model.Device
// @Failure 400 {object} error
// @Failure 401 {object} error
// @Failure 422 {object} error
// @Router /device [get]
func readDevices(c *gin.Context) {

	account, _, err := core.AuthFromContext(c, "")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account from context")
		return
	}
	clients, err := core.ReadDevicesForUser(account.Email)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to list clients")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, clients)
}

// getName returns the name of the server
func getName() string {
	name := os.Getenv("SERVER")

	name = strings.TrimPrefix(name, "https://")
	name = strings.TrimPrefix(name, "http://")

	return name
}

// StatusDevice reads state for a device
// @Summary Read state for a device
// @Description Read state for a device
// @Tags devices
// @Security apiKey
// @Produce  json
// @Param id path string true "Device ID"
// @Success 200 {object} model.Message
// @Failure 400 {object} error
// @Failure 401 {object} error
// @Failure 404 {object} error
// @Router /device/{id}/status [get]
func statusDevice(c *gin.Context) {

	deviceId := c.Param("id")

	if deviceId == "" || deviceId == "device-id-" {
		log.Error("deviceid cannot be empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "deviceid cannot be empty"})
		return
	}

	apikey := c.Request.Header.Get("X-API-KEY")
	etag := c.Request.Header.Get("If-None-Match")
	ip := c.ClientIP()

	device, err := core.ReadDevice(deviceId)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read client config")
		if err.Error() == "mongo: no documents in result" {
			c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	authorized := false

	if device.ApiKey == apikey {
		authorized = true
		if !device.Registered {
			device.Registered = true
			device.EZCode = ""
			_, err = core.UpdateDevice(device.Id, device, true)
			if err != nil {
				log.Error(err)
			}
		}
	}

	if !authorized && !device.Registered && (device.InstanceID != "" || device.EZCode != "") {
		authorized = true
	}

	if !authorized {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	m, _ := core.GetCache(device.Id)
	if m != nil {
		e := m.(string)
		if e == etag {
			c.AbortWithStatus(http.StatusNotModified)
			go func() {
				now := time.Now()
				device.LastSeen = &now
				_, err = core.UpdateDevice(device.Id, device, true)
				if err != nil {
					log.Error(err)
				}
			}()

			return
		}
	}

	nets, err := core.ReadVPN2("deviceid", device.Id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read client config")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var msg model.Message
	hconfig := make([]model.VPNConfig, len(nets))

	msg.Version = "3.0"
	msg.Name = getName()
	msg.Id = device.Id
	msg.Device = device
	msg.Config = hconfig

	for i, net := range nets {
		clients, err := core.ReadVPN2("netid", net.NetId)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to list clients")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		network, err := core.ReadNet(net.NetId)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to read network")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		onlyEndpoints := false
		if network.Policies.OnlyEndpoints {
			onlyEndpoints = true
		}

		isEndpoint := false
		if net.Current.Endpoint != "" {
			isEndpoint = true
		}

		msg.Config[i] = model.VPNConfig{}
		msg.Config[i].NetName = network.NetName
		msg.Config[i].NetId = network.Id
		msg.Config[i].Description = network.Description

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
					if client.DeviceID == device.Id {
						isIngress = true
					}
				}
				if client.Role == "Egress" {
					egress = client
					if client.DeviceID == device.Id {
						isEgress = true
					}
				}
			} else {
				log.Errorf("internal error")
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		if isEgress && hasIngress {
			public_subnets := []string{"0.0.0.0/5", "8.0.0.0/7", "11.0.0.0/8", "12.0.0.0/6", "16.0.0.0/4", "32.0.0.0/3", "64.0.0.0/3",
				"96.0.0.0/4", "112.0.0.0/5", "120.0.0.0/6", "124.0.0.0/7", "126.0.0.0/8", "128.0.0.0/3", "160.0.0.0/5", "168.0.0.0/6",
				"172.0.0.0/12", "172.32.0.0/11", "172.64.0.0/10", "172.128.0.0/9", "173.0.0.0/8", "174.0.0.0/7", "176.0.0.0/4", "192.0.0.0/9", "192.128.0.0/11",
				"192.160.0.0/13", "192.169.0.0/16", "192.170.0.0/15", "192.172.0.0/14", "192.176.0.0/12", "192.192.0.0/10", "193.0.0.0/8", "194.0.0.0/7",
				"196.0.0.0/6", "200.0.0.0/5", "208.0.0.0/4", "::/1", "8000::/2", "c000::/3", "e000::/4", "f000::/5", "f800::/6", "fe00::/9", "fe80::/10", "ff00::/8"}

			/*
				   AllowedIPs = 0.0.0.0/5, 8.0.0.0/7, 11.0.0.0/8, 12.0.0.0/6, 16.0.0.0/4, 32.0.0.0/3, 64.0.0.0/3, 96.0.0.0/4, 112.0.0.0/5, 120.0.0.0/6, 124.0.0.0/7, 126.0.0.0/8,
				   			 128.0.0.0/3, 160.0.0.0/5, 168.0.0.0/6, 172.0.0.0/12, 172.32.0.0/11, 172.64.0.0/10, 172.128.0.0/9, 173.0.0.0/8, 174.0.0.0/7, 176.0.0.0/4, 192.0.0.0/9,
				   			 192.128.0.0/11, 192.160.0.0/13, 192.169.0.0/16, 192.170.0.0/15, 192.172.0.0/14, 192.176.0.0/12, 192.192.0.0/10, 193.0.0.0/8, 194.0.0.0/7, 196.0.0.0/6,
				   			 200.0.0.0/5, 208.0.0.0/4, 224.0.0.0/3, ::/1, 8000::/2, c000::/3, e000::/4, f000::/5, f800::/6, fe00::/9, fe80::/10, ff00::/8
					Excludes 127.0.0.0/8, 10.0.0.0/8, 192.168.0.0/16, 172.16.0.0/12, fc00::/7, fec0::/10
			*/
			// If this is the egress device, only return the ingress and egress devices
			// and remove the 0.0.0.0/0 from allowedIPs on the ingress device
			allowed := ingress.Current.AllowedIPs
			// Remove 0.0.0.0/0 from the allowed IPs
			for x, ip := range allowed {
				for _, pub := range public_subnets {
					if ip == pub {
						allowed = append(allowed[:x], allowed[x+1:]...)
						break
					}
				}
				if ip == "0.0.0.0/0" {
					allowed = append(allowed[:x], allowed[x+1:]...)
				}
			}
			for x, ip := range allowed {
				if ip == "::/0" {
					allowed = append(allowed[:x], allowed[x+1:]...)
				}
			}

			msg.Config[i].VPNs = make([]model.VPN, 2)
			msg.Config[i].VPNs[0] = *ingress
			msg.Config[i].VPNs[0].Current.AllowedIPs = allowed
			msg.Config[i].VPNs[1] = *egress
		}

		for _, client := range clients {
			isClient := false

			// If this config isn't explicitly for this device, remove the private key
			// from the results
			if client.DeviceID != device.Id {

				client.Current.PrivateKey = ""
				client.Default = nil // &model.Settings{}
				client.Current.PreUp = ""
				client.Current.PostUp = ""
				client.Current.PreDown = ""
				client.Current.PostDown = ""

			} else {
				// This is the current client
				device2 := *device
				isClient = true
				if client.Current.SyncEndpoint && client.Role != "Ingress" {
					// If this client has syncEndpoint on see if the ip has changed
					update := false
					if client.Current.Endpoint == "" {
						// do nothing.  user must prove they can set an endpoint
					} else {
						if client.Current.ListenPort == 0 {
							client.Current.ListenPort = 51820
						}
						if client.Current.Endpoint != ip+":"+fmt.Sprintf("%d", client.Current.ListenPort) {
							client.Current.Endpoint = ip + ":" + fmt.Sprintf("%d", client.Current.ListenPort)
							update = true
						}
					}
					if update {
						client.UpdatedBy = device.Name
						core.UpdateVPN(client.Id, client, true)
					}
				}
				//
				// update device from id with new last seen
				go func() {
					now := time.Now()
					device2.LastSeen = &now
					_, err = core.UpdateDevice(device2.Id, &device2, true)
					if err != nil {
						log.Error(err)
					}
				}()
			}
			device.LastSeen = nil

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

			if onlyEndpoints && !isClient && !isEndpoint && client.Current.Endpoint == "" {
				// skip (network policy says clients can't see each other)
			} else {
				msg.Config[i].VPNs = append(msg.Config[i].VPNs, *client)
			}
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
		log.Infof("Etag for %s is %s", device.Id, md5)
		c.JSON(http.StatusOK, msg)
	}

	core.SetCache(device.Id, md5)
}
