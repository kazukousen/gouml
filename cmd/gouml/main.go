package main

import (
	"log"
	"os"

	"github.com/kazukousen/gouml/internal/gouml"

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
				out := c.String("out")
				return gouml.Gen(baseDir, out)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "dir, d",
					Value: "./",
					Usage: "Directory you want to parse",
				},
				cli.StringFlag{
					Name:  "out, o",
					Value: "class",
					Usage: "File Name (*.uml) you want to parsed",
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
