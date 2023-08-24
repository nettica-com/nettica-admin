package client

import (
	"archive/zip"
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

		device, err := core.ReadDevice(data.DeviceID)
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

	vpn, err := core.ReadVPN(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read vpn")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	vpns, err := core.ReadVPN2("netid", vpn.NetId)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read vpns")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err = core.DeleteVPN(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to remove client")
		c.AbortWithStatus(http.StatusInternalServerError)
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

}
