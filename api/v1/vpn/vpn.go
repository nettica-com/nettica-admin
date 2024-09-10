package client

import (
	"archive/zip"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	core "github.com/nettica-com/nettica-admin/core"
	model "github.com/nettica-com/nettica-admin/model"
	log "github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
)

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
	}

}

// CreateVPN creates a new VPN for a device
// @Summary Create a new VPN for a device
// @Description Create a new VPN for a device
// @Tags vpn
// @Security apiKey
// @Accept  json
// @Produce  json
// @Param vpn body model.VPN true "VPN"
// @Success 200 {object} model.VPN
// @Router /vpn [post]
func createVPN(c *gin.Context) {
	var data model.VPN

	if err := c.ShouldBindJSON(&data); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to bind")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	network, err := core.ReadNet(data.NetId)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read network")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	account, _, err := core.AuthFromContext(c, network.Id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account from context")
		return
	}

	data.CreatedBy = account.Email
	data.UpdatedBy = account.Email

	if data.AccountID == "" {
		data.AccountID = network.AccountID
	}

	if data.Current != nil && data.Current.Endpoint != "" && account.Role != "Admin" && account.Role != "Owner" {
		// check the policy of the network to see if the endpoint is allowed
		if !network.Policies.UserEndpoints {
			log.Infof("User %s tried to set endpoint %s for network %s", account.Email, data.Current.Endpoint, data.NetId)
			c.JSON(http.StatusForbidden, gin.H{"error": "Users cannot create endpoints for this network"})
			return
		}
	}

	vpn, err := core.CreateVPN(&data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to create client")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	vpns, err := core.ReadVPN2("netid", vpn.NetId)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read vpns")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, v := range vpns {
		// flush the cache for this vpn
		core.FlushCache(v.DeviceID)
	}

	c.JSON(http.StatusOK, vpn)
}

// ReadVPN reads a VPN
// @Summary Read a VPN
// @Description Read a VPN
// @Tags vpn
// @Produce  json
// @Param id path string true "VPN ID"
// @Security apiKey
// @Success 200 {object} model.VPN
// @Router /vpn/{id} [get]
func readVPN(c *gin.Context) {
	id := c.Param("id")

	account, v, err := core.AuthFromContext(c, id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account from context")
		return
	}
	vpn := v.(*model.VPN)

	authorized := false

	apikey := c.Request.Header.Get("X-API-KEY")

	if apikey != "" && strings.HasPrefix(apikey, "device-api-") {

		device, err := core.ReadDevice(vpn.DeviceID)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to read client config")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if device.ApiKey == apikey {
			authorized = true
		}

		if !authorized {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

	}

	if !authorized && account == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	if account != nil && account.Status == "Suspended" {
		log.Errorf("readVPN: account %s is suspended", account.Email)
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is suspended"})
		return
	}

	c.JSON(http.StatusOK, vpn)
}

// UpdateVPN updates a VPN
// @Summary Update a VPN
// @Description Update a VPN
// @Tags vpn
// @Security apiKey
// @Accept  json
// @Produce  json
// @Param id path string true "VPN ID"
// @Param vpn body model.VPN true "VPN"
// @Success 200 {object} model.VPN
// @Router /vpn/{id} [patch]
func updateVPN(c *gin.Context) {
	var data model.VPN
	id := c.Param("id")
	if id == "" {
		log.Error("vpnid cannot be empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "vpnid cannot be empty"})
		return
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
	vpn := v.(*model.VPN)

	authorized := false

	apikey := c.Request.Header.Get("X-API-KEY")

	if apikey != "" && strings.HasPrefix(apikey, "device-api-") {

		device, err := core.ReadDevice(vpn.DeviceID)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to read client config")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if device.ApiKey == apikey {
			authorized = true
		}

		if !authorized {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		data.UpdatedBy = device.Name

	} else {

		if vpn.CreatedBy == account.Email || account.Role == "Admin" || account.Role == "Owner" {
			authorized = true
		}

		if !authorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "You cannot update this VPN"})
			return
		}

		data.UpdatedBy = account.Email
	}

	// this is to allow the logic below to work from device updates without crashing.
	// the device has effective rights of a user so it will not be allowed to change
	// from a client to endpoint, or change the allowedIPs
	if account == nil {
		account = &model.Account{}
	}

	if data.Current.Endpoint != "" && vpn.Current.Endpoint == "" && account.Role != "Admin" && account.Role != "Owner" {
		// check the policy of the network to see if the endpoint is allowed
		network, err := core.ReadNet(vpn.NetId)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to read network")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !network.Policies.UserEndpoints {
			log.Infof("User %s tried to set endpoint %s for network %s", account.Email, data.Current.Endpoint, data.NetId)
			c.JSON(http.StatusForbidden, gin.H{"error": "User Endpoints are disabled for this network"})
			return
		}
	}

	// do not allow changes to the AllowedIPs unless it's an endpoint
	// admins are allowed to do this, for example, extending an AWS subnet through a relay
	// this code is in place to prevent users from breaking the VPN accidentally (or on purpose)
	if data.Current.Endpoint == "" && !CompareArrays(vpn.Current.AllowedIPs, data.Current.AllowedIPs) && account.Role != "Admin" && account.Role != "Owner" {
		log.Infof("User %s tried to change AllowedIPs for vpn %s (%v)", account.Email, id, data.Current.AllowedIPs)
		c.JSON(http.StatusForbidden, gin.H{"error": "You cannot change AllowedIPs"})
		return
	}

	result, err := core.UpdateVPN(id, &data, false)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to update vpn")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	vpns, err := core.ReadVPN2("netid", result.NetId)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read vpns")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, v := range vpns {
		// flush the cache for this vpn
		core.FlushCache(v.DeviceID)
	}

	c.JSON(http.StatusOK, result)
}

