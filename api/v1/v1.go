package apiv1

import (
	"github.com/gin-gonic/gin"
	user "github.com/nettica-com/nettica-admin/api/v1/Users"
	"github.com/nettica-com/nettica-admin/api/v1/account"
	"github.com/nettica-com/nettica-admin/api/v1/auth"
	device "github.com/nettica-com/nettica-admin/api/v1/device"
	"github.com/nettica-com/nettica-admin/api/v1/net"
	"github.com/nettica-com/nettica-admin/api/v1/server"
	"github.com/nettica-com/nettica-admin/api/v1/service"
	"github.com/nettica-com/nettica-admin/api/v1/subscription"
	vpn "github.com/nettica-com/nettica-admin/api/v1/vpn"
)

// ApplyRoutes apply routes to gin router
func ApplyRoutes(r *gin.RouterGroup, private bool) {
	v1 := r.Group("/v1.0")
	{
		if private {
			account.ApplyRoutes(v1)
			server.ApplyRoutes(v1)
			user.ApplyRoutes(v1)
			net.ApplyRoutes(v1)
			service.ApplyRoutes(v1)
			subscription.ApplyRoutes(v1)
			vpn.ApplyRoutes(v1)
			device.ApplyRoutes(v1)
		} else {
			auth.ApplyRoutes(v1)

		}
	}
}
