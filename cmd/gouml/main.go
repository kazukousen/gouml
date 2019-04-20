package main

import (
	"log"
	"os"
	"path"

	"github.com/kazukousen/gouml/internal/gouml/plantuml"

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
				gen := gouml.NewGenerator(plantuml.NewParser())
				outDir := "./"
				if dirs := c.StringSlice("dir"); len(dirs) > 0 {
					outDir = dirs[0]
					for _, dir := range dirs {
						if err := gen.ReadDir(dir); err != nil {
							return err
						}
					}
				}

				if files := c.StringSlice("file"); len(files) > 0 {
					for _, path := range files {
						if err := gen.Read(path); err != nil {
							return err
						}
					}
				}

				out := c.String("out")
				return gen.OutputFile(path.Join(outDir, out))
			},
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "dir, d",
					Usage: "Directory you want to parse",
				},
				cli.StringSliceFlag{
					Name: "file, f",
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
