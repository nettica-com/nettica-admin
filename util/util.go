package util

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"math/big"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	// AuthTokenHeaderName http header for token transport
	AuthTokenHeaderName = "Authorization"
	// RegexpEmail check valid email
	RegexpEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

func GetCleanAuthToken(c *gin.Context) string {
	token := c.Request.Header.Get(AuthTokenHeaderName)
	if len(token) > 0 && token[:7] == "Bearer " {
		token = token[7:]
		token = strings.Trim(token, "\"")
	}
	return token
}

// ReadFile file content
func ReadFile(path string) (bytes []byte, err error) {
	bytes, err = os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// WriteFile content to file
func WriteFile(path string, bytes []byte) (err error) {
	err = os.WriteFile(path, bytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

// FileExists check if file exists
func FileExists(name string) bool {
	info, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// DirectoryExists check if directory exists
func DirectoryExists(name string) bool {
	info, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// GetNetworkAddress gets the valid start of a subnet
func GetNetworkAddress(cidr string) (string, error) {

	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}
	networkAddr := ipnet.String()

	return networkAddr, nil

}

// GetAvailableCidr search for an available ip in cidr against a list of reserved ips
func GetAvailableCidr(cidr string, reserved []string) (string, error) {

	parts := strings.Split(cidr, "/")
	if len(parts) != 2 {
		return "", errors.New("invalid cidr")
	}

	prefix := parts[1]

	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}

	// these two addresses are not usable
	broadcastAddr := BroadcastAddr(ipnet).String()
	networkAddr := ipnet.IP.String()

	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ok := true
		address := ip.String()
		for _, r := range reserved {
			if address == r {
				ok = false
				break
			}
		}
		if ok && address != networkAddr && address != broadcastAddr {
			return address + "/" + prefix, nil
		}
	}

	return "", errors.New("no more available address from cidr")
}

// IsIPv6 check if given ip is IPv6
func IsIPv6(address string) bool {
	ip := net.ParseIP(address)
	if ip == nil {
		return false
	}
	return ip.To4() == nil
}

// IsValidIp check if ip is valid
func IsValidIp(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsValidCidr check if CIDR is valid
func IsValidCidr(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	return err == nil
}

// IsInCidr check if cidr is in subnet (also a cidr, eg, 10.0.0.1/32 is in 10.0.0.0/24
func IsInCidr(cidr string, subnet string) bool {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return false
	}
	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return ipnet.Contains(ip)
}

// GetIpFromCidr get ip from cidr
func GetIpFromCidr(cidr string) (string, error) {
	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}
	return ip.String(), nil
}

// http://play.golang.org/p/m8TNTtygK0
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// BroadcastAddr returns the last address in the given network, or the broadcast address.
func BroadcastAddr(n *net.IPNet) net.IP {
	// The golang net package doesn't make it easy to calculate the broadcast address. :(
	var broadcast net.IP
	if len(n.IP) == 4 {
		broadcast = net.ParseIP("0.0.0.0").To4()
	} else {
		broadcast = net.ParseIP("::")
	}
	for i := 0; i < len(n.IP); i++ {
		broadcast[i] = n.IP[i] | ^n.Mask[i]
	}
	return broadcast
}

// Compares two arrays for equivalence
func CompareArrays(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, x := range a {
		found := false
		for _, y := range b {
			if x == y {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

// RandomString returns a securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func RandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}
