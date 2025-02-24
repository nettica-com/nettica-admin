package auth

import (
	"encoding/base64"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	model "github.com/nettica-com/nettica-admin/model"
	"github.com/nettica-com/nettica-admin/shadow"
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
		g.POST("/validate", validate)
		g.POST("/login", login)
		g.GET("/user", user)
		g.GET("/logout", logout)
	}
}

/*
 * generate redirect url to get OAuth2 code or let client know that OAuth2 is disabled
 */
func oauth2URL(c *gin.Context) {
	cacheDb := c.MustGet("cache").(*cache.Cache)
	oauth2Client := c.MustGet("oauth2Client").(model.Authentication)

	referer := c.Request.URL.Query().Get("referer")

	var err error
	var state, clientId, codeUrl, audience, redirect_uri string
	if c.Request.URL.Query().Get("redirect_uri") == "com.nettica.agent://callback/agent" {
		clientId, err = util.GenerateRandomString(32)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to generate state random string")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		state, err = util.GenerateRandomString(32)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to generate state random string")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		codeUrl = oauth2Client.CodeUrl2(state)
		audience = os.Getenv("OAUTH2_AGENT_AUDIENCE")
		redirect_uri = os.Getenv("OAUTH2_AGENT_REDIRECT_URL")

		cacheDb.Set(clientId, state, 1*time.Hour)
	} else {

		state, err = util.GenerateRandomString(32)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to generate state random string")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		clientId, err = util.GenerateRandomString(32)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to generate state random string")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// save clientId and state so we can retrieve for verification
		cacheDb.Set(clientId, state, 1*time.Hour)
		codeUrl = oauth2Client.CodeUrl(state)
		if referer != "" {
			codeUrl = codeUrl + "&referer=" + referer
		}
	}

	data := &model.Auth{
		Oauth2:   true,
		ClientId: clientId,
		State:    state,
		CodeUrl:  codeUrl,
		Audience: audience,
		Redirect: redirect_uri,
	}

	log.Infof("model.Auth = %v", data)

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

	savedCode, exists := cacheDb.Get(loginVals.Code)
	if exists {
		loginVals.Code = savedCode.(string)
	}

	oauth2Client := c.MustGet("oauth2Client").(model.Authentication)

	var oauth2Token *oauth2.Token
	var err error

	if loginVals.Redirect != "com.nettica.agent://callback/agent" {
		oauth2Token, err = oauth2Client.Exchange(loginVals.Code)
	} else {
		oauth2Token, err = oauth2Client.Exchange2(loginVals.Code)
	}

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
	cacheDb.Set(oauth2Token.AccessToken, oauth2Token, 24*time.Hour)

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
	savedState, exists := cacheDb.Get(loginVals.Code)

	if !exists || savedState != loginVals.State {
		log.WithFields(log.Fields{
			"state":      loginVals.State,
			"savedState": savedState,
		}).Error("saved state and client provided state mismatch")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	oauth2Client := c.MustGet("oauth2Client").(model.Authentication)

	var oauth2Token *oauth2.Token
	var err error

	if loginVals.Redirect != "com.nettica.agent://callback/agent" {
		oauth2Token, err = oauth2Client.Exchange(loginVals.Code)
	} else {
		oauth2Token, err = oauth2Client.Exchange2(loginVals.Code)
	}
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to exchange code for token")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	cacheDb.Set(oauth2Token.AccessToken, oauth2Token, 24*time.Hour)

	c.JSON(http.StatusOK, oauth2Token.AccessToken)
	/*
	   //	cacheDb.Delete(loginVals.ClientId)
	   var token oauth2.Token
	   token.AccessToken = loginVals.Code
	   var token_map = make(map[string]interface{}, 1)
	   token_map["id_token"] = loginVals.Code
	   token2 := token.WithExtra(token_map)

	   cacheDb.Set(loginVals.Code, token2, cache.DefaultExpiration)

	   c.JSON(http.StatusOK, loginVals.Code)
	*/
}

