package main

import (
	"log"
	"os"

	"resnetworking/services/setting"
	"resnetworking/services/user"

	"github.com/go-redis/redis"
	"github.com/urfave/cli"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := client.Ping().Result()
	if err != nil {
		log.Fatalf("failed to ping redis: %s", err)
	}
	log.Printf("[proxy] redis %s", pong)

	userDatastore := user.NewDatastore(client)
	settingDatastore := setting.NewDatastore(client)

	app := cli.NewApp()
	app.Name = "commander"
	app.Usage = "fight the loneliness!"

	app.Commands = []cli.Command{
		{
			Name:  "user",
			Usage: "manage users",
			Action: func(c *cli.Context) error {
				u, err := userDatastore.Create("", c.Args().Get(0), c.Args().Get(1))
				if err != nil {
					return err
				}

				log.Printf("[commander] created user: %s", u.Username)
				return nil
			},
		},
		{
			Name:  "domain",
			Usage: "manage domains",
			Action: func(c *cli.Context) error {
				if err := settingDatastore.Blacklist(c.Args().First()); err != nil {
					return err
				}

				log.Printf("[commander] blocked: %s", c.Args().First())
				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("failed to run cli: %s", err)
	}
}
