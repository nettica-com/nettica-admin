package client

import (
	"archive/zip"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	core "github.com/nettica-com/nettica-admin/core"
	model "github.com/nettica-com/nettica-admin/model"
	util "github.com/nettica-com/nettica-admin/util"
	log "github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
	"golang.org/x/oauth2"
)

//var StatusCache *cache.Cache

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

	//	StatusCache = cache.New(1*time.Minute, 10*time.Minute)
}

func createVPN(c *gin.Context) {
	var data model.VPN

	if err := c.ShouldBindJSON(&data); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to bind")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	a := util.GetCleanAuthToken(c)
	log.Infof("%v", a)

	account, _, err := core.AuthFromContext(c, data.AccountID)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to get account from context")
		return
	}

	data.CreatedBy = account.Email
	data.UpdatedBy = account.Email

	if data.AccountID == "" {
		data.AccountID = account.Id
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

	if account.Status == "Suspended" {
		log.Errorf("readVPN: account %s is suspended", account.Email)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	c.JSON(http.StatusOK, vpn)
}

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
		c.AbortWithStatus(http.StatusUnprocessableEntity)
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

		data.UpdatedBy = device.Name

		if !authorized {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
	} else {

		if vpn.CreatedBy == account.Email || account.Role == "Admin" || account.Role == "Owner" {
			authorized = true
		}

		if !authorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		data.UpdatedBy = account.Email
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

func readVPNs(c *gin.Context) {
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
	clients, err := core.ReadVPNsForUser(user.Email)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to list clients")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, clients)
}

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
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	formatQr := c.DefaultQuery("qrcode", "false")
	zipcode := c.DefaultQuery("zip", "false")

	data, net, err := core.ReadVPNConfig(id)
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
