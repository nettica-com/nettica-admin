package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	auth "github.com/nettica-com/nettica-admin/auth"
	model "github.com/nettica-com/nettica-admin/model"
	util "github.com/nettica-com/nettica-admin/util"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// ApplyRoutes applies router to gin Router
func ApplyRoutes(r *gin.RouterGroup) {
	g := r.Group("/auth")
	{
		g.GET("/oauth2_url", oauth2URL)
		g.POST("/oauth2_exchange", oauth2Exchange)
		g.POST("/token", token)
		g.GET("/user", user)
		g.GET("/logout", logout)
	}
}

/*
 * generate redirect url to get OAuth2 code or let client know that OAuth2 is disabled
 */
func oauth2URL(c *gin.Context) {
	cacheDb := c.MustGet("cache").(*cache.Cache)

	state, err := util.GenerateRandomString(32)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to generate state random string")
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	clientId, err := util.GenerateRandomString(32)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to generate state random string")
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	// save clientId and state so we can retrieve for verification
	cacheDb.Set(clientId, state, 5*time.Minute)

	oauth2Client := c.MustGet("oauth2Client").(auth.Auth)

	data := &model.Auth{
		Oauth2:   true,
		ClientId: clientId,
		State:    state,
		CodeUrl:  oauth2Client.CodeUrl(state),
	}

	c.JSON(http.StatusOK, data)
}

/*
 * exchange code and get user infos, if OAuth2 is disable just send fake data
 */
func oauth2Exchange(c *gin.Context) {
	var loginVals model.Auth
	if err := c.ShouldBind(&loginVals); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("code and state fields are missing")
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}
	log.WithFields(log.Fields{
		"loginVals": loginVals,
	}).Info("loginVals")

	cacheDb := c.MustGet("cache").(*cache.Cache)
	savedState, exists := cacheDb.Get(loginVals.ClientId)

	if loginVals.State != "basic_auth" {
		if !exists || savedState != loginVals.State {
			log.WithFields(log.Fields{
				"state":      loginVals.State,
				"savedState": savedState,
			}).Error("saved state and client provided state mismatch")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}
	oauth2Client := c.MustGet("oauth2Client").(auth.Auth)

	oauth2Token, err := oauth2Client.Exchange(loginVals.Code)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to exchange code for token")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// normally we should delete this, but frankly it causes more errors on the website to do that.
	// Let it be expired out of the cache instead of deleting it.
	// cacheDb.Delete(loginVals.ClientId)
	cacheDb.Set(oauth2Token.AccessToken, oauth2Token, 4*time.Hour)

	c.JSON(http.StatusOK, oauth2Token.AccessToken)
}

/*
 * exchange code and get user infos, if OAuth2 is disable just send fake data
 */
func token(c *gin.Context) {
	var loginVals model.Auth
	if err := c.ShouldBindJSON(&loginVals); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("code and state fields are missing")
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}
	log.WithFields(log.Fields{
		"loginVals": loginVals,
	}).Info("loginVals")

	cacheDb := c.MustGet("cache").(*cache.Cache)
	//	savedState, exists := cacheDb.Get(loginVals.ClientId)

	//	if !exists || savedState != loginVals.State {
	//		log.WithFields(log.Fields{
	//			"state":      loginVals.State,
	//			"savedState": savedState,
	//		}).Error("saved state and client provided state mismatch")
	//		c.AbortWithStatus(http.StatusBadRequest)
	//		return
	//	}
	//oauth2Client := c.MustGet("oauth2Client").(auth.Auth)

	//	oauth2Token, err := oauth2Client.Exchange(loginVals.Code)
	//	if err != nil {
	//		log.WithFields(log.Fields{
	//			"err": err,
	//		}).Error("failed to exchange code for token")
	//		c.AbortWithStatus(http.StatusBadRequest)
	//		return
	//	}

	//	cacheDb.Delete(loginVals.ClientId)
	var token oauth2.Token
	token.AccessToken = loginVals.Code
	var token_map = make(map[string]interface{}, 1)
	token_map["id_token"] = loginVals.Code
	token2 := token.WithExtra(token_map)

	cacheDb.Set(loginVals.Code, token2, cache.DefaultExpiration)

	c.JSON(http.StatusOK, loginVals.Code)
}

func logout(c *gin.Context) {
	cacheDb := c.MustGet("cache").(*cache.Cache)
	cacheDb.Delete(c.Request.Header.Get(util.AuthTokenHeaderName))
	c.JSON(http.StatusOK, gin.H{})
}

func user(c *gin.Context) {
	cacheDb := c.MustGet("cache").(*cache.Cache)
	oauth2Token, exists := cacheDb.Get(util.GetCleanAuthToken(c))

	if exists && oauth2Token.(*oauth2.Token).AccessToken == util.GetCleanAuthToken(c) {
		oauth2Client := c.MustGet("oauth2Client").(auth.Auth)

		user, err := oauth2Client.UserInfo(oauth2Token.(*oauth2.Token))
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to get user from oauth2 AccessToken")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		c.JSON(http.StatusOK, user)
		return
	}

	log.WithFields(log.Fields{
		"exists":                 exists,
		util.AuthTokenHeaderName: util.GetCleanAuthToken(c),
	}).Error("oauth2 AccessToken is not recognized")

	c.AbortWithStatus(http.StatusUnauthorized)
}
