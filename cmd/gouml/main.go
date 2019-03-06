package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kazukousen/gouml"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Version = "0.0.1"
	app.Usage = "Automatically generate PlantUML from Go Code."
	app.Commands = []cli.Command{
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "Create *.uml",
			Action: func(c *cli.Context) error {
				baseDir := c.String("dir")
				return gouml.NewRunner().Run(baseDir)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "dir, d",
					Value: "./",
					Usage: "Directory you want to parse",
				},
			},
		},
		{
			Name:    "hello",
			Aliases: []string{"h"},
			Usage:   "Create *.uml",
			Action: func(c *cli.Context) error {
				baseDir := c.String("dir")
				fmt.Println(baseDir)
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "dir, d",
					Value: "./",
					Usage: "Directory you want to parse",
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
