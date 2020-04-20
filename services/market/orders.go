package market

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/esi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/null"
)

// CalculateRawMaterialCost attempts to look up a blueprint that produces the provided type id
// If a blueprint is found, the materials for the blueprint are queried and this process repeats
// itself in a recursive loop until a query for a blueprint fails. Then market price for that
// item is returned.
func (s *service) CalculateRawMaterialCost(id uint64, minDate, maxDate time.Time) float64 {

	printNSleep("Looking Blueprint that produces %d", id)

	product, err := s.universe.BlueprintProductByProductTypeID(context.Background(), id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logger.WithField("id", id).WithFields(logrus.Fields{
			"id":      id,
			"minDate": minDate.Format("2016-01-02 15-04-05"),
			"maxDate": maxDate.Format("2016-01-02 15-04-05"),
		}).WithError(err).Error("failed to lookup blueprint for item")
		return 0.00
	}

	// Product is not nil, so lets look up the inputs for this BP
	if err == nil {
		printNSleep("Found product for %d, looking up materials", id)
		materials, err := s.universe.BlueprintMaterials(context.Background(), product.TypeID)
		if err != nil {
			s.logger.WithField("id", id).WithFields(logrus.Fields{
				"id":      id,
				"minDate": minDate.Format("2016-01-02 15-04-05"),
				"maxDate": maxDate.Format("2016-01-02 15-04-05"),
			}).WithError(err).Error("failed to lookup blueprint materials for item")
			return 0.00
		}

		if len(materials) > 0 {
			printNSleep("Found %d materials", len(materials))
			var prices []float64
			for _, material := range materials {
				printNSleep("Calculating raw material cost for %d", material.MaterialTypeID)
				price := s.CalculateRawMaterialCost(material.MaterialTypeID, minDate, maxDate)

				cost := price * float64(material.Quantity)
				printNSleep("Type: %d, cost: %.f, units: %d, total: %.f to build", material.MaterialTypeID, price, material.Quantity, cost*float64(material.Quantity))
				prices = append(prices, cost)

			}
			sum := float64(0)
			for _, price := range prices {
				sum = sum + price
			}
			printNSleep("Sum For TypeID %d: %.4f", id, sum)
			return sum
		}

	}

	printNSleep("Did not find Blueprint that produces %d", id)
	printNSleep("Looking up order for %d", id)

	order, err := s.OrderByTime(context.Background(), id, minDate, maxDate)
	if err != nil {
		s.logger.WithField("id", id).WithFields(logrus.Fields{
			"id":      id,
			"minDate": minDate.Format("2016-01-02 15-04-05"),
			"maxDate": maxDate.Format("2016-01-02 15-04-05"),
		}).WithError(err).Error("failed to lookup order")
		return 0.00
	}

	printNSleep("Found order for %d, return price of %.f\n", id, order.LowPrice)

	return order.LowPrice

}

func printNSleep(frmt string, args ...interface{}) {
	fmt.Printf(frmt, args...)
	fmt.Println()
	// time.Sleep(time.Millisecond * 250)
	return
}

var rorderbytime = "order:%s:%s:%d"

func (s *service) OrderByTime(ctx context.Context, id uint64, minDate, maxDate time.Time) (*neo.Order, error) {

	var order = new(neo.Order)
	const format = "20160102150405"
	var key = fmt.Sprintf(rorderbytime, minDate.Format(format), maxDate.Format(format), id)

	result, err := s.redis.Get(key).Bytes()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, err
	}

	if len(result) > 0 {
		err = json.Unmarshal(result, &order)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal order from redis")
		}
		return order, nil
	}

	order, err = s.MarketRepository.OrderByTime(context.Background(), id, minDate, maxDate)
	if err != nil {
		s.logger.WithField("id", id).WithFields(logrus.Fields{
			"id":      id,
			"minDate": minDate.Format("2016-01-02 15-04-05"),
			"maxDate": maxDate.Format("2016-01-02 15-04-05"),
		}).WithError(err).Debug("failed to lookup order")
		return nil, err
	}

	byteSlice, err := json.Marshal(order)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal order for cache")
	}

	_, err = s.redis.Set(key, byteSlice, time.Minute*10).Result()

	return order, errors.Wrap(err, "failed to cache order in redis")

}

func (s *service) FetchOrders() {

	s.logger.Info("starting order fetcher")

	region := uint64(10000002)

	res, err := s.esi.HeadMarketsRegionIDOrders(region)
	if err != nil {
		s.logger.WithError(err).Error("failed to make head request for market orders")
		return
	}

	var strPages string
	var pages = uint(1)
	if strPages = res.Headers["X-Pages"]; strPages != "" {
		xpages, err := strconv.ParseUint(strPages, 10, 64)
		if err != nil {
			pages = 1
			s.logger.WithError(err).Error("unable to convert strPages to int. Defaulting to 1 page")
		}
		pages = uint(xpages)
	}

	s.logger.WithField("pages", pages).Info("successfully calculated number of pages")
	s.logger.Info("starting request loop to fetch orders")

	var orders = make([]*esi.Order, 0)
	for i := uint(1); i <= pages; i++ {

		res, err := s.esi.GetMarketsRegionIDOrders(region, null.NewUint(i, true))
		if err != nil {
			s.logger.WithError(err).WithField("page", i).Error("failed to make request for page of orders")
			continue
		}

		if res.Code != 200 {
			s.logger.WithFields(logrus.Fields{
				"page": i,
			}).Error("unexpected response code recieved from ESI for page of market data")
			continue
		}
		orders = append(orders, res.Data.([]*esi.Order)...)
	}

	s.logger.WithField("numorders", len(orders)).Info("done with loop fetching orders")

	mappedOrders := make(map[uint][]*esi.Order)

	for _, order := range orders {
		mappedOrders[order.TypeID] = append(mappedOrders[order.TypeID], order)
	}

	var averages = make([]*neo.Order, 0)

	s.logger.Info("starting loop to calculate averages")

	for typeID, orders := range mappedOrders {

		// Sort the order by Price from Highest to lowest
		sort.Slice(orders, func(i, j int) bool {
			return orders[i].Price > orders[j].Price
		})

		// Calculate the median index
		ordersLen := len(orders)
		mediani := ordersLen / 2

		// Get the median
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
		var newOrders []*esi.Order
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

	spew.Dump(averages)

	s.logger.Info("done with loop calculating averages")

	s.logger.Info("attempting to insert averages")

	txn, err := s.txn.Begin()
	if err != nil {
		s.logger.WithError(err).Error("unable to start transaction with repository ")
	}

	// // Now that we have averages out the orders for this round, lets store them
	// // in the database
	// _, err = s.CreateOrdersBulk(context.Background(), txn, averages)
	// if err != nil {
	// 	s.logger.WithError(err).Error("failed to insert averages into db")
	// }

	err = txn.Commit()
	if err != nil {
		s.logger.WithError(err).Error("failed to commit transaction")
	}

	s.logger.Info("done")

	return

}
