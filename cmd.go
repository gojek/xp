package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var version = "0.1.0"

func main() {
	var d *data

	app := cli.NewApp()
	app.Name = "xp"
	app.HelpName = "xp"
	app.Version = version
	app.Usage = "extreme programming made simple"

	xpConfig, err := homedir.Expand("~/.xp")
	if err != nil {
		log.Fatal(err)
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: xpConfig,
		},
	}

	app.Before = func(c *cli.Context) error {
		cfg := c.String("config")

		f, err := os.Open(cfg)
		if err != nil {
			if os.IsNotExist(err) {
				d = new(data)
				return nil
			}
			return errors.Wrapf(err, "could not open %s", cfg)
		}
		defer f.Close()

		d, err = load(f)
		if err != nil {
			return errors.Wrap(err, "load failed")
		}

		return nil
	}

	app.After = func(c *cli.Context) error {
		cfg := c.String("config")

		f, err := os.OpenFile(cfg, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return errors.Wrapf(err, "could not open %s for writing", cfg)
		}
		defer f.Close()

		if err := d.store(f); err != nil {
			return errors.Wrap(err, "store failed")
		}

		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:    "show-config",
			Aliases: []string{"sc"},
			Action: func(c *cli.Context) error {
				fmt.Print(d)
				return nil
			},
		},
		{
			Name: "add-info",
			Action: func(c *cli.Context) error {
				wd, err := os.Getwd()
				if err != nil {
					return errors.Wrap(err, "could not get wd")
				}

				msgFile := c.Args().Get(0)

				d.appendInfo(wd, msgFile)

				return nil
			},
		},
		{
			Name:    "dev",
			Aliases: []string{"d"},
			Subcommands: []cli.Command{
				{
					Name:    "add",
					Aliases: []string{"a"},
					Action: func(c *cli.Context) error {
						args := c.Args()

						id, name, email := args.Get(0), args.Get(1), args.Get(2)
						if id == "" || name == "" || email == "" {
							return errors.New("invalid id/name/email")
						}

						d.addDev(id, name, email)
						return nil
					},
				},
			},
		},
		{
			Name:    "repo",
			Aliases: []string{"r"},
			Subcommands: []cli.Command{
				{
					Name:    "init",
					Aliases: []string{"i"},
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name: "overwrite",
						},
						cli.StringSliceFlag{
							Name: "devs",
						},
						cli.StringFlag{
							Name: "story-id",
						},
					},
					Action: func(c *cli.Context) error {
						args := c.Args()

						dir := args.Get(0)
						if dir == "." {
							var err error
							dir, err = os.Getwd()
							if err != nil {
								return errors.Wrap(err, "could not get wd")
							}
						}

						overwrite := c.Bool("overwrite")
						if err := initRepo(dir, overwrite); err != nil {
							return errors.Wrap(err, "repo .git hook init failed")
						}

						devs := c.StringSlice("devs")
						storyID := c.String("story-id")

						if err := d.addRepo(dir, devs, storyID); err != nil {
							return errors.Wrap(err, "could add init repo")
						}

						return nil
					},
				},
				{
					Name:    "devs",
					Aliases: []string{"d"},
					Action: func(c *cli.Context) error {
						wd, err := os.Getwd()
						if err != nil {
							return errors.Wrap(err, "could not get wd")
						}
						devs := c.Args()

						if err := d.updateRepoDevs(wd, devs); err != nil {
							return errors.Wrap(err, "could not set devs")
						}

						return nil
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
