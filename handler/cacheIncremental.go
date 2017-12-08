package handler

import (
	"dataHandler/db"
	"dataHandler/models/app_wish"
	"dataHandler/util"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/astaxie/beego/orm"
	"github.com/jinzhu/now"
	redis "gopkg.in/redis.v5"
)

type data struct {
	Num  int   `json:"num"`
	Time int64 `json:"time"`
}

type RankData struct {
	ProductId string `json:"product_id"`
	SpeedUp   string `json:"speedup"`
}

type increlemts struct {
	SalerRanks           []data
	SevenDayOfSalerCount []data
	SevenDayOfCollection []data
}

type TempRank struct {
	Incremental int64
	TotalBought int64
	WishId      string
}

var log = logrus.New()

func init() {
	file, err := os.OpenFile("err.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		log.Out = file
	} else {
		log.Info("Failed to log to file, using default stderr")
	}
}

func CacheData() {

	db.RedisClient.Del(util.WEEK_SORT_PRODUCT_CACHE)
	var category []models.Category

	db.AppWish.QueryTable("category").All(&category)
	for _, c := range category {
		if len(c.FilterId) > 0 {
			db.RedisClient.Del(c.FilterId)
		}
	}

	var weekIncremental []models.WeekIncremental
	index := 0
	size := 10000
	var lastId = -1
	for {

		_, err := db.AppWish.QueryTable("week_incremental").
			Filter("sales_incremental__gt", 0).
			OrderBy("-sales_incremental").
			Limit(size, size*index).
			All(&weekIncremental)

		if err != nil || len(weekIncremental) <= 0 {
			break
		}

		if lastId == weekIncremental[len(weekIncremental)-1].Id {
			break
		}

		lastId = weekIncremental[len(weekIncremental)-1].Id

		var wg sync.WaitGroup
		p := util.New(15)

		for i, value := range weekIncremental {
			wg.Add(1)

			go func(productId uint32, index int) {
				p.Run(func() {
					cacheWeekSalesSort(productId, index)
				})
				wg.Done()
			}(value.ProductId, i+(index*size))
		}

		wg.Wait()
		p.Shutdown()
		index++
	}
}

func CacheSevenDayRingRate() {
	o := orm.NewOrm()

	db.RedisClient.Del(util.SEVEN_DAY_RING_RATE)

	index := 0
	size := 10000
	var lastId = -1
	for {
		var sort []models.Sort
		_, err := o.QueryTable("sort").
			OrderBy("-seven_day_sales_rate").
			Limit(size, size*index).
			All(&sort)

		if err != nil || len(sort) <= 0 {

			break
		}

		if lastId == sort[len(sort)-1].Id {
			break
		}

		lastId = sort[len(sort)-1].Id

		var wg sync.WaitGroup
		p := util.New(15)

		for i, value := range sort {
			wg.Add(1)

			go func(productId uint32, index int) {
				p.Run(func() {
					db.RedisClient.ZAdd(util.SEVEN_DAY_RING_RATE, redis.Z{Score: float64(index), Member: productId}).Result()
				})
				wg.Done()
			}(value.ProductId, i+(index*size))
		}

		wg.Wait()
		p.Shutdown()
		index++
	}
}

func CacheTotalSalesCache() {
	db.RedisClient.Del(util.TOTAL_SORT_PRODUCT_CACHE)

	var products []models.Product

	_, err := db.AppWish.QueryTable("product").
		OrderBy("-num_bought").
		Limit(100000).
		All(&products)

	if err != nil || len(products) <= 0 {
		return
	}

	var wg sync.WaitGroup
	p := util.New(15)

	for i, value := range products {
		wg.Add(1)

		go func(productId uint32, index int) {
			p.Run(func() {
				cacheTotalSalesSort(productId, index)
			})
			wg.Done()
		}(value.Id, i)
	}

	wg.Wait()
	p.Shutdown()
}

func cacheWeekSalesSort(productId uint32, index int) {
	//o := orm.NewOrm()
	_, err := db.RedisClient.ZAdd(util.WEEK_SORT_PRODUCT_CACHE, redis.Z{Score: float64(index), Member: productId}).Result()

	if err != nil {
		log.WithFields(logrus.Fields{
			"cacheIncremental.go": 191,
		}).Error(err)
		return
	}

	CacheIncremental(productId)
}

func cacheTotalSalesSort(productId uint32, index int) {

	_, err := db.RedisClient.ZAdd(util.TOTAL_SORT_PRODUCT_CACHE, redis.Z{Score: float64(index), Member: productId}).Result()
	if err != nil {
		log.WithFields(logrus.Fields{
			"cacheIncremental.go": 204,
		}).Error(err)
	}
}

