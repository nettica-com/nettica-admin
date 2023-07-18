package core

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"time"

	model "github.com/nettica-com/nettica-admin/model"
	mongo "github.com/nettica-com/nettica-admin/mongo"
	template "github.com/nettica-com/nettica-admin/template"
	util "github.com/nettica-com/nettica-admin/util"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"gopkg.in/gomail.v2"
)

// CreateHost host with all necessary data
func CreateHost(host *model.Host) (*model.Host, error) {

	u := uuid.NewV4()
	host.Id = u.String()

	if host.HostGroup == "" {
		host.HostGroup = host.Id
	}

	var err error
	if host.APIKey == "" {
		host.APIKey, err = util.RandomString(32)
		if err != nil {
			return nil, err
		}
	}

	// read the meshes and configure the default values
	meshes, err := ReadMeshes(host.CreatedBy)
	if err != nil {
		return nil, err
	}

	for _, mesh := range meshes {
		if mesh.MeshName == host.MeshName {
			host.Default = mesh.Default
			current := host.Current
			host.Current = mesh.Default
			host.Current.ListenPort = current.ListenPort
			host.Current.Endpoint = current.Endpoint
			host.Current.PrivateKey = current.PrivateKey
			host.Current.PublicKey = current.PublicKey
			host.Current.PostUp = current.PostUp
			host.Current.PostDown = current.PostDown
			host.Current.PersistentKeepalive = current.PersistentKeepalive
			host.MeshId = mesh.Id
			host.AccountId = mesh.AccountId
			host.Current.AllowedIPs = current.AllowedIPs
			host.Current.Dns = mesh.Default.Dns
		}
	}

	// if the host data already has a public key and empty private key,
	// we know the client has already generated a key pair
	if host.Current.PublicKey != "" && host.Current.PrivateKey == "" {
		log.Info("client has already generated a key pair")
	} else {
		// generate a new key pair
		log.Info("generating a new key pair")
		key, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			return nil, err
		}
		host.Current.PrivateKey = key.String()
		host.Current.PublicKey = key.PublicKey().String()
	}

	reserverIps, err := GetAllReservedMeshIps(host.MeshId)
	if err != nil {
		return nil, err
	}

	ips := make([]string, 0)
	ipsDns := make([]string, 0)
	for _, network := range host.Default.Address {
		ip, err := util.GetAvailableIp(network, reserverIps)
		if err != nil {
			return nil, err
		}
		ipsDns = append(ipsDns, ip)
		if util.IsIPv6(ip) {
			ip = ip + "/128"
		} else {
			ip = ip + "/32"
		}
		ips = append(ips, ip)
	}
	host.Current.Address = ips
	host.Current.AllowedIPs = append(host.Current.AllowedIPs, ips...)
	if host.Current.EnableDns {
		host.Current.Dns = append(host.Current.Dns, ipsDns...)
	}

	if host.Current.SubnetRouting && len(host.Current.PostUp) == 0 {
		host.Current.PostUp = fmt.Sprintf("iptables -A FORWARD -i %s -j ACCEPT; iptables -A FORWARD -o %s -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE", host.MeshName, host.MeshName)
	}
	if host.Current.SubnetRouting && len(host.Current.PostDown) == 0 {
		host.Current.PostDown = fmt.Sprintf("iptables -D FORWARD -i %s -j ACCEPT; iptables -D FORWARD -o %s -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE", host.MeshName, host.MeshName)
	}

	host.Created = time.Now().UTC()
	host.Updated = host.Created

	// check if host is valid
	errs := host.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("host validation error")
		}
		return nil, errors.New("failed to validate host")
	}

	err = mongo.Serialize(host.Id, "id", "hosts", host)
	if err != nil {
		return nil, err
	}

	v, err := mongo.Deserialize(host.Id, "id", "hosts", reflect.TypeOf(model.Host{}))
	if err != nil {
		return nil, err
	}
	host = v.(*model.Host)

	// data modified, dump new config
	return host, nil
}

