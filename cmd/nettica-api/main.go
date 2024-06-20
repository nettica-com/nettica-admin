package main

import (
	"fmt"
	"os"
	"strings"
	"time"
        "net/http"

	helmet "github.com/danielkov/gin-helmet"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	api "github.com/nettica-com/nettica-admin/api"
	auth "github.com/nettica-com/nettica-admin/auth"
	docs "github.com/nettica-com/nettica-admin/cmd/nettica-api/docs"
	"github.com/nettica-com/nettica-admin/mongo"
	util "github.com/nettica-com/nettica-admin/util"
	version "github.com/nettica-com/nettica-admin/version"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

var (
	cacheDb = cache.New(60*time.Minute, 10*time.Minute)
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stderr)
	log.SetLevel(log.DebugLevel)
}

//		@title			Nettica API
//		@description	Nettica API documentation
//		@BasePath		/api/v1.0
//		@host			my.nettica.com
//	 @contactName	Nettica
//	 @contactEmail	support@nettica.com
//	 @contactURL	https://nettica.com
//		@schemes		https
//		@produce		json
//		@consumes		json
//		@license		MIT
//	 @securityDefinitions.apiKey apiKey
//	 @in header
//	 @name X-API-KEY
func main() {
	log.Infof("Starting Nettica version: %s", version.Version)

	// load .env environment variables
	err := godotenv.Load()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("failed to load .env file")
	}

	if os.Getenv("GIN_MODE") == "debug" {
		// set gin release debug
		gin.SetMode(gin.DebugMode)
	} else {
		// set gin release mode
		gin.SetMode(gin.ReleaseMode)
		// disable console color
		gin.DisableConsoleColor()
		// log level info
		log.SetLevel(log.InfoLevel)
	}

	// creates a gin router with default middleware: logger and recovery (crash-free) middleware
	app := gin.New()
	app.Use(gin.Logger())
	app.Use(gin.Recovery())

	docs.SwaggerInfo.BasePath = "/api/v1.0"
	docs.SwaggerInfo.Host = os.Getenv("SERVER")[8:]
	docs.SwaggerInfo.Schemes = []string{"https"}

	// cors middleware
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AddAllowHeaders("Authorization", util.AuthTokenHeaderName)
	config.AddAllowHeaders("X-API-KEY")
	app.Use(cors.New(config))

	// protection middleware
	app.Use(helmet.Default())

	// add cache storage to gin app
	app.Use(func(ctx *gin.Context) {
		ctx.Set("cache", cacheDb)
		ctx.Next()
	})

	// serve static files
	app.Use(static.Serve("/", static.LocalFile("./ui/dist", false)))

	serveIndex := func(c *gin.Context) {
	        c.Writer.WriteHeader(http.StatusOK)
		log.Error("Show me the money")
        	http.ServeFile(c.Writer, c.Request, "./ui/dist/index.html")
		log.Error("***** Is this being called? *****")
		c.AbortWithStatus(http.StatusOK)
	}
	// make a catch-all route to serve the index.html
//	app.NoRoute(serveIndex)
	app.GET("/login", serveIndex)
	app.GET("/consent", serveIndex)
	app.GET("/join", serveIndex)

	// setup Oauth2 client
	oauth2Client, err := auth.GetAuthProvider()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("failed to setup Oauth2")
	}

	app.Use(func(ctx *gin.Context) {
		ctx.Set("oauth2Client", oauth2Client)
		ctx.Next()
	})

	// apply api routes public
	api.ApplyRoutes(app, false)

	// simple middleware to check auth
	app.Use(func(c *gin.Context) {
		cacheDb := c.MustGet("cache").(*cache.Cache)

		token := util.GetCleanAuthToken(c)

		oauth2Token, exists := cacheDb.Get(token)
		if exists && oauth2Token.(*oauth2.Token).AccessToken == token {
			// will be accessible in auth endpoints
			c.Set("oauth2Token", oauth2Token)
			c.Next()
			return
		}

		// avoid 401 page for refresh after logout
		if !strings.Contains(c.Request.URL.Path, "/api/") {
			c.Redirect(301, "/index.html")
			return
		}

		c.Next()

		//		c.AbortWithStatus(http.StatusUnauthorized)
		//		return
	})

	// apply api router private
	api.ApplyRoutes(app, true)

	// Initialize the database
	err = mongo.Initialize()
	if err != nil {
		log.Error(err)
	}

	err = app.Run(fmt.Sprintf("%s:%s", os.Getenv("LISTEN_ADDR"), os.Getenv("PORT")))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("failed to start server")
	}
}
