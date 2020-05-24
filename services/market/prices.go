package market

import (
	"context"
	"time"

	"github.com/eveisesi/neo"
)

func (s *service) FetchPrices() {

	s.logger.Info("fetching hourly market prices")

	prices, m := s.esi.GetMarketsPrices()
	if m.IsError() {
		s.logger.WithError(m.Msg).Error("failed to fetch market prices")
		return
	}

	records := make([]*neo.HistoricalRecord, 0)
	for _, price := range prices {

		record := &neo.HistoricalRecord{
			TypeID: price.TypeID,
			Date:   &neo.Date{Time: time.Now()},
		}

		// Select the greater of the two
		p := float64(0)
		if price.AdjustedPrice > price.AveragePrice {
			p = price.AdjustedPrice
		} else if price.AveragePrice > price.AdjustedPrice {
			p = price.AveragePrice
		}

		if p <= 0.01 {
			continue
		}

		record.Price = p

		records = append(records, record)

	}

	chunks := chunkRecords(records, 1000)
	for _, chunk := range chunks {
		_, err := s.MarketRepository.CreateHistoricalRecord(context.Background(), chunk)
		if err != nil {
			s.logger.WithError(err).Error("failed to insert prices chunk into db")
			return
		}

		time.Sleep(time.Millisecond * 50)
	}

}
