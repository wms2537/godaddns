package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/crypto"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/joho/godotenv"
)

type DDNSRequest struct {
	NodeId    string `json:"nodeId" binding:"required"`
	NewIP     string `json:"newIP" binding:"required"`
	Signature string `json:"signature" binding:"required"`
	PubKey    string `json:"pubKey" binding:"required"`
}

func parseKeys(privateKeyString string, password string) (cryptotypes.PrivKey, cryptotypes.PubKey, sdktypes.AccAddress, error) {
	// Create a private key object
	key, _, err := crypto.UnarmorDecryptPrivKey(privateKeyString, password)
	if err != nil {
		return nil, nil, nil, err
	}
	pub := key.PubKey()
	addr := sdktypes.AccAddress(pub.Address().Bytes())
	return key, pub, addr, nil
}

func getOwnIPv6(ipProvider string) (string, error) {
	resp, err := http.Get(ipProvider)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	return buf.String(), nil
}

func checkDNS(domain string, expectedIPv6 string) (bool, error) {
	// Resolve the IPv6 address (AAAA record) for the domain.
	ips, err := net.LookupIP(domain)
	if err != nil {
		fmt.Println("Error looking up IPv6 address:", err)
		return false, err
	}

	// Find the IPv6 address in the resolved addresses.
	var resolvedIPv6 string
	for _, ip := range ips {
		if ip.To4() == nil {
			resolvedIPv6 = ip.String()
			break
		}
	}

	if resolvedIPv6 == "" {
		fmt.Println("No IPv6 address found for the domain.")
		return true, nil
	}

	// Compare the resolved IPv6 address with the expected IPv6 address.
	if resolvedIPv6 == expectedIPv6 {
		return false, nil
	} else {
		fmt.Printf("DNS record (IPv6) needs to be updated. Current IPv6: %s, Expected IPv6: %s\n", resolvedIPv6, expectedIPv6)
		return true, nil
	}
}

func main() {
	// Define the URL of the DDNS server.

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	SERVER_URL := os.Getenv("SERVER_URL")
	NODE_ID := os.Getenv("NODE_ID")
	PASSWORD := os.Getenv("PASSWORD")
	IP_PROVIDER := os.Getenv("IP_PROVIDER")
	PRIV_KEY_PATH := os.Getenv("PRIV_KEY_PATH")
	BECH32_PREFIX := os.Getenv("BECH32_PREFIX")

	sdktypes.GetConfig().SetBech32PrefixForAccount(BECH32_PREFIX, BECH32_PREFIX+"pub")

	privateKeyBytes, err := os.ReadFile(PRIV_KEY_PATH)
	if err != nil {
		fmt.Println("Error reading private key file:", err)
		return
	}
	fmt.Println(string(privateKeyBytes))
	privKey, pubKey, _, err := parseKeys(string(privateKeyBytes), PASSWORD)
	if err != nil {
		fmt.Println("Error parsing keys:", err)
		return
	}
	serverURL := SERVER_URL + "/update-dns"

	signature, err := privKey.Sign([]byte(BECH32_PREFIX))
	if err != nil {
		fmt.Println("Error signing message:", err)
		return
	}

	interval := 30 * time.Second

	// Use a loop to execute the function at regular intervals.
	for {
		ownIP, err := getOwnIPv6(IP_PROVIDER)
		if err != nil {
			log.Fatal(err)
		}

		needsUpdate, err := checkDNS(NODE_ID+".beautifood.io", ownIP)
		if err != nil {
			log.Fatal(err)
		}
		if needsUpdate {
			// Define the DDNS request.
			request := DDNSRequest{
				NodeId:    NODE_ID,
				NewIP:     ownIP,
				Signature: base64.StdEncoding.EncodeToString(signature),
				PubKey:    crypto.ArmorPubKeyBytes(legacy.Cdc.Amino.MustMarshalBinaryBare(pubKey), ""),
			}

			// Serialize the request to JSON.
			requestData, err := json.Marshal(request)
			if err != nil {
				log.Fatal(err)
				return
			}

			// Send a POST request to the DDNS server.
			resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(requestData))
			if err != nil {
				log.Fatal(err)
				return
			}
			defer resp.Body.Close()

			// Read the response from the server.
			response, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
				return
			}
			// Print the response from the server.
			fmt.Printf("Server Response: %s\n", string(response))
		}

		time.Sleep(interval) // Sleep for the specified interval.
	}
}
