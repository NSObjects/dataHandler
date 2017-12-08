package handler

import (
	"strconv"
	"sync"
	"time"

	"dataHandler/db"
	"dataHandler/models/app_wish"
	"dataHandler/models/new_wish"
	"dataHandler/util"

	"github.com/Sirupsen/logrus"
	"github.com/astaxie/beego/orm"
	"github.com/jinzhu/now"
)

const (
	Day           = 60 * 60 * 24
	NumConcurrent = 15
)

func MergeIncremental(start, end time.Time) {

	_, err := db.AppWish.Raw("truncate week_incremental").Exec()
	if err != nil {
		log.Fatal(err)
		return
	}

	var category []models.Category
	if _, err := db.AppWish.
		QueryTable("category").
		All(&category); err == nil {
		for _, c := range category {
			db.RedisClient.Del(c.FilterId)
		}
	}

	var list orm.ParamsList

	s := time.Unix(start.Unix()-24*14*60*60, 0)

	_, err = db.DataSource.Raw("select DISTINCT product_id from t_incremental where created>=? and created<=?", s, end).
		ValuesFlat(&list)

	if err != nil {
		log.WithFields(logrus.Fields{
			"incremental.go": 50,
		}).Error(err)
		return
	}

	var wg sync.WaitGroup
	p := util.New(20)

	for _, productId := range list {
		id, ok := productId.(string)
		if ok == false {
			continue
		}

		pid, err := strconv.ParseInt(id, 0, 0)

		if err != nil {
			continue
		}
		wg.Add(1)

		go func(productId uint32, start, end time.Time) {
			p.Run(func() {
				merge(productId, start, end)
				SevenDaySalesRingRate(productId)
			})

			wg.Done()
		}(uint32(pid), start, end)
	}

	wg.Wait()
	p.Shutdown()

}

func SevenDaySalesRingRate(productId uint32) {

	t1 := time.Unix(now.BeginningOfDay().Unix()-(7*Day), 0)
	t2 := time.Unix(now.BeginningOfDay().Unix()-(14*Day), 0)

	var incrementals1 []models.EveryDayIncremental

	_, err := db.AppWish.QueryTable("every_day_incremental").
		Filter("product_id", productId).
		Filter("created__gt", t1).
		All(&incrementals1)
	if len(incrementals1) != 7 || err != nil {
		return
	}

	var incrementals2 []models.EveryDayIncremental
	_, err = db.AppWish.QueryTable("every_day_incremental").
		Filter("product_id", productId).
		Filter("created__gt", t2).
		Filter("created__lt", t1).
		All(&incrementals2)
	if len(incrementals2) != 7 || err != nil {
		return
	}

	var rate1 int
	for _, v := range incrementals1 {
		rate1 += v.NumBoughtIncremental
	}
	if rate1 <= 100 {
		return
	}

	var rate2 int
	for _, v := range incrementals2 {
		rate2 += v.NumBoughtIncremental
	}

	if rate2 <= 100 {
		return
	}

	rate := float32(rate1-rate2) / float32(rate2) * 100
	sort := models.Sort{
		SevenDaySalesRate: rate,
		ProductId:         productId,
	}

	_, err = db.AppWish.Insert(&sort)
	if err != nil {
		log.WithFields(logrus.Fields{
			"incremental.go": 338,
		}).Error(err)
	}

}

func InsertRank(start, end time.Time, rankType int) {

	page := 0
	size := 10000
	for {
		var productIncremental []models.EveryDayIncremental
		var err error
		if rankType == 1 {
			_, err = db.AppWish.QueryTable("every_day_incremental").
				Filter("created", time.Unix(start.Unix()+60, 0)).
				Filter("num_bought_incremental__gt", 0).
				OrderBy("-num_bought_incremental", "id").
				Limit(size, size*page).
				All(&productIncremental)

		} else {
			_, err = db.AppWish.QueryTable("every_day_incremental").
				Filter("created", time.Unix(start.Unix()+60, 0)).
				Filter("num_entered_incremental__gt", 0).
				OrderBy("-num_entered_incremental", "id").
				Limit(size, size*page).
				All(&productIncremental)
		}

		if err != nil || len(productIncremental) <= 0 {
			break
		}

		var wg sync.WaitGroup
		p := util.New(NumConcurrent)
		for i, incremental := range productIncremental {
			wg.Add(1)

			go func(productId uint32,
				position int,
				start time.Time,
				end time.Time,
				rankType int) {

				p.Run(func() {
					rankWork(productId, position, start, end, rankType)
				})
				wg.Done()
			}(incremental.ProductId, i+(size*page)+1, start, end, rankType)
		}
		wg.Wait()
		p.Shutdown()
		page++
	}

}