// GetAllReservedIps the list of all reserved IPs, client and server
func GetAllReservedMeshIps(meshId string) ([]string, error) {
	clients, err := mongo.ReadAllHosts("meshid", meshId)

	if err != nil {
		return nil, err
	}
	reserverIps := make([]string, 0)

	for _, client := range clients {
		if client.MeshId == meshId {
			for _, cidr := range client.Current.Address {
				ip, err := util.GetIpFromCidr(cidr)
				if err != nil {
					log.WithFields(log.Fields{
						"err":  err,
						"cidr": cidr,
					}).Error("failed to ip from cidr")
				} else {
					reserverIps = append(reserverIps, ip)
				}
			}
		}
	}

	return reserverIps, nil
}

// ReadHost host by id
func ReadHost(id string) (*model.Host, error) {
	v, err := mongo.Deserialize(id, "id", "hosts", reflect.TypeOf(model.Host{}))
	if err != nil {
		return nil, err
	}
	host := v.(*model.Host)

	return host, nil
}

// UpdateHost preserve keys
func UpdateHost(Id string, host *model.Host, flag bool) (*model.Host, error) {
	v, err := mongo.Deserialize(Id, "id", "hosts", reflect.TypeOf(model.Host{}))
	if err != nil {
		return nil, err
	}
	current := v.(*model.Host)

	if current.Id != host.Id {
		return nil, errors.New("records Id mismatch")
	}

	if len(host.Current.Address) == 0 ||
		(len(host.Default.Address) > 0 && len(current.Default.Address) > 0 &&
			(host.Default.Address[0] != current.Default.Address[0])) {
		reserverIps, err := GetAllReservedMeshIps(host.MeshId)
		if err != nil {
			return nil, err
		}

		ips := make([]string, 0)
		ipsDns := make([]string, 0)

		for _, network := range host.Default.Address {
			ip, err := util.GetAvailableIp(network, reserverIps)
			if err != nil {
				return nil, err
			}
			ipsDns = append(ipsDns, ip)
			if util.IsIPv6(ip) {
				ip = ip + "/128"
			} else {
				ip = ip + "/32"
			}
			ips = append(ips, ip)
		}
		host.Current.Address = ips
		host.Current.AllowedIPs = ips
		if host.Current.EnableDns {
			host.Current.Dns = append(host.Current.Dns, ipsDns...)
		}
	}

	if host.Current.SubnetRouting && len(host.Current.PostUp) == 0 {
		host.Current.PostUp = fmt.Sprintf("iptables -A FORWARD -i %s -j ACCEPT; iptables -A FORWARD -o %s -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE", host.MeshName, host.MeshName)
	}
	if host.Current.SubnetRouting && len(host.Current.PostDown) == 0 {
		host.Current.PostDown = fmt.Sprintf("iptables -D FORWARD -i %s -j ACCEPT; iptables -D FORWARD -o %s -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE", host.MeshName, host.MeshName)
	}

	// check if host is valid
	errs := host.IsValid()
	if len(errs) != 0 {
		for _, err := range errs {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("host validation error")
		}
		return nil, errors.New("failed to validate host")
	}

	if !flag {
		host.Updated = time.Now().UTC()
	}

	err = mongo.Serialize(host.Id, "id", "hosts", host)
	if err != nil {
		return nil, err
	}

	v, err = mongo.Deserialize(Id, "id", "hosts", reflect.TypeOf(model.Host{}))
	if err != nil {
		return nil, err
	}
	host = v.(*model.Host)

	// data modified, dump new config
	return host, nil
}

// DeleteHost from database
func DeleteHost(id string) error {

	return mongo.DeleteHost(id, "hosts")
}

// ReadHost2 host by param and id
func ReadHost2(param string, id string) ([]*model.Host, error) {
	return mongo.ReadAllHosts(param, id)
}

// ReadHosts all hosts
func ReadHosts() ([]*model.Host, error) {
	return mongo.ReadAllHosts("", "")
}

