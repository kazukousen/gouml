package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kazukousen/gouml"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Version = "0.2"
	app.Usage = "Automatically generate PlantUML from Go Code."
	app.Commands = []cli.Command{
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "Create *.uml",
			Action: func(c *cli.Context) error {
				gen := gouml.NewGenerator(gouml.PlantUMLParser(), c.Bool("verbose"))

				if ignoreFiles := c.StringSlice("ignore"); len(ignoreFiles) > 0 {
					if err := gen.UpdateIgnore(ignoreFiles); err != nil {
						return err
					}
				}

				if files := c.StringSlice("file"); len(files) > 0 {
					if err := gen.Read(files); err != nil {
						return err
					}
				}

				out, err := filepath.Abs(c.String("out"))
				if err != nil {
					return err
				}

				buf := &bytes.Buffer{}
				gen.WriteTo(buf)

				uml, err := os.Create(out)
				if err != nil {
					return err
				}
				defer uml.Close()
				fmt.Fprintf(uml, buf.String())
				fmt.Printf("output to file: %s\n", out)

				fmt.Printf("SVG: http://plantuml.com/plantuml/svg/%s\n", gouml.Compress(buf.String()))
				return nil
			},
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "file, f",
					Value: &cli.StringSlice{"./"},
					Usage: "File or Directory you want to parse",
				},
				cli.StringSliceFlag{
					Name:  "ignore, I",
					Usage: "File or Directory you want to ignore parsing",
				},
				cli.StringFlag{
					Name:  "out, o",
					Value: "class.uml",
					Usage: "File Name you want to parsed",
				},
				cli.BoolFlag{
					Name:  "verbose",
					Usage: "debugging",
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
