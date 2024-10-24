package subscription

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nettica-com/nettica-admin/core"
	model "github.com/nettica-com/nettica-admin/model"
	"github.com/nettica-com/nettica-admin/mongo"
	"github.com/nettica-com/nettica-admin/util"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// ApplyRoutes applies router to gin Router
func ApplyRoutes(r *gin.RouterGroup) {
	g := r.Group("/subscriptions")
	{

		g.POST("/helio", createHelioSubscription)
		g.POST("", createSubscription)
		g.POST("/apple", createSubscriptionApple)
		g.GET("/:id", readSubscription)
		g.PATCH("/:id", updateSubscription)
		g.DELETE("/:id", deleteSubscription)
		g.GET("", readSubscriptions)
	}
}

func createSubscriptionApple(c *gin.Context) {
	var receipt model.PurchaseRceipt

	err := json.NewDecoder(c.Request.Body).Decode(&receipt)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// Validate the receipt with Apple
	valid, err := validateReceiptApple(receipt.Receipt)
	if err != nil || !valid {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid receipt"})
		return
	}

	customer_name := receipt.Name

	credits := 0
	name := ""
	description := ""
	devices := 0
	networks := 0
	members := 0
	relays := 0
	autoRenew := false
	issued := time.Now()
	expires := time.Now().AddDate(1, 0, 0)

	switch receipt.ProductID {
	case "basic_monthly":
		name = "Basic Service (monthly)"
	case "basic_yearly":
		name = "Basic Service (yearly)"
	case "premium_monthly":
		name = "Premium Service (monthly)"
	case "premium_yearly":
		name = "Premium Service (yearly)"
	case "professional_monthly":
		name = "Professional Service (monthly)"
	case "professional_yearly":
		name = "Professional Service (yearly)"
	default:
		log.Errorf("unknown sku %s", receipt.ProductID)
	}
	// set the credits, name, and description based on the sku
	switch receipt.ProductID {
	case "24_hours":
		credits = 1
		name = "24 Hours"
		description = "Service in any region for 24 hours"
		devices = 5
		networks = 1
		members = 2
		relays = 1
		autoRenew = false
		expires = time.Now().Add(24 * time.Hour)
	case "1_month":
		credits = 1
		name = "1 Month"
		description = "Service in any region for 1 month"
		devices = 5
		networks = 1
		members = 2
		relays = 1
		autoRenew = false
		expires = time.Now().AddDate(0, 1, 0)
	case "1_week":
		credits = 1
		name = "1 Week"
		description = "Service in any region for 1 week"
		devices = 5
		networks = 1
		members = 2
		relays = 1
		autoRenew = false
		expires = time.Now().AddDate(0, 0, 7)
	case "basic_monthly", "basic_yearly":
		description = "A single tunnel or relay in any region"
		devices = 5
		networks = 1
		members = 2
		relays = 1
		autoRenew = true
	case "premium_monthly", "premium_yearly":
		credits = 5
		name = "Premium"
		description = "Up to 5 tunnels or relays in any region"
		devices = 25
		networks = 10
		members = 5
		relays = 5
		autoRenew = true
	case "professional_monthly", "professional_yearly":
		credits = 10
		name = "Professional"
		description = "Up to 10 tunnels or relays in any region"
		devices = 100
		networks = 25
		members = 25
		relays = 10
		autoRenew = true
	default:
		log.Errorf("unknown sku %s", receipt.ProductID)
	}

	// set the limits based on the sku
	accounts, err := core.ReadAllAccounts(receipt.Email)
	if err != nil {
		log.Error(err)
	} else {
		//  If there's no error and no account, create one.
		if len(accounts) == 0 {
			var account model.Account
			account.Name = customer_name
			account.AccountName = "Company"
			account.Email = receipt.Email
			account.Role = "Owner"
			account.Status = "Active"
			account.CreatedBy = receipt.Email
			account.UpdatedBy = receipt.Email
			account.Picture = "/account-circle.svg"

			a, err := core.CreateAccount(&account)
			log.Infof("CREATE ACCOUNT = %v", a)
			if err != nil {
				log.Error(err)
			}
			accounts, err = core.ReadAllAccounts(receipt.Email)
			if err != nil {
				log.Error(err)
			}

		}
	}

	var account *model.Account
	for i := 0; i < len(accounts); i++ {
		if accounts[i].Id == accounts[i].Parent {
			account = accounts[i]
			break
		}
	}

	if account == nil {
		log.Errorf("account not found for email %s", receipt.Email)
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	limits, err := core.ReadLimits(account.Id)
	if err != nil {
		log.Error(err)
		limits_id, err := util.GenerateRandomString(8)
		if err != nil {
			log.Error(err)
		}
		limits_id = "limits-" + limits_id

		limits = &model.Limits{
			Id:          limits_id,
			AccountID:   account.Id,
			MaxDevices:  0,
			MaxNetworks: 0,
			MaxMembers:  0,
			MaxServices: 0,
			Tolerance:   core.GetDefaultTolerance(),
			CreatedBy:   receipt.Email,
			UpdatedBy:   receipt.Email,
			Created:     time.Now(),
			Updated:     time.Now(),
		}
	}

	limits.MaxDevices += devices
	limits.MaxNetworks += networks
	limits.MaxMembers += members
	limits.MaxServices += relays

	errs := limits.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("limits validation error")
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Limits validation error"})
		return
	}

	// save limits to mongodb
	mongo.Serialize(limits.Id, "id", "limits", limits)

	// generate a random subscription id
	id, err := util.RandomString(8)
	if err != nil {
		log.Error(err)
	}
	id = receipt.Source + "-" + id

	// construct a subscription object
	lu := time.Now()
	subscription := model.Subscription{
		Id:          id,
		AccountID:   account.Id,
		Email:       receipt.Email,
		Name:        name,
		Description: description,
		Issued:      &issued,
		LastUpdated: &lu,
		Expires:     &expires,
		Credits:     credits,
		Sku:         receipt.ProductID,
		Status:      "active",
		AutoRenew:   autoRenew,
		Receipt:     receipt.Receipt,
	}

	errs = subscription.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("subscription validation error")
		}
		return
	}

	// save subscription to mongodb
	mongo.Serialize(subscription.Id, "id", "subscriptions", subscription)

	c.JSON(http.StatusOK, subscription)
	return

}

