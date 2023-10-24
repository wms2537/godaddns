package main

import (
	"log"

	"godaddns/ddns"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Use your custom authMiddleware to handle authentication and authorization.
	r.POST("/update-dns", ddns.UpdateDNSHandler)

	// if err := storage.AddUserToWhitelist("jomtx1ls2que64qf5s5lrlw2zq2hqj3ts3kk43mks9zh", "1be9f909932fc0c7c5685439f0bd109ced722114"); err != nil {
	// 	log.Fatal(err)
	// }

	err := r.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
