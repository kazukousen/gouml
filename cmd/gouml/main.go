package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/kazukousen/gouml"
	"github.com/urfave/cli"
)

func main() {
	logger := log.NewLogfmtLogger(os.Stdout)
	logger = log.With(logger, "ts", log.DefaultTimestamp)

	flags := []cli.Flag{
		&cli.StringSliceFlag{
			Name:  "file, f",
			Usage: "File or Directory you want to parse",
		},
		&cli.StringSliceFlag{
			Name:  "ignore, I",
			Usage: "File or Directory you want to ignore parsing",
		},
		&cli.BoolFlag{
			Name:  "verbose",
			Usage: "debugging",
		},
	}
	app := cli.NewApp()
	app.Version = "0.2"
	app.Usage = "Automatically generate PlantUML from Go Code."
	app.Commands = []cli.Command{
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "Create *.puml",
			Action: func(c *cli.Context) error {
				buf := &bytes.Buffer{}
				buf.WriteString("@startuml\n")
				if err := generate(logger, buf, c.StringSlice("ignore"), c.StringSlice("file"), c.Bool("verbose")); err != nil {
					return err
				}
				buf.WriteString("@enduml\n")

				out := c.String("out")
				out, err := filepath.Abs(out)
				if err != nil {
					return err
				}
				if err := writeFile(out, buf); err != nil {
					return err
				}
				fmt.Printf("output to file: %s\n", out)
				return nil
			},
			Flags: append(flags, []cli.Flag{
				&cli.StringFlag{
					Name:  "out, o",
					Value: "file.puml",
					Usage: "File Name you want to parsed",
				},
			}...),
		},
		{
			Name:    "encode",
			Aliases: []string{"e"},
			Usage:   "encode base64",
			Action: func(c *cli.Context) error {
				buf := &bytes.Buffer{}
				if err := generate(logger, buf, c.StringSlice("ignore"), c.StringSlice("file"), c.Bool("verbose")); err != nil {
					return err
				}

				fmt.Printf(gouml.Compress(buf.String()))
				return nil
			},
			Flags: append(flags, []cli.Flag{}...),
		},
	}

	if err := app.Run(os.Args); err != nil {
		level.Error(logger).Log("msg", "failed to run", "error", err)
	}
}

func generate(logger log.Logger, buf *bytes.Buffer, ignores []string, targets []string, verbose bool) error {
	gen := gouml.NewGenerator(logger, gouml.PlantUMLParser(logger), verbose)
	if len(ignores) > 0 {
		if err := gen.UpdateIgnore(ignores); err != nil {
			return err
		}
	}
	if len(targets) == 0 {
		targets = []string{"./"}
	}
	if err := gen.Read(targets); err != nil {
		return err
	}

	gen.WriteTo(buf)
	return nil
}

func writeFile(file string, buf io.Reader) (e error) {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			e = err
		}
	}()
	io.Copy(f, buf)
	return
}
