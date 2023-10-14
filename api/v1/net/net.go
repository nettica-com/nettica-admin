package net

import (
	"net/http"

	"github.com/gin-gonic/gin"
	core "github.com/nettica-com/nettica-admin/core"
	model "github.com/nettica-com/nettica-admin/model"
	"github.com/nettica-com/nettica-admin/util"
	log "github.com/sirupsen/logrus"
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

// CreateNet creates a new network
// @Summary Create a new network
// @Description Create a new network
// @tags net
// @Accept  json
// @Produce  json
// @Security apiKey
// @Param net body model.Network true "Network"
// @Success 200 {object} model.Network
// @Failure 400 {object} string
// @Router /net [post]
func createNet(c *gin.Context) {
	var data model.Network

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
		}).Error("failed to get account from context")
		return
	}

	if account.Role != "Admin" && account.Role != "Owner" {
		log.Infof("createNet: user %s is not an admin of %s", account.Email, account.Id)
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not an admin of this account"})
		return
	}

	if account.NetId != "" {
		log.Infof("createNet: user %s cannot create new nets in this account", account.Email)
		c.JSON(http.StatusForbidden, gin.H{"error": "You cannot create new nets in this account"})
		return
	}

	data.CreatedBy = account.Email
	data.UpdatedBy = account.Email

	if data.AccountID == "" {
		data.AccountID = account.Parent
	}

	if core.EnforceLimits() {
		// check if the account has reached the limits
		nets, err := core.ReadNetworksForAccount(data.AccountID)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to read networks")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		limits, err := core.ReadLimits(data.AccountID)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to read limits")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if limits.NetworksLimitReached(len(nets)) {
			log.Infof("createNet: %s has reached the networks limit", account.Email)
			c.JSON(http.StatusForbidden, gin.H{"error": "Network limit reached"})
			return
		}
	}

	client, err := core.CreateNet(&data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to create net")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, client)
}

// ReadNet reads a network
// @Summary Read a network
// @Description Read a network
// @tags net
// @Produce  json
// @Security apiKey
// @Param id path string true "Network ID"
// @Success 200 {object} model.Network
// @Failure 400 {object} string
// @Router /net/{id} [get]
func readNet(c *gin.Context) {
	id := c.Param("id")

	account, net, err := core.AuthFromContext(c, id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account from context")
		return
	}

	if account.Status == "Suspended" {
		log.Infof("readNet: account %s is suspended", account.Email)
		c.JSON(http.StatusForbidden, gin.H{"error": "account is suspended"})
		return
	}

	c.JSON(http.StatusOK, net)
}