func validateReceiptApple(receipt string) (bool, error) {
	// Apple receipt validation URL
	url := "https://buy.itunes.apple.com/verifyReceipt"

	// Create the request payload
	payload := map[string]string{
		"receipt-data": receipt,
		"password":     os.Getenv("APPLE_ITUNES_SHARED_SECRET"), // Replace with your app's shared secret
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}

	// Send the request to Apple
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("invalid response from Apple: %s", resp.Status)
	}

	// Parse the response to check the receipt status
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return false, err
	}

	// Check if the receipt is valid
	if status, ok := result["status"].(float64); ok && status == 0 {
		return true, nil
	}

	return false, nil
}

func createHelioSubscription(c *gin.Context) {
	var body string
	var sub map[string]interface{}

	HelioPaylinkIdRelay := os.Getenv("HELIO_PAYLINK_ID_RELAY")
	HelioPaylinkIdPremium := os.Getenv("HELIO_PAYLINK_ID_PREMIUM")
	HelioPaylinkIdPro := os.Getenv("HELIO_PAYLINK_ID_PRO")

	HelioPaylinkTokenRelay := os.Getenv("HELIO_PAYLINK_TOKEN_RELAY")
	HelioPaylinkTokenPremium := os.Getenv("HELIO_PAYLINK_TOKEN_PREMIUM")
	HelioPaylinkTokenPro := os.Getenv("HELIO_PAYLINK_TOKEN_PRO")

	// read and log the request body

	bytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read request body")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	body = string(bytes)
	// remove all the backslashes from the body (is this needed?)
	body = strings.Replace(body, "\\", "", -1)
	log.Info(body)
	bytes = []byte(body)

	// unmarshal the body into a map
	err = json.Unmarshal(bytes, &sub)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to unmarshal request body")
		// c.AbortWithStatus(http.StatusUnprocessableEntity)
		// return with no error so webhook doesn't get disabled
		c.JSON(http.StatusOK, body)
		return
	}

	log.Info(sub)

	var sku string
	var sharedKey string

	// get the transact id from the sub object

	transact := sub["transaction"].(map[string]interface{})

	paylinkId := transact["paystream"].(string)

	switch paylinkId {
	case HelioPaylinkIdRelay:
		sku = "Relay-1"
		sharedKey = "Bearer " + HelioPaylinkTokenRelay
	case HelioPaylinkIdPremium:
		sku = "Premium-5"
		sharedKey = "Bearer " + HelioPaylinkTokenPremium
	case HelioPaylinkIdPro:
		sku = "Pro-10"
		sharedKey = "Bearer " + HelioPaylinkTokenPro
	default:
		log.Errorf("unknown paylink_id %s", paylinkId)
	}

	// Read the Authorization header
	authorization := c.Request.Header.Get("Authorization")
	if authorization == "" {
		log.Error("Authorization header is empty")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is empty"})
		return
	}

	// Compare the sharedKey with the Authorization header

	if sharedKey != authorization {
		log.Errorf("Authenication denied")
		// c.JSON(http.StatusUnauthorized, gin.H{"error": "Authenication denied"})
		// return
	}

	// walk the json and find the email address

	meta := transact["meta"].(map[string]interface{})
	log.Info(meta)

	customer := meta["customerDetails"].(map[string]interface{})
	log.Info(customer)

	email := customer["email"].(string)
	log.Info(email)

	transStatus := sub["event"].(string)

	var status string

	switch transStatus {
	case "STARTED":
		status = "active"
	case "STOPPED":
		status = "expired"
	case "CANCELLED":
		status = "cancelled"
	default:
		log.Errorf("unknown transaction status %s", transStatus)
	}

	endedAt := transact["endedAt"].(string)
	log.Info(endedAt)
	// convert the endedAt string of unixtime to a time.Time object
	endedAtUnix, err := time.Parse(time.RFC3339, endedAt)
	if err != nil {
		log.Error(err)
	}

	start := transact["startedAt"].(map[string]interface{})
	id := start["transactionSignature"].(string)
	log.Info(id)

	go func() {

		customer_name := "Me"

		credits := 0
		name := ""
		description := ""
		devices := 0
		networks := 0
		members := 0
		relays := 0

		// set the credits, name, and description based on the sku
		switch sku {
		case "Starter-0":
			credits = 0
			name = "Starter"
			description = "The Starter subscription"
			devices = core.GetDefaultMaxDevices()
			networks = core.GetDefaultMaxNetworks()
			members = core.GetDefaultMaxMembers()
			relays = core.GetDefaultMaxServices()

		case "Relay-1":
			fallthrough
		case "RelayYear-1":
			credits = 1
			name = "Relay Service"
			description = "A single tunnel or relay in any Region"
			devices = 5
			networks = 1
			members = 2
			relays = 1
		case "Premium-5":
			fallthrough
		case "PremiumYear-5":
			credits = 5
			name = "Premium"
			description = "Up to 5 tunnels or relays in any Region"
			devices = 25
			networks = 10
			members = 5
			relays = 5
		case "Pro-10":
			fallthrough
		case "ProYear-10":
			credits = 10
			name = "Professional"
			description = "Up to 10 tunnels or relays in any Region"
			devices = 100
			networks = 25
			members = 25
			relays = 10
		default:
			log.Errorf("unknown sku %s", sku)
		}

		// set the limits based on the sku
		accounts, err := core.ReadAllAccounts(email)
		if err != nil {
			log.Error(err)
		} else {
			//  If there's no error and no account, create one.
			if len(accounts) == 0 {
				var account model.Account
				account.Name = customer_name
				account.AccountName = "Company"
				account.Email = email
				account.Role = "Owner"
				account.Status = "Active"
				account.CreatedBy = email
				account.UpdatedBy = email
				account.Picture = "/account-circle.svg"

				a, err := core.CreateAccount(&account)
				log.Infof("CREATE ACCOUNT = %v", a)
				if err != nil {
					log.Error(err)
				}
				accounts, err = core.ReadAllAccounts(email)
				if err != nil {
					log.Error(err)
				}

			}
		}

		var account *model.Account
		for i := 0; i < len(accounts); i++ {
			if accounts[i].Id == accounts[i].Parent {
				account = accounts[i]
				break
			}
		}

		if account == nil {
			log.Errorf("account not found for email %s", email)
			return
		}

		limits, err := core.ReadLimits(account.Id)
		if err != nil {
			log.Error(err)
			limits_id, err := util.GenerateRandomString(8)
			if err != nil {
				log.Error(err)
			}
			limits_id = "limits-" + limits_id

			limits = &model.Limits{
				Id:          limits_id,
				AccountID:   account.Id,
				MaxDevices:  0,
				MaxNetworks: 0,
				MaxMembers:  0,
				MaxServices: 0,
				Tolerance:   core.GetDefaultTolerance(),
				CreatedBy:   email,
				UpdatedBy:   email,
				Created:     time.Now(),
				Updated:     time.Now(),
			}
		}

		limits.MaxDevices += devices
		limits.MaxNetworks += networks
		limits.MaxMembers += members
		limits.MaxServices += relays

		errs := limits.IsValid()
		if len(errs) != 0 {
			for _, err := range errs {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("limits validation error")
			}
			return
		}

		// save limits to mongodb
		mongo.Serialize(limits.Id, "id", "limits", limits)

		// construct a subscription object
		issued := time.Now()
		lu := time.Now()
		expires := endedAtUnix
		subscription := model.Subscription{
			Id:          id,
			AccountID:   account.Id,
			Email:       email,
			Name:        name,
			Description: description,
			Issued:      &issued,
			LastUpdated: &lu,
			Expires:     &expires,
			Credits:     credits,
			Sku:         sku,
			Status:      status,
		}

		errs = subscription.IsValid()
		if len(errs) != 0 {
			for _, err := range errs {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("subscription validation error")
			}
			return
		}

		// save subscription to mongodb
		mongo.Serialize(subscription.Id, "id", "subscriptions", subscription)

	}()

	c.JSON(http.StatusOK, body)
}

