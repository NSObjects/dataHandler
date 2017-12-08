package main

import (
	"dataHandler/db"
	"flag"
	"os"
	"time"

	"dataHandler/handler"

	"github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/now"
)

func init() {
	db.Setup()
}

func main() {
	defer func() {
		if err := recover(); err != nil {

		}
	}()

	var flagvar int
	flag.IntVar(&flagvar, "task", 0, "task")
	flag.Parse()
	switch flagvar {
	case 1:
		task1()
	case 2:
		task2()
	case 3:
		task3()
	}
}

func task1() {

	day := 24 * 1
	start := time.Unix(now.BeginningOfDay().Unix()-(60*60*int64(day)), 0)
	end := time.Unix(now.BeginningOfDay().Unix()-(1+(60*60*int64(day-24))), 0)
	runTask(60, "缓存总销量", func() {
		handler.CacheTotalSalesCache()
	})

	runTask(1200, "合并Viewing", func() {
		handler.MergeViewings(start, end)
	})

	runTask(600, "缓存Viewing", func() {
		handler.CacheViewingStatistics(start)
	})
}

func task2() {

	day := 24 * 1
	start := time.Unix(now.BeginningOfDay().Unix()-(60*60*int64(day)), 0)
	runTask(60, "迁移快照", func() {
		handler.MigrateSnapshot(start)
	})

	runTask(600, "计算快照变化", func() {
		handler.VarationHandler(start)
	})
}

func task3() {

	day := 24 * 1
	start := time.Unix(now.BeginningOfDay().Unix()-(60*60*int64(day)), 0)
	end := time.Unix(now.BeginningOfDay().Unix()-(1+(60*60*int64(day-24))), 0)
	runTask(600, "合并一天增量", func() {
		handler.MergeIncremental(start, end)
	})

	runTask(60, "缓存一周销量", func() {
		handler.CacheData()
	})

	runTask(60, "缓存七天环比", func() {
		handler.CacheSevenDayRingRate()
	})

	runTask(60, "购买量排名", func() {
		handler.InsertRank(start, end, 1)
	})

	//runTask(60, "收藏量排名", func() {
	//	handler.InsertRank(start, end, 2)
	//})
	//
	//runTask(60, "计算加速度", func() {
	//	handler.CalculateRank(start)
	//})
	//
	//runTask(60, "缓存排名加速度", func() {
	//	handler.CacheRank(start)
	//})
}

func runTask(timeout int64, msg string, f func()) {
	begin := time.Now()
	f()
	over := time.Now()
	t := over.Unix() - begin.Unix()
	var log = logrus.New()
	file, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		log.Out = file
	} else {
		log.Info("Failed to log to file, using default stderr")
	}
	if t < timeout {
		log.WithFields(logrus.Fields{"异常用时": t}).Debug(msg)
	}
	log.WithFields(logrus.Fields{"用时": t}).Debug(msg)

}
