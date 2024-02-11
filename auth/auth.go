package auth

import (
	"fmt"
	"os"

	"github.com/nettica-com/nettica-admin/auth/basic"
	"github.com/nettica-com/nettica-admin/auth/fake"
	"github.com/nettica-com/nettica-admin/auth/github"
	"github.com/nettica-com/nettica-admin/auth/google"
	"github.com/nettica-com/nettica-admin/auth/microsoft"
	"github.com/nettica-com/nettica-admin/auth/microsoft2"
	"github.com/nettica-com/nettica-admin/auth/oauth2oidc"
	model "github.com/nettica-com/nettica-admin/model"
	log "github.com/sirupsen/logrus"
)

// GetAuthProvider  get an instance of auth provider based on config
func GetAuthProvider() (model.Authentication, error) {
	var oauth2Client model.Authentication
	var err error

	switch os.Getenv("OAUTH2_PROVIDER_NAME") {
	case "fake":
		log.Warn("Oauth is set to fake, no actual authentication will be performed")
		oauth2Client = &fake.Fake{}

	case "oauth2oidc":
		log.Warn("Oauth is set to oauth2oidc, must be RFC implementation on server side")
		oauth2Client = &oauth2oidc.Oauth2idc{}

	case "microsoft":
		log.Warn("Oauth is set to Microsoft")
		oauth2Client = &microsoft.Oauth2Msft{}

	case "microsoft2":
		log.Warn("Oauth is set to Microsoft2")
		oauth2Client = &microsoft2.Oauth2Microsoft{}

	case "basic":
		log.Warn("Oauth is set to basic.  Authenication against the shadow file")
		oauth2Client = &basic.Oauth2Basic{}

	case "github":
		log.Warn("Oauth is set to github, no openid will be used")
		oauth2Client = &github.Github{}

	case "google":
		log.Warn("Oauth is set to Google")
		oauth2Client = &google.OAuth2Google{}

	default:
		return nil, fmt.Errorf("auth provider name %s unknown", os.Getenv("OAUTH2_PROVIDER_NAME"))
	}

	err = oauth2Client.Setup()

	return oauth2Client, err
}