func createSubscription(c *gin.Context) {
	var body string
	var sub map[string]interface{}

	// get the secret and hash of the body
	secret := os.Getenv("WC_SECRET")
	signature := c.Request.Header.Get("x-wc-webhook-signature")

	// read and log the request body

	bytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read request body")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	body = string(bytes)
	// remove all the backslashes from the body (is this needed?)
	body = strings.Replace(body, "\\", "", -1)
	log.Info(body)
	bytes = []byte(body)

	// hash the body and compare it to the signature
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(bytes)
	expected := h.Sum(nil)
	if !hmac.Equal([]byte(signature), expected) {
		log.WithFields(log.Fields{
			"signature": signature,
			"expected":  expected,
			"body":      string(bytes),
		}).Error("failed to verify signature")
	} else {
		log.Info("signature verified")
	}

	// unmarshal the body into a map
	err = json.Unmarshal(bytes, &sub)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to unmarshal request body")
		// c.AbortWithStatus(http.StatusUnprocessableEntity)
		// return with no error so webhook doesn't get disabled
		c.JSON(http.StatusOK, body)
		return
	}

	log.Info(sub)

	// walk the json and find the customer href
	links := sub["_links"].(map[string]interface{})
	log.Info(links)

	// get the sku from the line_items
	sku := sub["line_items"].([]interface{})[0].(map[string]interface{})["sku"].(string)
	//status := sub["status"].(string)

	status := "active"

	customer := links["customer"].([]interface{})
	log.Info(customer)

	customer0 := customer[0].(map[string]interface{})
	log.Info(customer0)

	href := customer0["href"].(string)
	log.Info(href)

	go func() {

		// make http request with basic authentication using href as url to get the customer object
		req, err := http.NewRequest("GET", href, nil)
		if err != nil {
			return
		}

		req.SetBasicAuth(os.Getenv("WC_USERNAME"), os.Getenv("WC_PASSWORD"))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Errorf("http.client.Do = %v", err)
			return
		}

		if resp.StatusCode != 200 {
			log.Errorf("http status %s expect 200 OK", resp.Status)
			return
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error(err)
			return
		}
		defer resp.Body.Close()

		var data map[string]interface{}
		err = json.Unmarshal(bodyBytes, &data)
		if err != nil {
			log.Error(err)
			return
		}

		log.Info(data)

		// get the email from the customer object
		email := data["email"].(string)
		log.Info(email)

		customer_name := data["first_name"].(string) + " " + data["last_name"].(string)

		// generate a random subscription id
		id, err := util.RandomString(8)
		if err != nil {
			log.Error(err)
		}

		credits := 0
		name := ""
		description := ""
		devices := 0
		networks := 0
		members := 0
		relays := 0
		expires := time.Now().AddDate(1, 0, 0)

		// set the credits, name, and description based on the sku
		switch sku {
		case "Starter-0":
			credits = 0
			name = "Starter"
			description = "The Starter subscription"
			devices = core.GetDefaultMaxDevices()
			networks = core.GetDefaultMaxNetworks()
			members = core.GetDefaultMaxMembers()
			relays = core.GetDefaultMaxServices()

		case "Relay-1":
			fallthrough
		case "RelayYear-1":
			credits = 1
			name = "Relay Service"
			description = "A single tunnel or relay in any Region"
			devices = 5
			networks = 1
			members = 2
			relays = 1
		case "Premium-5":
			fallthrough
		case "PremiumYear-5":
			credits = 5
			name = "Premium"
			description = "Up to 5 tunnels or relays in any Region"
			devices = 25
			networks = 10
			members = 5
			relays = 5
		case "Pro-10":
			fallthrough
		case "ProYear-10":
			credits = 10
			name = "Professional"
			description = "Up to 10 tunnels or relays in any Region"
			devices = 100
			networks = 25
			members = 25
			relays = 10
		default:
			log.Errorf("unknown sku %s", sku)
		}

		// set the limits based on the sku
		accounts, err := core.ReadAllAccounts(email)
		if err != nil {
			log.Error(err)
		} else {
			//  If there's no error and no account, create one.
			if len(accounts) == 0 {
				var account model.Account
				account.Name = customer_name
				account.AccountName = "Company"
				account.Email = email
				account.Role = "Owner"
				account.Status = "Active"
				account.CreatedBy = email
				account.UpdatedBy = email
				account.Picture = "/account-circle.svg"

				a, err := core.CreateAccount(&account)
				log.Infof("CREATE ACCOUNT = %v", a)
				if err != nil {
					log.Error(err)
				}
				accounts, err = core.ReadAllAccounts(email)
				if err != nil {
					log.Error(err)
				}

			}
		}

		var account *model.Account
		for i := 0; i < len(accounts); i++ {
			if accounts[i].Id == accounts[i].Parent {
				account = accounts[i]
				break
			}
		}

		if account == nil {
			log.Errorf("account not found for email %s", email)
			return
		}

		limits, err := core.ReadLimits(account.Id)
		if err != nil {
			log.Error(err)
			limits_id, err := util.GenerateRandomString(8)
			if err != nil {
				log.Error(err)
			}
			limits_id = "limits-" + limits_id

			limits = &model.Limits{
				Id:          limits_id,
				AccountID:   account.Id,
				MaxDevices:  0,
				MaxNetworks: 0,
				MaxMembers:  0,
				MaxServices: 0,
				Tolerance:   core.GetDefaultTolerance(),
				CreatedBy:   email,
				UpdatedBy:   email,
				Created:     time.Now(),
				Updated:     time.Now(),
			}
		}

		limits.MaxDevices += devices
		limits.MaxNetworks += networks
		limits.MaxMembers += members
		limits.MaxServices += relays

		errs := limits.IsValid()
		if len(errs) != 0 {
			for _, err := range errs {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("limits validation error")
			}
			return
		}

		// save limits to mongodb
		mongo.Serialize(limits.Id, "id", "limits", limits)

		// construct a subscription object
		issued := time.Now()
		lu := time.Now()
		subscription := model.Subscription{
			Id:          id,
			AccountID:   account.Id,
			Email:       email,
			Name:        name,
			Description: description,
			Issued:      &issued,
			LastUpdated: &lu,
			Expires:     &expires,
			Credits:     credits,
			Sku:         sku,
			Status:      status,
		}

		errs = subscription.IsValid()
		if len(errs) != 0 {
			for _, err := range errs {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("subscription validation error")
			}
			return
		}

		// save subscription to mongodb
		mongo.Serialize(subscription.Id, "id", "subscriptions", subscription)

	}()

	c.JSON(http.StatusOK, body)
}

func readSubscription(c *gin.Context) {
	id := c.Param("id")

	client, err := core.ReadSubscription(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to read client")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, client)
}

func updateSubscription(c *gin.Context) {
	var data model.Subscription
	id := c.Param("id")

	if err := c.ShouldBindJSON(&data); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to bind")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// get update user from token and add to client infos
	oauth2Token := c.MustGet("oauth2Token").(*oauth2.Token)
	oauth2Client := c.MustGet("oauth2Client").(model.Authentication)
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

	client, err := core.UpdateSubscription(id, &data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to update client")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, client)
}

func deleteSubscription(c *gin.Context) {
	id := c.Param("id")

	err := core.DeleteSubscription(id)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to remove client")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func readSubscriptions(c *gin.Context) {
	value, exists := c.Get("oauth2Token")
	if !exists {
		c.AbortWithStatus(401)
		return
	}
	oauth2Token := value.(*oauth2.Token)
	oauth2Client := c.MustGet("oauth2Client").(model.Authentication)
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
		c.JSON(http.StatusForbidden, gin.H{"error": "This error has been logged"})
	}

	subscriptions, err := core.ReadSubscriptions(user.Email)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to list clients")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}