// login with basic auth, but do it the OAuth way
func login(c *gin.Context) {
	var loginVals model.Auth
	var err error
	if err = c.ShouldBind(&loginVals); err != nil {
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

	// code contains the username and password base64 encoded
	// base64 decode the string

	var userpass []byte

	userpass, err = base64.StdEncoding.DecodeString(loginVals.Code)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("not base64 encoded")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// split the username and password
	parts := strings.SplitN(string(userpass), ":", 2)
	if len(parts) != 2 {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("invalid username or password")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	user := parts[0]
	pass := parts[1]

	// validate the username and password
	err = shadow.ShadowAuthPlain(user, pass)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("shadow: invalid username or password")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	code, err := util.GenerateRandomString(32)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to generate random string")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// save the code and basic auth in the cache for
	// later retrieval in oauth2_exchange
	cacheDb.Set(code, loginVals.Code, 1*time.Hour)

	redirect := "/?code=" + code + "&state=" + loginVals.State

	if loginVals.Redirect != "" {
		redirect = loginVals.Redirect + "?code=" + code + "&state=" + loginVals.State

		// If it's the nettica agent application send a redirect
		if loginVals.Redirect == "com.nettica.agent://callback/agent" {

			c.Redirect(http.StatusPermanentRedirect, redirect)
			return
		}
	}

	// otherwise send a JSON body with the result to the browser.  it will do the redirect.
	loginVals.Redirect = redirect

	c.JSON(http.StatusOK, loginVals)
}

func validate(c *gin.Context) {
	var t model.OAuth2Token
	if err := c.ShouldBindJSON(&t); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("could not validate tokens")
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}
	cacheDb := c.MustGet("cache").(*cache.Cache)
	oauth2Token, exists := cacheDb.Get(util.GetCleanAuthToken(c))

	if exists && oauth2Token.(*oauth2.Token).AccessToken == util.GetCleanAuthToken(c) {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	// validate the JWT with our private key

	// verify the jwt signature

	token := t.AccessToken

	// parse the token
	parsedToken, err := jwt.ParseWithClaims(token, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return os.Getenv("OAUTH2_CLIENT_SECRET"), nil
	})

	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to parse jwt token")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// validate the claims
	if claims, ok := parsedToken.Claims.(*CustomClaims); ok && parsedToken.Valid {
		log.WithFields(log.Fields{
			"claims": claims,
		}).Info("claims")

		// create a new oauth2.Token from the claims
		oauth2Token := &oauth2.Token{
			AccessToken:  token,
			TokenType:    "Bearer",
			RefreshToken: "",
			Expiry:       time.Now().Add(4 * time.Hour),
		}

		oauth2Token = oauth2Token.WithExtra(map[string]interface{}{ // Add the ID token to the extra parameters
			"id_token": token})

		cacheDb.Set(oauth2Token.AccessToken, oauth2Token, 24*time.Hour)

		c.JSON(http.StatusOK, gin.H{})
		return
	}

	// otherwise we have an invalid token

	log.Error("oauth2 AccessToken is not recognized")

	c.AbortWithStatus(http.StatusUnauthorized)
}

// A custom struct to hold the jwt claims
type CustomClaims struct {
	Email string `json:"email"`

	jwt.StandardClaims
}

func logout(c *gin.Context) {

	cacheDb := c.MustGet("cache").(*cache.Cache)

	if c.Request.URL.Query().Get("user") != "" {
		oauth2Client := c.MustGet("oauth2Client").(model.Authentication)
		// delete all tokens for this user
		items := cacheDb.Items()
		for _, token := range items {
			// check to see if the item is a string or a token
			if _, ok := token.Object.(*oauth2.Token); ok {

				user, err := oauth2Client.UserInfo(token.Object.(*oauth2.Token))
				if err != nil {
					log.WithFields(log.Fields{
						"err": err,
					}).Error("failed to get user from oauth2 AccessToken")
					c.AbortWithStatus(http.StatusBadRequest)
					return
				}
				if user.Email == c.Request.URL.Query().Get("user") {
					cacheDb.Delete(token.Object.(*oauth2.Token).AccessToken)
				}
			}
		}
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	cacheDb.Delete(c.Request.Header.Get(util.AuthTokenHeaderName))

	var logoutUrl string

	if c.Request.URL.Query().Get("redirect_uri") != "" {
		logoutUrl = os.Getenv("OAUTH2_AGENT_LOGOUT_URL")
		if logoutUrl != "" {
			c.Redirect(http.StatusTemporaryRedirect, logoutUrl)
			return
		}
	}

	logoutUrl = os.Getenv("OAUTH2_LOGOUT_URL")
	if logoutUrl != "" {
		c.Redirect(http.StatusTemporaryRedirect, logoutUrl)
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, "/")
}

func user(c *gin.Context) {
	cacheDb := c.MustGet("cache").(*cache.Cache)
	token := util.GetCleanAuthToken(c)
	oauth2Token, exists := cacheDb.Get(token)
	id_token := c.Request.Header.Get("X-OAUTH2-ID-TOKEN")

	if id_token != "" {
		new_token := &oauth2.Token{
			AccessToken:  token,
			TokenType:    "Bearer",
			RefreshToken: "",
			Expiry:       time.Now().Add(time.Hour * 24),
		}
		m := make(map[string]interface{})
		m["id_token"] = id_token
		new_token = new_token.WithExtra(m)

		// check if token is valid
		var err error
		oauth2Token, err = util.ValidateToken(new_token.AccessToken)
		if err != nil {
			log.WithFields(log.Fields{
				"err":   err,
				"token": oauth2Token,
			}).Error("failed to get token info")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		oauth2Token = new_token
	}

	if id_token != "" || (exists && oauth2Token.(*oauth2.Token).AccessToken == util.GetCleanAuthToken(c)) {
		oauth2Client := c.MustGet("oauth2Client").(model.Authentication)

		user, err := oauth2Client.UserInfo(oauth2Token.(*oauth2.Token))
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("failed to get user from oauth2 AccessToken")
			cacheDb.Delete(token)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.JSON(http.StatusOK, user)
		return
	}

	log.Error("oauth2 AccessToken is not recognized")

	c.AbortWithStatus(http.StatusUnauthorized)
}
