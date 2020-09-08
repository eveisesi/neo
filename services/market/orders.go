package market

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/korovkin/limiter"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/volatiletech/null"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/eveisesi/neo"
)

var region = uint(10000002)

func (s *service) FetchTypePrice(id uint, date time.Time) float64 {

	var price float64 = 0.01

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

	price = s.getCalculatedPrice(id, date)
	if price > 0.00 {
		return price
	}

	if !invType.Published {
		return price
	}

	history, err := s.MarketRepository.HistoricalRecord(context.Background(), id, date, null.NewInt(33, true))
	if err != nil {
		return 0.00
	}

	// We need 33 records at least to do this correctly
	neededData := 33
	priceList := make([]*neo.HistoricalRecord, 0)
	// We have more than enough
	if len(history) > 0 {
		priceList = history
		// Ok, do we don't have 33. Lets take what we can get
	} else {
		priceList = append(priceList, &neo.HistoricalRecord{Price: 0.01})
	}

	// Sort it if it is sortable
	if len(priceList) >= 2 {
		sort.Slice(priceList, func(i, j int) bool {
			return priceList[i].Price > priceList[j].Price
		})
	}

	// Lets try to get rid of gouging and low cuts
	if len(priceList) == neededData {
		priceList = priceList[2:]
		priceList = priceList[:len(priceList)-1]
		// Fuck that, just take what we can get
	} else if len(priceList) > 6 {
		priceList = priceList[:len(priceList)-2]
	}

	total := float64(0)
	for _, v := range priceList {
		total += v.Price
	}

	// Average it all up
	avgPrice := total / float64(len(priceList))

	// Is the average worthless?
	if avgPrice <= 0.01 {
		avgPrice = s.getBuildPrice(id, date)
	}

	// Is the average on this day in history greater than what we calculated
	dateRecord := getPriceFromHistorySlice(history, date)
	if dateRecord != nil && dateRecord.Price > avgPrice {

		// Yes, well than take that instead of our calculated average
		avgPrice = dateRecord.Price
	}

	return avgPrice
}

func getPriceFromHistorySlice(history []*neo.HistoricalRecord, day time.Time) *neo.HistoricalRecord {

	for _, v := range history {
		if day.Format("2006-01-02") == v.Date {
			return v
		}
	}

	return nil
}

func (s *service) getBuildPrice(id uint, date time.Time) float64 {

	built, err := s.MarketRepository.BuiltPrice(context.Background(), id, date)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
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

func (s *service) getFixedPrice(id uint, date time.Time) float64 {
	// TODO: Build out engine or something for looking up fixed prices. Maybe a query to the DB, possibly to redis.
	// Not sure

	switch id {
	case 670, 33328:
		return 10000.0000
	case 4318:
		return 0.01
	}

	return 0.00
}

func (s *service) getCalculatedPrice(id uint, date time.Time) float64 {

	// TODO: Same as TODO in getFixedPrice

	switch id {
	case 2233: // Planatary Customs Office
		gantry := s.FetchTypePrice(3962, date)
		nodes := s.FetchTypePrice(2867, date)
		modules := s.FetchTypePrice(2871, date)
		mainframes := s.FetchTypePrice(2876, date)
		cores := s.FetchTypePrice(2872, date)
		total := gantry + ((nodes + modules + mainframes + cores) * 8)
		return total
	}

	return 0.00
}

func (s *service) FetchHistory(ctx context.Context) {
	s.logger.Info("fetching market groups")

	groups, m := s.esi.GetMarketGroups(ctx)
	if m.IsErr() {
		s.logger.WithError(m.Msg).Error("failed to fetch market groups")
		return
	}

	limiter := limiter.NewConcurrencyLimiter(20)

	for _, v := range groups {
	LoopStart:
		proceed := s.tracker.Watchman(ctx)
		if !proceed {
			time.Sleep(time.Second)
			goto LoopStart
		}
		limiter.Execute(func() {
			s.processGroup(ctx, v)
		})
	}

	s.logger.Info("done fetching market data")

}

func (s *service) processGroup(ctx context.Context, v int) {

	txn := newrelic.FromContext(ctx).NewGoroutine()
	defer txn.End()

	ctx = newrelic.NewContext(ctx, txn)

	s.logger.WithField("group_id", v).Info("processing group")

	group, m := s.esi.GetMarketGroupsMarketGroupID(ctx, v)
	if m.IsErr() {
		s.logger.WithError(m.Msg).WithField("market_group_id", v).Error("failed to fetch types for market group")
		return
	}

	for _, t := range group.Types {

		s.logger.WithField("type_id", t).Info("processing historical records for type")

		info, err := s.universe.Type(ctx, t)
		if err != nil {
			s.logger.WithError(err).WithField("type_id", t).Error("failed to fetch item info")
			return
		}

		if !info.Published {
			continue
		}

		records, m := s.esi.GetMarketsRegionIDHistory(ctx, region, t)
		if m.IsErr() {
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
			_, err := s.MarketRepository.CreateHistoricalRecord(ctx, chunk)
			if err != nil {
				s.logger.WithError(err).WithField("type_id", t).Error("failed to insert chunk of historical records into db")
			}
		}
		s.logger.WithField("type_id", t).Debug("successfully processed historical records for type")

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
