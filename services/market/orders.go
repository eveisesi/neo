package market

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strconv"
	"time"

	"github.com/volatiletech/null"

	"github.com/korovkin/limiter"

	"github.com/eveisesi/neo"
)

var region = uint64(10000002)

// CalculateRawMaterialCost attempts to look up a blueprint that produces the provided type id
// If a blueprint is found, the materials for the blueprint are queried and this process repeats
// itself in a recursive loop until a query for a blueprint fails. Then market price for that
// item is returned.
func (s *service) CalculateRawMaterialCost(id uint64, days int, maxDate time.Time) float64 {

	// 	product, err := s.universe.BlueprintProductByProductTypeID(context.Background(), id)
	// 	if err != nil && !errors.Is(err, sql.ErrNoRows) {
	// 		s.logger.WithField("id", id).WithFields(logrus.Fields{
	// 			"id":      id,
	// 			"days":    days,
	// 			"maxDate": maxDate.Format("2016-01-02"),
	// 		}).WithError(err).Error("failed to lookup blueprint for item")
	// 		return 0.00
	// 	}

	// 	// Product is not nil, so lets look up the inputs for this BP
	// 	if err == nil {
	// 		materials, err := s.universe.BlueprintMaterials(context.Background(), product.TypeID)
	// 		if err != nil {
	// 			s.logger.WithField("id", id).WithFields(logrus.Fields{
	// 				"id":      id,
	// 				"days":    days,
	// 				"maxDate": maxDate.Format("2016-01-02"),
	// 			}).WithError(err).Error("failed to lookup blueprint materials for item")
	// 			return 0.00
	// 		}

	// 		if len(materials) > 0 {
	// 			var prices []float64
	// 			for _, material := range materials {
	// 				price := s.CalculateRawMaterialCost(material.MaterialTypeID, days, maxDate)

	// 				cost := price * float64(material.Quantity)
	// 				prices = append(prices, cost)

	// 			}
	// 			sum := float64(0)
	// 			for _, price := range prices {
	// 				sum = sum + price
	// 			}
	// 			return sum
	// 		}

	// 	}

	// 	price, err := s.AvgOfTypeLowPrice(context.Background(), id, days, maxDate)
	// 	if err != nil {
	// 		s.logger.WithField("id", id).WithFields(logrus.Fields{
	// 			"id":      id,
	// 			"days":    days,
	// 			"maxDate": maxDate.Format("2016-01-02"),
	// 		}).WithError(err).Error("failed to lookup order")
	// 	}

	return 0.00

}

func (s *service) FetchTypePrice(id uint64, date time.Time) float64 {

	var price float64

	invType, err := s.universe.Type(context.Background(), id)
	if err != nil {
		return price
	}

	invGroup, err := s.universe.TypeGroup(context.Background(), invType.GroupID)
	if err != nil {
		return price
	}

	// Skins are worthless. Return 0.01
	if invGroup.CategoryID == 91 {
		price = 0.01
		return price
	}

	// For some reason, we build all rigs
	if invGroup.CategoryID == 66 {
		price = s.getBuildPrice(id, date)
		if price > 0.01 {
			return price
		}
	}

	price = s.getFixedPrice(id, date)
	if price > 0.00 {
		return price
	}
	history, err := s.MarketRepository.HistoricalRecord(context.Background(), id, date, null.NewInt(33, true))
	if err != nil {
		return 0.00
	}
	// fmt.Printf("\n\t\t\t---Start RawHistory---\n")
	// for _, v := range history {
	// 	fmt.Printf("%d || %s || %.4f\n", v.TypeID, v.Date.Format("2006-01-02"), v.Price)
	// }
	// fmt.Printf("\n\t\t\t---End RawHistory---\n\n")

	neededData := 33
	priceList := make([]*neo.HistoricalRecord, 0)
	if len(history) >= neededData {
		priceList = history
	} else if len(history) > 0 {
		priceList = history[0 : len(history)-1]
	} else {
		priceList = append(priceList, &neo.HistoricalRecord{Price: 0.01})
	}

	if len(priceList) >= 2 {
		sort.Slice(priceList, func(i, j int) bool {
			return priceList[i].Price > priceList[j].Price
		})
	}

	// fmt.Printf("\n\t\t\t---Start RawHistory---\n")
	// for _, v := range history {
	// 	fmt.Printf("%d,%s,%.4f\n", v.TypeID, v.Date.Format("2006-01-02"), v.Price)
	// }
	// fmt.Printf("\n\t\t\t---End RawHistory---\n\n")

	if len(priceList) == neededData {
		priceList = priceList[2:]
		priceList = priceList[:len(priceList)-1]
	} else if len(priceList) > 6 {
		priceList = priceList[:len(priceList)-2]
	}

	total := float64(0)
	for _, v := range priceList {
		// fmt.Printf("%d,%s,%.4f\n", v.TypeID, v.Date.Format("2006-01-02"), v.Price)
		total += v.Price
	}

	// fmt.Printf("Calculated Total: %.4f || Total Number of Records: %d\n", total, len(priceList))
	avgPrice := total / float64(len(priceList))
	// fmt.Printf("Calculated Average: %.4f\n", avgPrice)

	if avgPrice <= 0.01 {
		avgPrice = s.getBuildPrice(id, date)
	}

	dateRecord := getPriceFromHistorySlice(history, date)
	if dateRecord != nil && dateRecord.Price > avgPrice {
		avgPrice = dateRecord.Price
	}

	return avgPrice
}

