package swagger

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupSwagger sets up swagger documentation
// ApplyRoutes apply routes to gin engine
func ApplyRoutes(r *gin.RouterGroup, private bool) {
	if private {
		return
	}
	r.GET("/docs/*any", swagger)
}

func swagger(c *gin.Context) {
	if c.Request.URL.Path == "/api/docs/" {
		c.Redirect(302, "/api/docs/index.html")
		return
	}

	ginSwagger.WrapHandler(swaggerfiles.Handler)(c)
}
