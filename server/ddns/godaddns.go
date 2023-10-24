package ddns

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"godaddns/storage"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/crypto"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var GODADDY_KEY string
var GODADDY_SECRET string
var DOMAIN string

type Record struct {
	Data     string `json:"data"`
	Port     int64  `json:"port"`
	Priority int64  `json:"priority"`
	Protocol string `json:"protocol"`
	Service  string `json:"service"`
	TTL      int64  `json:"ttl"`
	Weight   int64  `json:"weight"`
}

func init() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	GODADDY_KEY = os.Getenv("GODADDY_KEY")
	GODADDY_SECRET = os.Getenv("GODADDY_SECRET")
	DOMAIN = os.Getenv("DOMAIN")

	sdktypes.GetConfig().SetBech32PrefixForAccount("jomtx", "jomtxpub")
}

func GetDomainIPv6(subdomain string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.godaddy.com/v1/domains/%s/records/AAAA/%s", DOMAIN, subdomain), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", GODADDY_KEY, GODADDY_SECRET))
	c := new(http.Client)
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	in := make([]struct {
		Data string `json:"data"`
	}, 1)
	json.NewDecoder(resp.Body).Decode(&in)
	if len(in) == 0 {
		return "", nil
	}
	return in[0].Data, nil
}

func PutNewIP(ip string, subdomain string) error {
	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode([]Record{{
		strings.TrimSuffix(ip, "\n"),
		65535,
		0,
		"",
		"",
		600,
		0,
	}})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT",
		fmt.Sprintf("https://api.godaddy.com/v1/domains/%s/records/AAAA/%s", DOMAIN, subdomain),
		&buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", GODADDY_KEY, GODADDY_SECRET))
	c := new(http.Client)
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == 200 {
		return nil
	} else {
		return fmt.Errorf("failed with HTTP status code %d", resp.StatusCode)
	}
}
func UpdateDNSHandler(c *gin.Context) {

	// Extract the new DNS data from the request.
	var dnsData struct {
		NodeId    string `json:"nodeId" binding:"required"`
		NewIP     string `json:"newIP" binding:"required"`
		Signature string `json:"signature" binding:"required"`
		PubKey    string `json:"pubKey" binding:"required"`
	}
	if err := c.ShouldBindJSON(&dnsData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	signatureValid, addr, err := verifySignature(dnsData.PubKey, dnsData.Signature)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if !signatureValid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	isWhitelist, err := storage.IsUserNodeInWhitelist(addr, dnsData.NodeId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !isWhitelist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorised"})
		return
	}

	domainIP, err := GetDomainIPv6(dnsData.NodeId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if domainIP == "" || domainIP != dnsData.NewIP {
		if err := PutNewIP(dnsData.NewIP, dnsData.NodeId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "DNS record updated successfully"})
}

func verifySignature(pubKeyString string, signatureString string) (bool, string, error) {
	fmt.Println(pubKeyString)
	pubBytes, _, err := crypto.UnarmorPubKeyBytes(pubKeyString)
	if err != nil {
		return false, "", err
	}

	pub, err := legacy.PubKeyFromBytes(pubBytes)
	if err != nil {
		return false, "", err
	}

	signature, err := base64.StdEncoding.DecodeString(signatureString)
	if err != nil {
		return false, "", err
	}
	if !pub.VerifySignature([]byte("jomtx"), signature) {
		return false, "", fmt.Errorf("failed to verify signature")
	}

	addr := sdktypes.AccAddress(pub.Address().Bytes())

	return true, addr.String(), nil
}