// ReadHosts all hosts
func ReadHostsForUser(email string) ([]*model.Host, error) {
	accounts, err := mongo.ReadAllAccounts(email)

	results := make([]*model.Host, 0)

	for _, account := range accounts {
		if account.Status == "Active" {

			if account.MeshId != "" {
				hosts, err := mongo.ReadAllHosts("meshid", account.MeshId)
				if err != nil {
					return nil, err
				}
				results = append(results, hosts...)

			} else {
				hosts, err := mongo.ReadAllHosts("accountid", account.Parent)
				if err != nil {
					return nil, err
				}
				results = append(results, hosts...)
			}

		}
	}

	return results, err
}

// ReadHostConfig in wg format
func ReadHostConfig(id string) ([]byte, *string, error) {

	meshName := ""
	host, err := ReadHost(id)
	if err != nil {
		return nil, nil, err
	}
	hosts, err := ReadHost2("meshid", host.MeshId)
	if err != nil {
		return nil, nil, err
	}

	index := 0
	for j := 0; j < len(hosts); j++ {
		if hosts[j].Id == id {
			index = j
			break
		}
	}

	if index == -1 {
		log.Errorf("Error reading Mesh: %v", hosts)
	} else {
		host := hosts[index]
		meshName = host.MeshName
		hosts = append(hosts[:index], hosts[index+1:]...)

		for i := 0; i < len(hosts); i++ {
			// if the current host doesn't have an endpoint specified it is a client, so it does not
			// need the public keys of other clients since they can't connect to each other.  If there
			// is an endpoint specified, keep all the clients in the config.
			if host.Current.Endpoint == "" && hosts[i].Current.Endpoint == "" {
				hosts = append(hosts[:i], hosts[i+1:]...)
				i--
			}
		}

		config, err := template.DumpWireguardConfig(host, hosts)
		if err != nil {
			return nil, nil, err
		}

		return config, &meshName, nil
	}
	return nil, nil, err
}

// EmailHost send email to host
func EmailHost(id string) error {

	host, err := ReadHost(id)
	if err != nil {
		return err
	}

	configData, _, err := ReadHostConfig(id)
	if err != nil {
		return err
	}

	// conf as .conf file
	tmpfileCfg, err := ioutil.TempFile("", "wireguard-vpn-*.conf")
	if err != nil {
		return err
	}
	if _, err := tmpfileCfg.Write(configData); err != nil {
		return err
	}
	if err := tmpfileCfg.Close(); err != nil {
		return err
	}
	defer os.Remove(tmpfileCfg.Name()) // clean up

	// conf as png image
	png, err := qrcode.Encode(string(configData), qrcode.Medium, 280)
	if err != nil {
		return err
	}
	tmpfilePng, err := ioutil.TempFile("", "qrcode-*.png")
	if err != nil {
		return err
	}
	if _, err := tmpfilePng.Write(png); err != nil {
		return err
	}
	if err := tmpfilePng.Close(); err != nil {
		return err
	}
	defer os.Remove(tmpfilePng.Name()) // clean up

	// get email body
	emailBody, err := template.DumpEmail(host, filepath.Base(tmpfilePng.Name()))
	if err != nil {
		return err
	}

	// port to int
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return err
	}

	d := gomail.NewDialer(os.Getenv("SMTP_HOST"), port, os.Getenv("SMTP_USERNAME"), os.Getenv("SMTP_PASSWORD"))
	s, err := d.Dial()
	if err != nil {
		return err
	}
	m := gomail.NewMessage()

	m.SetHeader("From", os.Getenv("SMTP_FROM"))
	m.SetAddressHeader("To", host.Email, host.Name)
	m.SetHeader("Subject", "WireGuard VPN Configuration")
	m.SetBody("text/html", string(emailBody))
	m.Attach(tmpfileCfg.Name())
	m.Embed(tmpfilePng.Name())

	err = gomail.Send(s, m)
	if err != nil {
		return err
	}

	return nil
}
