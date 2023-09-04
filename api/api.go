package api

import (
	"github.com/gin-gonic/gin"
	apiv1 "github.com/nettica-com/nettica-admin/api/v1"
	swagger "github.com/nettica-com/nettica-admin/swagger"
)

// ApplyRoutes apply routes to gin engine
func ApplyRoutes(r *gin.Engine, private bool) {
	api := r.Group("/api")
	{
		apiv1.ApplyRoutes(api, private)
		swagger.ApplyRoutes(api, private)
	}
}
