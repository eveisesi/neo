package market

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/urfave/cli"
)

type Order struct {
	OrderID    int64   `json:"order_id"`
	LocationID int     `json:"location_id"`
	SystemID   int     `json:"system_id"`
	TypeID     int     `json:"type_id"`
	Price      float64 `json:"price"`
	// Duration     int       `json:"duration"`
	// IsBuyOrder   bool      `json:"is_buy_order"`
	// MinVolume    int       `json:"min_volume"`
	// Range        string    `json:"range"`
	// VolumeRemain int       `json:"volume_remain"`
	// VolumeTotal  int       `json:"volume_total"`
	// Issued       time.Time `json:"issued"`
}

func Action(c *cli.Context) {

	res, err := http.Get("https://esi.evetech.net/latest/markets/10000002/orders/?order_type=sell&type_id=44992")
	if err != nil {
		log.Fatal(err)
	}

	var orders []*Order
	err = json.NewDecoder(res.Body).Decode(&orders)
	if err != nil {
		log.Fatal(err)
	}

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

	maxPrice := median * 2
	minPrice := median / 2
	var newOrders []*Order
	for _, v := range orders {
		if v.Price > maxPrice || v.Price < minPrice {
			continue
		}

		newOrders = append(newOrders, v)
	}

	// Calculate the average which sum(newORders.Price)/ len(orders)

	var sum float64
	for _, v := range newOrders {
		sum += v.Price
	}

	average := sum / float64(len(newOrders))

	fmt.Printf("%.4f\n", average)

}
