package apiv1

import (
	"github.com/gin-gonic/gin"
	user "github.com/nettica-com/nettica-admin/api/v1/Users"
	"github.com/nettica-com/nettica-admin/api/v1/account"
	"github.com/nettica-com/nettica-admin/api/v1/auth"
	host "github.com/nettica-com/nettica-admin/api/v1/host"
	"github.com/nettica-com/nettica-admin/api/v1/mesh"
	"github.com/nettica-com/nettica-admin/api/v1/server"
	"github.com/nettica-com/nettica-admin/api/v1/service"
	"github.com/nettica-com/nettica-admin/api/v1/subscription"
)

// ApplyRoutes apply routes to gin router
func ApplyRoutes(r *gin.RouterGroup, private bool) {
	v1 := r.Group("/v1.0")
	{
		if private {
			account.ApplyRoutes(v1)
			host.ApplyRoutes(v1)
			server.ApplyRoutes(v1)
			user.ApplyRoutes(v1)
			mesh.ApplyRoutes(v1)
			service.ApplyRoutes(v1)
			subscription.ApplyRoutes(v1)
		} else {
			auth.ApplyRoutes(v1)

		}
	}
}
