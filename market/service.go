package market

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/eveisesi/neo"
	core "github.com/eveisesi/neo/app"
	"github.com/urfave/cli"
)

type (
	App struct {
		*core.App
	}
	ESIOrder struct {
		OrderID    uint64  `json:"order_id"`
		LocationID uint    `json:"location_id"`
		SystemID   uint    `json:"system_id"`
		TypeID     uint    `json:"type_id"`
		Price      float64 `json:"price"`
	}
)

func Action(c *cli.Context) {

	app := App{
		core.New(),
	}

	// x, err := app.Market.Orders(context.Background(), 17194)
	// if err != nil {
	// 	app.Logger.Fatal("something is fucked")
	// }

	// spew.Dump(x)
	// return

	app.Logger.Info("make head request to orders endpoint")

	// Make a Head Reques to determine the number of pages to fetch
	req, err := http.NewRequest(http.MethodHead, "https://esi.evetech.net/latest/markets/10000029/orders/?order_type=sell", nil)
	if err != nil {
		app.Logger.WithError(err).Error("unable to build request")
		return
	}

	res, err := app.Client.Do(req)
	if err != nil {
		app.Logger.WithError(err).Error("failed to make request")
		return
	}

	var pages string
	if pages = res.Header.Get("X-Pages"); pages == "" {
		app.Logger.WithField("headers", res.Header).Error("expected to find header x-pages, none found")
		return
	}

	numpages, err := strconv.Atoi(pages)
	if err != nil {
		app.Logger.WithError(err).Error("unable to convert x-pages header to int")
		return
	}

	app.Logger.WithField("pages", pages).Info("successfully calculated numpages")
	app.Logger.Info("starting request loop to fetch orders")

	// Loop over those pages and collect the orders
	var orders = make([]*ESIOrder, 0)
	for i := 1; i <= numpages; i++ {

		q := url.Values{}
		q.Set("order_type", "sell")
		q.Set("page", strconv.Itoa(i))

		uri := url.URL{
			Scheme:   "https",
			Host:     "esi.evetech.net",
			Path:     "/v1/markets/10000002/orders/",
			RawQuery: q.Encode(),
		}

		req, err := http.NewRequest(http.MethodGet, uri.String(), nil)
		if err != nil {
			app.Logger.WithError(err).WithField("url", uri.String()).Error("failed to build request")
			continue
		}

		res, err := app.Client.Do(req)
		if err != nil {
			app.Logger.WithError(err).Error("failed to make request")
			continue
		}

		if res.StatusCode != 200 {
			data, _ := ioutil.ReadAll(res.Body)
			app.Logger.WithField("body", string(data)).Error("unexpected status code received from ESI")
			continue
		}

		var innerOrders = make([]*ESIOrder, 0)
		err = json.NewDecoder(res.Body).Decode(&innerOrders)
		if err != nil {
			app.Logger.WithError(err).Error("unable to unmarshal response into struct slice")
			continue
		}

		orders = append(orders, innerOrders...)
	}

	app.Logger.WithField("numorders", len(orders)).Info("done with loop fetching orders")

	var mappedOrders = make(map[uint][]*ESIOrder)
	// Loops over the orders and moving each order into the appropriate index on the map
	for _, order := range orders {
		mappedOrders[order.TypeID] = append(mappedOrders[order.TypeID], order)
	}

	// Lets create another slice to stores the averages in
	var averages = make([]*neo.Order, 0)

	app.Logger.Info("starting loop to calculate averages")

	// Loop over the mappedOrders, run our calculation and store in the database. The killmail ingresstor will use these values to total up killmails
	for typeID, orders := range mappedOrders {

		// Sort the order by Price from Highest to lowest
		sort.Slice(orders, func(i, j int) bool {
			return orders[i].Price > orders[j].Price
		})

		// Now get the median of the prices
		ordersLen := len(orders)
		mediani := len(orders) / 2

		var median float64
		if ordersLen%2 != 0 {
			median = orders[mediani].Price
		} else {
			median = (orders[mediani-1].Price + orders[mediani].Price) / 2
		}
		// Trying to weed out scammers
		maxAcceptablePrice := median * 2

		// Trying to weed out RMTers
		minAcceptablePrice := median / 2

		// Use the max and min, get rid of the orders that are outside of that range
		var newOrders []*ESIOrder
		for _, v := range orders {
			if v.Price > maxAcceptablePrice || v.Price < minAcceptablePrice {
				continue
			}

			newOrders = append(newOrders, v)
		}
		// Now that we have this new slice, lets ensure that the newOrder slice is sorted appropriately as well
		sort.Slice(newOrders, func(i, j int) bool {
			return newOrders[i].Price > newOrders[j].Price
		})

		// Need to calculate average, so sum up the remaining prices
		var sum float64
		for _, v := range newOrders {
			sum += v.Price
		}

		// Take the average
		average := sum / float64(len(newOrders))
		lowest := newOrders[len(newOrders)-1]
		highest := newOrders[0]

		// Assemble the DB Model
		order := &neo.Order{
			TypeID:    typeID,
			Date:      time.Now(),
			LowPrice:  lowest.Price,
			HighPrice: highest.Price,
			AvgPrice:  average,
			Tenfold:   true,
		}

		averages = append(averages, order)

	}

	app.Logger.Info("done with loop calculating averages")

	app.Logger.Info("attempting to insert averages")

	// Now that we have averages out the orders for this round, lets store them
	// in the database
	_, err = app.Market.CreateOrdersBulk(context.Background(), averages)
	if err != nil {
		app.Logger.WithError(err).Error("failed to insert averages into db")
	}

	app.Logger.Info("done")

}
