package main

import (
	"log"
	"os"

	"godaddns/ddns"
	"godaddns/storage"

	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
)

func main() {
	var isCli bool
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "whitelist",
				Usage:       "add whitelist",
				Destination: &isCli,
			},
			&cli.StringFlag{
				Name:  "addr",
				Usage: "User Address",
			},
			&cli.StringFlag{
				Name:  "node-id",
				Usage: "Node ID",
			},
		},
		Action: func(cCtx *cli.Context) error {
			if isCli {
				if err := storage.AddUserToWhitelist(cCtx.String("addr"), cCtx.String("node-id")); err != nil {
					log.Fatal(err)
				}
			} else {
				r := gin.Default()

				// Use your custom authMiddleware to handle authentication and authorization.
				r.POST("/update-dns", ddns.UpdateDNSHandler)

				err := r.Run(":8080")
				if err != nil {
					log.Fatal(err)
				}
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
