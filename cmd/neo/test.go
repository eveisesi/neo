package main

import (
	"github.com/urfave/cli"
)

func test() cli.Command {
	return cli.Command{}
	// cli.Command{
	// 	Name: "test",
	// 	Action: func(c *cli.Context) error {
	// 		app := core.New("f", false)
	// 		current := time.Date(2020, 9, 11, 0, 0, 0, 0, time.UTC)

	// 		ctx := context.Background()

	// 		countKillmailsFilters := []neo.Modifier{
	// 			neo.GreaterThanEqualTo{Column: "killmailTime", Value: time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, time.UTC)},
	// 			neo.LessThanEqualTo{Column: "killmailTime", Value: time.Date(current.Year(), current.Month(), current.Day(), 23, 59, 59, 0, time.UTC)},
	// 		}

	// 		kr := mdb.NewKillmailRepository(app.MongoDB)

	// 		localKillmails, err := kr.Killmails(ctx, countKillmailsFilters...)
	// 		if err != nil {
	// 			app.Logger.WithError(err).Error("encountered error querying killmail count for date")
	// 		}

	// 		res, err := http.Get("https://zkillboard.com/api/history/20200911.json")
	// 		if err != nil {
	// 			app.Logger.WithError(err).Error("failed to make request to zkillboard api")

	// 		}

	// 		defer res.Body.Close()
	// 		if res.StatusCode != http.StatusOK {
	// 			app.Logger.Fatal("bad status code received from zkill api")
	// 		}

	// 		data, err := ioutil.ReadAll(res.Body)
	// 		if err != nil {
	// 			app.Logger.WithError(err).Fatal("failed to read response body")
	// 		}

	// 		var remoteMap = make(map[string]string)
	// 		err = json.Unmarshal(data, &remoteMap)
	// 		if err != nil {
	// 			app.Logger.WithError(err).Fatal("failed to decode response body")
	// 		}

	// 		for _, localKM := range localKillmails {
	// 			if _, ok := remoteMap[strconv.Itoa(int(localKM.ID))]; !ok {

	// 				app.Logger.WithFields(logrus.Fields{
	// 					"id":   localKM.ID,
	// 					"hash": localKM.Hash,
	// 				}).Info("localKM does not show up in remote source")

	// 			}

	// 		}

	// 		return nil
	// 	},
	// }
}