func CacheIncremental(productId uint32) {

	db.RedisClient.Del(fmt.Sprintf("%d", productId))
	n := now.BeginningOfDay()
	value := increlemts{}

	var threeMonthAgo int64 = 7776000

	var productIncrementals []models.EveryDayIncremental
	_, err := db.AppWish.QueryTable("every_day_incremental").
		Filter("product_id", productId).
		Filter("created__gte", time.Unix(n.Unix()-threeMonthAgo, 0)).
		OrderBy("created").
		All(&productIncrementals)

	if err != nil {
		log.WithFields(logrus.Fields{
			"cacheIncremental.go": 226,
		}).Error(err)
		return
	}

	var sevenDayOfCollection []data
	var sevenDayOfSalerCount []data

	for _, incremental := range productIncrementals {

		sevenDayOfCollection = append(sevenDayOfCollection, data{Num: incremental.NumEnteredIncremental, Time: incremental.Created.Unix()})
		sevenDayOfSalerCount = append(sevenDayOfSalerCount, data{Num: incremental.NumBoughtIncremental, Time: incremental.Created.Unix()})
	}

	value.SevenDayOfSalerCount = sevenDayOfSalerCount
	value.SevenDayOfCollection = sevenDayOfCollection

	var erverDaysalesRanks []models.EveryDayRank

	_, err = db.AppWish.QueryTable("every_day_rank").
		Filter("product_id", productId).
		Filter("created__gte", time.Unix(n.Unix()-threeMonthAgo, 0)).
		Filter("created__lte", time.Unix(n.Unix()-1, 0)).
		OrderBy("created").
		All(&erverDaysalesRanks)

	if err != nil {
		log.WithFields(logrus.Fields{
			"cacheIncremental.go": 254,
		}).Error(err)
		return
	}

	var erverDaySalesRank []data

	for _, incremental := range erverDaysalesRanks {
		erverDaySalesRank = append(erverDaySalesRank, data{Num: incremental.Position, Time: incremental.Created.Unix()})
	}

	value.SalerRanks = erverDaySalesRank
	s, e := json.Marshal(&value)
	if e != nil {
		log.WithFields(logrus.Fields{
			"cacheIncremental.go": 269,
		}).Error(e)
	}

	if err := db.RedisClient.Set(fmt.Sprintf("%d", productId), string(s), 0).Err(); err != nil {
		log.WithFields(logrus.Fields{
			"cacheIncremental.go": 275,
		}).Error(err)
	}
}

func CacheRank(date time.Time) {
	o := orm.NewOrm()
	db.RedisClient.Del(util.COLLECTION_RANK_CACHE)
	db.RedisClient.Del(util.BOUGHT_RANK_CACHE)

	start := time.Unix(date.Unix()-int64(60*60*24*7), 0)

	var maps []orm.Params
	_, err := o.Raw("select product_id,sum_speed from (select product_id,sum(speed_up) as sum_speed from (select * from three_speed_up where created >= ? and created <= ? and speed_up_type = 1)d group by product_id)s ORDER BY sum_speed DESC", start, time.Now()).
		Values(&maps)
	if err != nil {
		log.WithFields(logrus.Fields{
			"cacheIncremental.go": 292,
		}).Error(err)
	} else {
		for _, value := range maps {

			if len(value["sum_speed"].(string)) > 0 &&
				len(value["product_id"].(string)) > 0 {
				s, e := json.Marshal(&RankData{ProductId: value["product_id"].(string), SpeedUp: value["sum_speed"].(string)})
				if e == nil {
					db.RedisClient.RPush(util.BOUGHT_RANK_CACHE, s)
				} else {
					log.WithFields(logrus.Fields{
						"cacheIncremental.go": 304,
					}).Error(e)
				}
			}

		}
	}

	_, err = o.Raw("select product_id,sum_speed from (select product_id,sum(speed_up) as sum_speed from (select * from three_speed_up where created >= ? and created <= ? and speed_up_type = 2)d group by product_id)s ORDER BY sum_speed DESC", start, time.Now()).Values(&maps)
	if err != nil {
		fmt.Println(err)
	} else {
		for _, value := range maps {
			if len(value["sum_speed"].(string)) > 0 &&
				len(value["product_id"].(string)) > 0 {
				s, e := json.Marshal(&RankData{ProductId: value["product_id"].(string), SpeedUp: value["sum_speed"].(string)})
				if e == nil {
					db.RedisClient.RPush(util.COLLECTION_RANK_CACHE, s)
				} else {
					log.WithFields(logrus.Fields{
						"cacheIncremental.go": 324,
					}).Error(err)
				}
			}

		}
	}

}

