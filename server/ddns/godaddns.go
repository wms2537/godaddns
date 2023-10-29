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
var BECH32_PREFIX string

type Record struct {
	Data     string `json:"data"`
	Name     string `json:"name"`
	Port     int64  `json:"port"`
	Priority int64  `json:"priority"`
	Protocol string `json:"protocol"`
	Service  string `json:"service"`
	TTL      int64  `json:"ttl"`
	Type     string `json:"type"`
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
	BECH32_PREFIX = os.Getenv("BECH32_PREFIX")

	sdktypes.GetConfig().SetBech32PrefixForAccount(BECH32_PREFIX, BECH32_PREFIX+"pub")
}

func GetDomainIPv6(subdomain string) ([]*Record, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.godaddy.com/v1/domains/%s/records/AAAA", DOMAIN), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", GODADDY_KEY, GODADDY_SECRET))
	c := new(http.Client)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	var in []Record
	json.NewDecoder(resp.Body).Decode(&in)
	if len(in) == 0 {
		return nil, nil
	}
	var result []*Record
	for _, v := range in {
		if v.Type == "AAAA" {
			result = append(result, &Record{
				v.Data,
				v.Name,
				v.Port,
				v.Priority,
				v.Protocol,
				v.Service,
				v.TTL,
				v.Type,
				v.Priority,
			})
		}
	}
	return result, nil
}

func PutNewIP(records []*Record) error {
	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(records)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT",
		fmt.Sprintf("https://api.godaddy.com/v1/domains/%s/records/AAAA", DOMAIN),
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
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed with HTTP status code %d", resp.StatusCode)
	}
	fmt.Println("Updated domain")
	return nil
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

	domainIPv6Records, err := GetDomainIPv6(dnsData.NodeId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	needsUpdate := false
	domainExists := false
	wildcardExists := false
	for _, v := range domainIPv6Records {
		if v.Name == dnsData.NodeId {
			domainExists = true
			if v.Data != dnsData.NewIP {
				v.Data = dnsData.NewIP
				needsUpdate = true
				fmt.Printf("Domain for %s needs update\n", v.Name)
			}
		} else if v.Name == "*."+dnsData.NodeId {
			wildcardExists = true
			if v.Data != dnsData.NewIP {
				v.Data = dnsData.NewIP
				needsUpdate = true
				fmt.Printf("Wildcard for %s needs update\n", v.Name)
			}
		}
		v.Port = 65535
	}
	if !domainExists {
		fmt.Printf("Domain for %s needs create\n", dnsData.NodeId)
		needsUpdate = true
		domainIPv6Records = append(domainIPv6Records, &Record{
			strings.TrimSuffix(dnsData.NewIP, "\n"),
			dnsData.NodeId,
			65535,
			0,
			"",
			"",
			600,
			"AAAA",
			0,
		})
	}
	if !wildcardExists {
		fmt.Printf("Wildcard for %s needs create\n", dnsData.NodeId)
		needsUpdate = true
		domainIPv6Records = append(domainIPv6Records, &Record{
			strings.TrimSuffix(dnsData.NewIP, "\n"),
			"*." + dnsData.NodeId,
			65535,
			0,
			"",
			"",
			600,
			"AAAA",
			0,
		})
	}
	if needsUpdate {
		if err := PutNewIP(domainIPv6Records); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "DNS record updated successfully"})
}

func verifySignature(pubKeyString string, signatureString string) (bool, string, error) {
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
	if !pub.VerifySignature([]byte(BECH32_PREFIX), signature) {
		return false, "", fmt.Errorf("failed to verify signature")
	}

	addr := sdktypes.AccAddress(pub.Address().Bytes())

	return true, addr.String(), nil
}