func getPriceFromHistorySlice(history []*neo.HistoricalRecord, day time.Time) *neo.HistoricalRecord {

	for _, v := range history {
		if day.Format("2006-01-02") == v.Date.Format("2006-01-02") {
			return v
		}
	}

	return nil
}

func (s *service) getBuildPrice(id uint64, date time.Time) float64 {

	built, err := s.MarketRepository.BuiltPrice(context.Background(), id, date)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logger.WithError(err).WithField("type_id", id).WithField("date", date).Error("unexpected error encountered looking up prices build")
	}

	if err == nil {
		return built.Price
	}

	blueprint, err := s.universe.BlueprintProductByProductTypeID(context.Background(), id)
	if err != nil {
		s.logger.WithError(err).WithField("type_id", id).Error("unable to retrieve blueprint for type")
		return 0.00
	}

	materials, err := s.universe.BlueprintMaterials(context.Background(), blueprint.TypeID)
	if err != nil {
		s.logger.WithError(err).WithField("type_id", id).Error("unable to retrieve blueprint materials")
		return 0.00
	}

	total := float64(0)
	for _, m := range materials {
		p := s.FetchTypePrice(m.MaterialTypeID, date)
		p = p * float64(m.Quantity)
		total += p
	}

	return total
}

func (s *service) getFixedPrice(id uint64, date time.Time) float64 {
	// TODO: Build out engine or something for looking up fixed prices. Maybe a query to the DB, possibly to redis.
	// Not sure
	return 0.00
}

// func printNSleep(frmt string, args ...interface{}) {
// 	fmt.Printf(frmt, args...)
// 	fmt.Println()
// 	// time.Sleep(time.Millisecond * 2)
// 	return
// }

func (s *service) FetchHistory(from int) {
	s.logger.Info("fetching market groups")

	groups, m := s.esi.GetMarketGroups()
	if m.IsError() {
		s.logger.WithError(m.Msg).Error("failed to fetch market groups")
		return
	}

	limiter := limiter.NewConcurrencyLimiter(20)

	for _, v := range groups {
		if from > 0 && from > v {
			continue
		}
		limiter.Execute(func() {
			s.processGroup(v)
			return
		})
	}

}

func (s *service) processGroup(v int) {

	s.logger.WithField("group_id", v).Info("processing group")

	group, m := s.esi.GetMarketGroupsMarketGroupID(v)
	if m.IsError() {
		s.logger.WithError(m.Msg).WithField("market_group_id", v).Error("failed to fetch types for market group")
		return
	}

	for _, t := range group.Types {

		s.logger.WithField("type_id", t).Info("processing historical records for type")

		records, m := s.esi.GetMarketsRegionIDHistory(region, strconv.FormatUint(t, 10))
		if m.IsError() {
			s.logger.WithError(m.Msg).WithField("type_id", t).Error("failed to pull market history for type")
			continue
		}

		if len(records) == 0 {
			s.logger.WithField("type_id", t).Info("skipping type. No history exists")
			continue
		}

		chunks := chunkRecords(records, 250)
		for _, chunk := range chunks {
			for _, record := range chunk {
				record.TypeID = t
			}
			_, err := s.MarketRepository.CreateHistoricalRecord(context.Background(), chunk)
			if err != nil {
				s.logger.WithError(err).WithField("type_id", t).Error("failed to insert chunk of historical records into db")
			}
		}
		s.logger.WithField("type_id", t).Info("successfully processed historical records for type")

	}
	s.logger.WithField("group_id", v).Info("done processing group")
	time.Sleep(time.Millisecond * 100)
}

func chunkRecords(records []*neo.HistoricalRecord, limit int) [][]*neo.HistoricalRecord {
	chunks := make([][]*neo.HistoricalRecord, 0)
	for i := 0; i <= len(records)-1; i += limit {
		end := i + limit
		if end > len(records) {
			end = len(records)
		}

		chunks = append(chunks, records[i:end])
	}

	return chunks
}

func (s *service) ProcessType(workerID, id int) {
	s.logger.WithField("type_id", id).Info("requesting records for type")

	records, m := s.esi.GetMarketsRegionIDHistory(region, strconv.Itoa(id))
	if m.IsError() {
		s.logger.WithError(m).WithField("type_id", id).Error("failed to make request for historical records of type")
		return
	}

	if m.Code != 200 {
		s.logger.WithField("type_id", id).WithField("code", m.Code).Error("unexpected response code received from ESI fro page of market averages")
		return
	}

	if len(records) == 0 {
		return
	}

	for _, t := range records {
		t.TypeID = uint64(id)
	}

	_, err := s.MarketRepository.CreateHistoricalRecord(context.Background(), records)
	if err != nil {
		s.logger.WithError(err).WithField("type_id", id).Error("unable to insert historical record into db")
	}

}
