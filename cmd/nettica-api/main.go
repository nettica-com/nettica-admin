package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	helmet "github.com/danielkov/gin-helmet"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	api "github.com/nettica-com/nettica-admin/api"
	auth "github.com/nettica-com/nettica-admin/auth"
	core "github.com/nettica-com/nettica-admin/core"
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

func main() {
	log.Infof("Starting Nettica version: %s", version.Version)

	// load .env environment variables
	err := godotenv.Load()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("failed to load .env file")
	}

	// check directories or create it
	if !util.DirectoryExists(filepath.Join(os.Getenv("WG_CONF_DIR"))) {
		err = os.Mkdir(filepath.Join(os.Getenv("WG_CONF_DIR")), 0755)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
				"dir": filepath.Join(os.Getenv("WG_CONF_DIR")),
			}).Fatal("failed to create directory")
		}
	}

	// check if mesh.json exists otherwise create it with default values
	if !util.FileExists(filepath.Join(os.Getenv("WG_CONF_DIR"), "mesh.json")) {
		_, err = core.ReadServer()
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Fatal("mesh.json does not exist or can not read it")
		}
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
	app := gin.Default()

	// cors middleware
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AddAllowHeaders("Authorization", util.AuthTokenHeaderName)
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
		return

		//		c.AbortWithStatus(http.StatusUnauthorized)
		//		return
	})

	// apply api router private
	api.ApplyRoutes(app, true)

	err = app.Run(fmt.Sprintf("%s:%s", os.Getenv("SERVER"), os.Getenv("PORT")))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("failed to start server")
	}
}