func CompareArrays(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, x := range a {
		found := false
		for _, y := range b {
			if x == y {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// DeleteVPN deletes a VPN
// @Summary Delete a VPN
// @Description Delete a VPN
// @Tags vpn
// @Security apiKey
// @Success 200 {object} string "OK"
// @Param id path string true "VPN ID"
// @Router /vpn/{id} [delete]
func deleteVPN(c *gin.Context) {
	id := c.Param("id")

	account, v, err := core.AuthFromContext(c, id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account from context")
		return
	}
	vpn := v.(*model.VPN)

	apikey := c.Request.Header.Get("X-API-KEY")

	if apikey != "" && strings.HasPrefix(apikey, "device-api-") {

		device, err := core.ReadDeviceByApiKey(apikey)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to read client config")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

		log.Infof("Device %s deleted VPN %s", device.Name, id)

	} else {

		if vpn.CreatedBy != account.Email && account.Role != "Admin" && account.Role != "Owner" {
			log.Infof("User %s is not an admin of %s", account.Email, account.Id)
			c.JSON(http.StatusForbidden, gin.H{"error": "You cannot delete this VPN"})
			return
		}

		log.Infof("User %s deleted vpn %s", account.Email, id)
	}

	vpns, err := core.ReadVPN2("netid", vpn.NetId)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read vpns")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = core.DeleteVPN(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to remove client")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, v := range vpns {
		// flush the cache for this vpn

		core.FlushCache(v.DeviceID)
	}

	c.JSON(http.StatusOK, gin.H{})
}

// ReadVPNs reads all VPNs
// @Summary Read all VPNs
// @Description Read all VPNs
// @Tags vpn
// @Security apiKey
// @Produce  json
// @Success 200 {object} []model.VPN
// @Router /vpn [get]
func readVPNs(c *gin.Context) {

	account, _, err := core.AuthFromContext(c, "")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account from context")
		return
	}

	clients, err := core.ReadVPNsForUser(account.Email)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to list clients")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, clients)
}

// ConfigVPN returns the wireguard configuration file for a VPN in a .zip file
// @Summary Get VPN config
// @Description Get VPN config
// @Tags vpn
// @Security apiKey
// @Produce  application/zip
// @Param id path string true "VPN ID"
// @Router /vpn/{id}/config [get]
// @Success 200 {array} byte
func configVPN(c *gin.Context) {

	id := c.Param("id")

	if id == "" {
		log.Error("vpnid cannot be empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "vpnid cannot be empty"})
	}

	account, _, err := core.AuthFromContext(c, id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account from context")
		return
	}

	if account.Status == "Suspended" {
		log.Errorf("configVPN: account %s is suspended", account.Email)
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is suspended"})
		return
	}

	formatQr := c.DefaultQuery("qrcode", "false")
	zipcode := c.DefaultQuery("zip", "false")

	data, net, err := core.ReadVPNConfig(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read vpn config")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		_, err = f.Write(data)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to write zip file")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "image/png", png)

}