// UpdateNet updates a network
// @Summary Update a network
// @Description Update a network
// @tags net
// @Accept  json
// @Produce  json
// @Security apiKey
// @Param id path string true "Network ID"
// @Param net body model.Network true "Network"
// @Success 200 {object} model.Network
// @Failure 400 {object} error
// @Router /net/{id} [patch]
func updateNet(c *gin.Context) {
	var data model.Network
	id := c.Param("id")

	if err := c.ShouldBindJSON(&data); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to bind")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if data.Id != id {
		log.WithFields(log.Fields{
			"id":  id,
			"req": data.Id,
		}).Error("id mismatch")
		c.JSON(http.StatusBadRequest, gin.H{"error": "id mismatch"})
		return
	}

	account, v, err := core.AuthFromContext(c, id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account from context")
		return
	}
	net := v.(*model.Network)

	authorized := false

	if net.CreatedBy == account.Email || account.Role == "Admin" || account.Role == "Owner" {
		authorized = true
	}

	if !authorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this network"})
		return
	}

	data.UpdatedBy = account.Email

	result, err := core.UpdateNet(id, &data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to update network")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updateMTU := false
	updateAddress := false
	updateAllowed := false
	if data.ForceUpdate {
		log.Infof("updateNet: force update for %s %s", data.NetName, id)
		if data.Default.Mtu != net.Default.Mtu {
			log.Infof("updateNet: updateMTU for %s %d", data.NetName, data.Default.Mtu)
			updateMTU = true
		}
		if !util.CompareArrays(data.Default.Address, net.Default.Address) {
			log.Infof("updateNet: updateAddress for %s %v", data.NetName, data.Default.Address)
			updateAddress = true
		}
		if !util.CompareArrays(data.Default.AllowedIPs, net.Default.AllowedIPs) {
			log.Infof("updateNet: updateAllowed for %s %v", data.NetName, data.Default.AllowedIPs)
			updateAllowed = true
		}
	}
	// Clear the device cache for policy changes
	vpns, err := core.ReadVPN2("netid", net.Id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read vpns")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, v := range vpns {
		changed := false
		if updateMTU && (v.Current.Mtu != data.Default.Mtu || v.Default.Mtu != data.Default.Mtu) {
			log.Infof("updateNet: updateMTU for %s %d", v.Id, data.Default.Mtu)
			v.Default.Mtu = data.Default.Mtu
			v.Current.Mtu = data.Default.Mtu
			changed = true
		}
		if updateAddress && !util.CompareArrays(v.Default.Address, data.Default.Address) {
			log.Infof("updateNet: updateAddress for %s %v", v.Id, data.Default.Address)
			v.Default.Address = data.Default.Address
			v.Current.Address = make([]string, 0)
			changed = true
		}
		if updateAllowed && !util.CompareArrays(v.Default.AllowedIPs, data.Default.AllowedIPs) {
			log.Infof("updateNet: updateAllowed for %s %v", v.Id, data.Default.AllowedIPs)
			allowedIPs := make([]string, 0)
			for _, subnet := range v.Default.AllowedIPs {
				for _, cidr := range v.Current.AllowedIPs {
					if !util.IsInCidr(cidr, subnet) {
						allowedIPs = append(allowedIPs, cidr)
					} else if cidr == subnet {
						allowedIPs = append(allowedIPs, data.Default.AllowedIPs...)
					}
				}
			}
			v.Current.AllowedIPs = allowedIPs
			v.Default.AllowedIPs = data.Default.AllowedIPs
			changed = true
		}
		if changed {
			_, err := core.UpdateVPN(v.Id, v, true)
			if err != nil {
				log.Errorf("forceUpdate: failed to update vpn %s %v", v.Id, err)
			}
		}

		// flush the cache for this vpn
		core.FlushCache(v.DeviceID)
	}

	c.JSON(http.StatusOK, result)
}

// DeleteNet deletes a network
// @Summary Delete a network
// @Description Delete a network
// @tags net
// @Security apiKey
// @Param id path string true "Network ID"
// @Success 200 {object} string
// @Failure 400 {object} error
// @Router /net/{id} [delete]
func deleteNet(c *gin.Context) {
	id := c.Param("id")

	account, v, err := core.AuthFromContext(c, id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account from context")
		return
	}
	net := v.(*model.Network)

	if account.Status == "Suspended" {
		log.Infof("deleteNet: account %s is suspended", account.Email)
		c.JSON(http.StatusForbidden, gin.H{"error": "account is suspended"})
		return
	}

	if (account.Role != "Admin" && account.Role != "Owner") && account.Email != net.CreatedBy {
		log.Infof("deleteNet: user %s is not an admin of %s", account.Email, account.Id)
		c.JSON(http.StatusForbidden, gin.H{"error": "You cannot delete this network"})
		return
	}

	err = core.DeleteNet(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to delete network")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

// ReadNetworks reads all networks
// @Summary Read all networks
// @Description Read all networks
// @tags net
// @Produce  json
// @Security apiKey
// @Success 200 {array} []model.Network
// @Failure 400 {object} error
// @Router /net [get]
func readNetworks(c *gin.Context) {

	account, _, err := core.AuthFromContext(c, "")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account from context")
		return
	}

	nets, err := core.ReadNetworks(account.Email)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to list nets")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, nets)
}