func CacheViewingStatistics(date time.Time) {

	o := orm.NewOrm()

	var categorys []models.Category
	if _, err := o.QueryTable("category").All(&categorys); err == nil {
		for _, v := range categorys {
			db.RedisClient.Del(fmt.Sprintf("%d-%s-viewing", date.YearDay(), v.Id))
		}
	}

	key := fmt.Sprintf("%d-viewing", date.YearDay())
	db.RedisClient.Del(key)
	db.RedisClient.Del(fmt.Sprintf("cr_%d_viewing", date.YearDay()))
	db.RedisClient.Del(fmt.Sprintf("bought_%d_viewing", date.YearDay()))

	var vs []models.ViewingStatistics
	var viewingStatistics []models.ViewingStatistics

	c, _ := o.QueryTable("viewing_statistics").
		Filter("date", date).
		Count()
	_, err := o.QueryTable("viewing_statistics").
		Filter("date", date).
		Limit(c).
		All(&viewingStatistics)

	if err != nil {
		log.WithFields(logrus.Fields{
			"cacheIncremental.go": 362,
		}).Error(err)
		return
	}

	for _, v := range viewingStatistics {
		vs = append(vs, v)
		var p models.EveryDayIncremental

		err := o.QueryTable("every_day_incremental").
			Filter("product_id", v.ProductId).
			Filter("created", time.Unix(now.New(date).BeginningOfDay().Unix()+60, 0)).
			One(&p)
		if err == nil {

			v.ConversionRate = float32(p.NumBoughtIncremental) / float32(v.MeanValue*60*24)
			v.Bought = p.NumBoughtIncremental
			vs = append(vs, v)
		} else {
			log.WithFields(logrus.Fields{
				"cacheIncremental.go": 382,
			}).Error(err)
		}
	}

	sort.Sort(ByConversionRate(vs))

	for _, v := range vs {
		v.Counts = ""
		if jv, err := json.Marshal(&v); err == nil {
			var pc []models.ProductCategory
			if _, err := o.QueryTable("product_category").
				Filter("product_id", v.ProductId).
				All(&pc); err == nil {
				for _, c := range pc {

					category := models.Category{Id: c.CategoryId}
					if err := o.Read(&category); err == nil {
						err := db.RedisClient.RPush(fmt.Sprintf("%d-%s-viewing", date.YearDay(), category.FilterId), v.ProductId).Err()
						if err != nil {
							log.WithFields(logrus.Fields{
								"cacheIncremental.go": 403,
							}).Error(err)
						}
					} else {
						log.WithFields(logrus.Fields{
							"cacheIncremental.go": 408,
						}).Error(err)
					}
				}
			} else {
				log.WithFields(logrus.Fields{
					"cacheIncremental.go": 414,
				}).Error(err)
			}
			err := db.RedisClient.HSet(key, fmt.Sprint(v.ProductId), jv).Err()
			if err != nil {
				log.WithFields(logrus.Fields{
					"cacheIncremental.go": 420,
				}).Error(err)
			}
			err = db.RedisClient.RPush(fmt.Sprintf("cr_%d_viewing", date.YearDay()), v.ProductId).Err()
			if err != nil {
				log.WithFields(logrus.Fields{
					"cacheIncremental.go": 425,
				}).Error(err)
			}

		} else {
			log.WithFields(logrus.Fields{
				"cacheIncremental.go": 432,
			}).Error(err)
		}

	}

	sort.Sort(ByBought(vs))
	for _, b := range vs {
		db.RedisClient.RPush(fmt.Sprintf("bought_%d_viewing", date.YearDay()), b.ProductId)
	}

}

type ByBought []models.ViewingStatistics

func (a ByBought) Len() int           { return len(a) }
func (a ByBought) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByBought) Less(i, j int) bool { return a[i].Bought > a[j].Bought }

func CacheCategory() {
	o := orm.NewOrm()
	var categorys []models.Category
	if _, err := o.QueryTable("category").All(&categorys); err != nil {
		return
	}

	for _, c := range categorys {
		var pcs []models.ProductCategory
		if _, err := o.QueryTable("product_category").
			Filter("category_id", c.Id).
			All(&pcs); err != nil {
			fmt.Println(err)
			continue
		}

		var weekSales []models.WeekIncremental

		for _, pc := range pcs {
			var w models.WeekIncremental
			err := o.QueryTable("week_incremental").
				Filter("product_id", pc.ProductId).One(&w)
			if err == nil {
				weekSales = append(weekSales, w)
			}
		}

		sort.Sort(ByWeekSales(weekSales))
		for _, w := range weekSales {
			db.RedisClient.RPush(c.FilterId, w.ProductId)
		}
	}
}

type ByConversionRate []models.ViewingStatistics

func (a ByConversionRate) Len() int           { return len(a) }
func (a ByConversionRate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByConversionRate) Less(i, j int) bool { return a[i].ConversionRate > a[j].ConversionRate }

type ByWeekSales []models.WeekIncremental

func (a ByWeekSales) Len() int           { return len(a) }
func (a ByWeekSales) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByWeekSales) Less(i, j int) bool { return a[i].SalesIncremental < a[j].SalesIncremental }

// By is the type of a "less" function that defines the ordering of its Planet arguments.
type By func(p1, p2 *TempRank) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by By) Sort(productIncrementl []TempRank) {
	ps := &productIncrementlSorter{
		productIncrementl: productIncrementl,
		by:                by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ps)
}

// planetSorter joins a By function and a slice of Planets to be sorted.
type productIncrementlSorter struct {
	productIncrementl []TempRank
	by                func(p1, p2 *TempRank) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *productIncrementlSorter) Len() int {
	return len(s.productIncrementl)
}

// Swap is part of sort.Interface.
func (s *productIncrementlSorter) Swap(i, j int) {
	s.productIncrementl[i], s.productIncrementl[j] = s.productIncrementl[j], s.productIncrementl[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *productIncrementlSorter) Less(i, j int) bool {
	return s.by(&s.productIncrementl[i], &s.productIncrementl[j])
}