func CalculateRank(date time.Time) {

	page := 0
	size := 10000
	for {
		var wishProduct []models.WeekIncremental
		_, err := db.AppWish.QueryTable("week_incremental").
			Filter("sales_incremental__gt", 0).
			Limit(size, page*size).All(&wishProduct)

		if err != nil || len(wishProduct) <= 0 {
			break
		}

		var wg sync.WaitGroup
		p := util.New(NumConcurrent)

		for _, product := range wishProduct {
			wg.Add(1)
			go func(productId uint32) {
				p.Run(func() {
					calculateWork(productId, date)
				})
				wg.Done()
			}(product.ProductId)
		}
		wg.Wait()
		p.Shutdown()
		page++
	}

}

func rankWork(productId uint32,
	position int,
	start time.Time,
	end time.Time, rankType int) {

	qs := db.AppWish.QueryTable("every_day_rank")

	var collectionRank models.EveryDayRank
	err := qs.Filter("product_id", productId).
		Filter("created__gte", start).
		Filter("created__lte", end).
		Filter("rank_type", rankType).
		One(&collectionRank)

	collectionRank.ProductId = productId
	collectionRank.Position = position
	collectionRank.RankType = rankType
	collectionRank.Created = time.Unix(start.Unix()+60, 0)

	if err != nil {
		if err == orm.ErrNoRows {
			if _, err := db.AppWish.Insert(&collectionRank); err != nil {
				log.WithFields(logrus.Fields{
					"incremental.go": 419,
				}).Error(err)
			}
		} else {
			log.WithFields(logrus.Fields{
				"incremental.go": 424,
			}).Error(err)
		}
	} else {
		if _, err := db.AppWish.Update(&collectionRank, "position"); err != nil {
			log.WithFields(logrus.Fields{
				"incremental.go": 430,
			}).Error(err)
		}
	}
}

func calculateWork(productId uint32, date time.Time) {
	calculateRankWith(productId, 1, date)
	calculateRankWith(productId, 2, date)
}

func calculateRankWith(productId uint32, rankType int, date time.Time) {

	fourDayago := time.Unix(date.Unix()-int64(Day*4), 0)

	var salesRankings []models.EveryDayRank

	_, err := db.AppWish.QueryTable("every_day_rank").
		Filter("product_id", productId).
		Filter("created__gte", fourDayago).
		Filter("created__lte", date).
		Filter("rank_type", rankType).
		OrderBy("created").
		All(&salesRankings)

	if err != nil || len(salesRankings) < 3 {
		return
	}

	v1 := salesRankings[0].Position - salesRankings[1].Position
	v2 := salesRankings[2].Position - salesRankings[1].Position

	rank := models.ThreeSpeedUp{}
	rank.SpeedUp = v1 - v2
	rank.SpeedUpType = rankType
	rank.Created = date
	rank.ProductId = productId
	_, err = db.AppWish.Insert(&rank)
	if err != nil {
		log.WithFields(logrus.Fields{
			"incremental.go": 302,
		}).Error(err)
	}
}

func merge(productId uint32, start time.Time, end time.Time) {
	mergeIncrementalWith(productId, start, end)
	mergeSalesWith(productId, start)
}

