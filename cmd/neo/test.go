package main

import (
	"github.com/urfave/cli"
)

func test() cli.Command {
	return cli.Command{
		Name: "test",
		Action: func(c *cli.Context) error {
			// app := core.New("f", false)

			return nil
		},
	}
}
