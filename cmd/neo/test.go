package main

import (
	"context"

	core "github.com/eveisesi/neo/app"
	"github.com/urfave/cli"
)

func test() cli.Command {
	return cli.Command{
		Name: "test",
		// Action: func(c *cli.Context) error {
		// 	app := core.New("f", false)
		// 	var killmails = make([]*neo.Killmail, 0)

		// 	mods := []neo.Modifier{
		// 		neo.ColValOr{
		// 			Values: []neo.Modifier{
		// 				neo.EqualTo{Column: "victim.characterID", Value: 92068157},
		// 				neo.EqualTo{Column: "attackers.characterID", Value: 92068157},
		// 			},
		// 		},
		// 	}

		// 	filters := mdb.BuildFilters(mods...)
		// 	// spew.Dump(filters)
		// 	// return nil

		// 	result, err := app.MongoDB.Collection("killmails").Find(
		// 		context.TODO(),
		// 		filters,
		// 		options.Find().SetLimit(5),
		// 	)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	err = result.All(context.TODO(), &killmails)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	spew.Config.MaxDepth = 3
		// 	spew.Dump(killmails)

		// 	return nil
		// },

		Action: func(c *cli.Context) error {
			app := core.New("f", false)
			// "/v1/killmails/86837511/f08a9f72f4095726d3bccdeb6f9a1273d24e28c2/"
			_, m := app.ESI.GetKillmailsKillmailIDKillmailHash(context.Background(), 86837511, "f08a9f72f4095726d3bccdeb6f9a1273d24e28c2")
			if m.IsErr() {
				return m.Msg
			}
			return nil
		},
	}
}