func mergeIncrementalWith(productId uint32, start time.Time, end time.Time) {

	qs := db.DataSource.QueryTable("t_incremental")

	var datas []tmodels.TIncremental
	count, _ := qs.Filter("product_id", productId).
		Filter("created__gte", start).
		Filter("created__lte", end).Count()

	if _, err := qs.Filter("product_id", productId).
		Filter("created__gte", start).
		Filter("created__lte", end).
		OrderBy("created").
		Limit(count).
		All(&datas); err != nil {
		log.WithFields(logrus.Fields{
			"incremental.go": 352,
		}).Error(err)
		return
	}

	if len(datas) <= 0 {
		return
	}

	var numBoughtIncremental int
	var ratingCountIncremental int
	var priceIncremental float64

	if len(datas) > 1 {
		numBoughtIncremental = datas[len(datas)-1].NumBought - datas[0].NumBought
		ratingCountIncremental = datas[len(datas)-1].RatingCount - datas[0].RatingCount
		//numEnteredIncremental = datas[len(datas)-1].NumCollection - datas[0].NumCollection
		priceIncremental = datas[len(datas)-1].Price - datas[0].Price
	}

	yestday := time.Unix((start.Unix()-24*60*60)+60, 0)
	var yestdayincrmental models.EveryDayIncremental

	if err := db.AppWish.QueryTable("every_day_incremental").
		Filter("product_id", productId).
		Filter("created", yestday).
		One(&yestdayincrmental); err == nil {
		if yestdayincrmental.NumBought > 0 {
			numBoughtIncremental += datas[0].NumBought - yestdayincrmental.NumBought
			ratingCountIncremental += datas[0].RatingCount - yestdayincrmental.RatingCount
			//numEnteredIncremental += datas[0].NumCollection - yestdayincrmental.NumEntered
			priceIncremental += datas[0].Price - yestdayincrmental.Price
		}
	}

	if numBoughtIncremental <= 0 {
		return
	}

	var productIncremental models.EveryDayIncremental
	err := db.AppWish.QueryTable("every_day_incremental").
		Filter("product_id", productId).
		Filter("created", time.Unix(start.Unix()+60, 0)).
		One(&productIncremental)

	if err != nil {
		if err == orm.ErrNoRows {
			productIncremental.ProductId = productId
			productIncremental.Created = time.Unix(start.Unix()+60, 0)
			productIncremental.Updated = time.Now()
			_, err = db.AppWish.Insert(&productIncremental)
			if err != nil {
				log.WithFields(logrus.Fields{
					"incremental.go": 328,
				}).Error(err)
			}
		} else {
			log.WithFields(logrus.Fields{
				"incremental.go": 333,
			}).Error(err)
		}
	}

	if numBoughtIncremental > 0 ||
		priceIncremental > 0 {
		productIncremental.Updated = time.Now()
		productIncremental.NumBoughtIncremental = numBoughtIncremental
		productIncremental.RatingCountIncremental = ratingCountIncremental
		//productIncremental.NumEnteredIncremental = numEnteredIncremental
		productIncremental.PriceIncremental = priceIncremental

		productIncremental.BeginningOfDay = datas[0].Updated
		productIncremental.EndOfDay = datas[len(datas)-1].Updated
		productIncremental.NumBought = datas[len(datas)-1].NumBought
		productIncremental.NumEntered = datas[len(datas)-1].NumCollection
		productIncremental.RatingCount = datas[len(datas)-1].RatingCount
		productIncremental.Price = datas[len(datas)-1].Price
	}

	productIncremental.ProductId = productId
	productIncremental.Updated = time.Now()
	_, err = db.AppWish.Update(&productIncremental, "price_incremental", "price", "num_entered_incremental",
		"num_entered", "updated", "num_bought_incremental", "num_bought", "rating_count_incremental", "rating_count")
	if err != nil {
		log.WithFields(logrus.Fields{
			"incremental.go": 192,
		}).Error(err)
	}

	//product := models.Product{Id: productId}
	//if err := db.AppWish.Read(&product); err == nil {
	//	if len(product.TrueTagIds) == 0 {
	//		var pc []models.ProductCategory
	//		if _, err := db.AppWish.QueryTable("product_category").
	//			Filter("product_id", productId).
	//			All(&pc); err == nil {
	//			for _, c := range pc {
	//				category := models.Category{Id: c.CategoryId}
	//				if err := db.AppWish.Read(&category); err == nil {
	//					db.RedisClient.RPush(category.FilterId, c.ProductId)
	//				} else {
	//					log.WithFields(logrus.Fields{
	//						"incremental.go": 209,
	//					}).Error(err)
	//				}
	//			}
	//		} else {
	//			log.WithFields(logrus.Fields{
	//				"incremental.go": 215,
	//			}).Error(err)
	//		}
	//	} else {
	//		tags := strings.Split(product.TrueTagIds, ",")
	//		for _, t := range tags {
	//			db.RedisClient.RPush(t, productId)
	//		}
	//	}
	//}

}

func mergeSalesWith(productId uint32, t time.Time) {

	var productIncrementals []models.EveryDayIncremental

	qp := db.AppWish.QueryTable("every_day_incremental")
	_, err := qp.Filter("product_id", productId).
		Filter("created__gte", time.Unix(t.Unix()-24*6*60*60, 0)).
		All(&productIncrementals)

	if err != nil {
		log.WithFields(logrus.Fields{
			"incremental.go": 458,
		}).Error(err)
	}

	if len(productIncrementals) <= 0 {
		return
	}

	var weeknumBoughtIncremental int
	var weeknumEnteredIncremental int
	for _, ps := range productIncrementals {
		weeknumBoughtIncremental += ps.NumBoughtIncremental
		weeknumEnteredIncremental += ps.NumEnteredIncremental
	}

	var weekIncremental models.WeekIncremental

	if err := db.AppWish.QueryTable("week_incremental").
		Filter("product_id", productId).
		One(&weekIncremental); err != nil {
		if err == orm.ErrNoRows {
			weekIncremental.ProductId = productId
			weekIncremental.SalesIncremental = weeknumBoughtIncremental
			weekIncremental.CollectionIncremental = weeknumEnteredIncremental
			if _, err = db.AppWish.Insert(&weekIncremental); err != nil {
				log.WithFields(logrus.Fields{
					"incremental.go": 484,
				}).Error(err)
			}
		} else {
			log.WithFields(logrus.Fields{
				"incremental.go": 489,
			}).Error(err)
		}
	} else {
		if weeknumBoughtIncremental > 0 || weeknumEnteredIncremental > 0 {
			weekIncremental.SalesIncremental = weeknumBoughtIncremental
			weekIncremental.CollectionIncremental = weeknumEnteredIncremental
			if _, err := db.AppWish.Update(&weekIncremental, "sales_incremental", "collection_incremental"); err != nil {
				log.WithFields(logrus.Fields{
					"incremental.go": 498,
				}).Error(err)
			}
		}
	}

}
