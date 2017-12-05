package handler

import (
	"dataHandler/db"
	"dataHandler/models/app_wish"
	"dataHandler/models/new_wish"
	"dataHandler/util"
	"encoding/json"

	"math"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/astaxie/beego/orm"
	"github.com/jinzhu/now"
)

func MergeViewings(start time.Time, end time.Time) {

	var list orm.ParamsList

	_, err := db.DataSource.Raw("select DISTINCT product_id from t_viewings where created>? and created<?", start, end).
		ValuesFlat(&list)

	if err != nil {
		log.WithFields(logrus.Fields{
			"merge_viwings.go": 30,
		}).Error(err)
	}

	var wg sync.WaitGroup
	p := util.New(15)
	for _, productId := range list {
		if id, ok := productId.(string); ok {
			if pid, err := strconv.ParseInt(id, 0, 0); err == nil {
				wg.Add(1)
				go func(productId uint32, date time.Time) {
					p.Run(func() {
						mergeViewnings(productId, date)
					})
					wg.Done()
				}(uint32(pid), start)

			} else {
				log.Fatal(err)
			}
		}

	}
	wg.Wait()
	p.Shutdown()
}

func Range(vs []viewings) int {
	if len(vs) < 2 {
		return 0
	}
	sort.Sort(ByViewningDate(vs))
	return MaximumValue(vs) - MinmunValue(vs)
}

func StartEndDifference(vs []viewings) int {
	if len(vs) < 2 {
		return 0
	}
	sort.Sort(ByViewningDate(vs))
	return StartValue(vs) - EndValue(vs)
}

func EndValue(vs []viewings) int {
	if len(vs) <= 0 {
		return 0
	}
	sort.Sort(ByViewningDate(vs))
	return vs[len(vs)-1].Count
}

func StartValue(vs []viewings) int {
	if len(vs) <= 0 {
		return 0
	}
	sort.Sort(ByViewningDate(vs))
	return vs[0].Count
}

func MinmunValue(vs []viewings) int {
	if len(vs) <= 0 {
		return 0
	}

	minmumValue := vs[0].Count
	for _, v := range vs {
		if v.Count < minmumValue {
			minmumValue = v.Count
		}
	}
	return minmumValue
}

func MaximumValue(vs []viewings) int {
	maxValue := 0
	for _, v := range vs {
		if v.Count > maxValue {
			maxValue = v.Count
		}
	}
	return maxValue
}

func Median(vs []viewings) int {

	if len(vs) <= 2 {
		return 0
	}
	sort.Sort(ByViewningDate(vs))
	return vs[len(vs)/2].Count

}

func MeanValue(vs []viewings) int {

	if len(vs) <= 0 {
		return 0
	}

	value := 0
	for _, v := range vs {
		value += v.Count
	}

	return value / len(vs)

}

type viewings struct {
	Count   int
	Created time.Time
}

func mergeViewnings(productId uint32, date time.Time) {

	var vs []viewings
	t := 0

	l := now.New(date)
	for {

		start := time.Unix(l.Unix()+int64((60*30)*t), 0)
		end := time.Unix(l.Unix()+int64((60*30)*(t+1)), 0)

		if t >= 48 {
			t = 0
			break
		}

		var viewing []tmodels.TViewings
		_, err := db.DataSource.QueryTable("t_viewings").
			Filter("product_id", productId).
			Filter("created__gte", start).
			Filter("created__lte", end).
			OrderBy("created").
			All(&viewing)

		if err != nil {
			log.WithFields(logrus.Fields{
				"merge_viwings.go": 83,
			}).Error(err)
			break
		}

		if len(viewing) > 0 {
			var count int
			for _, v := range viewing {
				count += v.Count
			}
			viewning := viewings{
				Count:   count / len(viewing),
				Created: end,
			}
			vs = append(vs, viewning)
		}
		t++

	}
	sort.Sort(ByViewningDate(vs))
	if len(vs) > 3 {
		var viewningsStatistics models.ViewingStatistics
		err := db.AppWish.QueryTable("viewing_statistics").
			Filter("date", l.Time).
			Filter("product_id", productId).
			One(&viewningsStatistics)
		if err != nil {
			if err == orm.ErrNoRows {
				viewningsStatistics := models.ViewingStatistics{
					MeanValue:          MeanValue(vs),
					Median:             Median(vs),
					MaximumValue:       MaximumValue(vs),
					MinmunValue:        MinmunValue(vs),
					StartValue:         StartValue(vs),
					EndValue:           EndValue(vs),
					StartEndDifference: StartEndDifference(vs),
					Range:              Range(vs),
					StandardDeviation:  standardDeiation(vs),
					Date:               l.Time,
					Counts:             string(counts(vs)),
					ProductId:          productId,
					Created:            time.Now(),
					Updated:            time.Now(),
				}
				_, err := db.AppWish.Insert(&viewningsStatistics)
				if err != nil {
					log.WithFields(logrus.Fields{
						"merge_viwings.go": 128,
					}).Error(err)
				}
			} else {
				log.WithFields(logrus.Fields{
					"merge_viwings.go": 133,
				}).Error(err)
			}
		} else {
			viewningsStatistics.MeanValue = MeanValue(vs)
			viewningsStatistics.Median = Median(vs)
			viewningsStatistics.MaximumValue = MaximumValue(vs)
			viewningsStatistics.MinmunValue = MinmunValue(vs)
			viewningsStatistics.StartValue = StartValue(vs)
			viewningsStatistics.EndValue = EndValue(vs)
			viewningsStatistics.StartEndDifference = StartEndDifference(vs)
			viewningsStatistics.Range = Range(vs)
			viewningsStatistics.StandardDeviation = standardDeiation(vs)
			viewningsStatistics.Date = l.Time
			viewningsStatistics.Counts = string(counts(vs))
			viewningsStatistics.ProductId = productId

			_, err := db.AppWish.Update(&viewningsStatistics)
			if err != nil {
				log.WithFields(logrus.Fields{
					"merge_viwings.go": 153,
				}).Error(err)
			}
		}

	}
}

func counts(vs []viewings) []byte {
	cs := make([]map[string]interface{}, 0)
	sort.Sort(ByViewningDate(vs))
	for _, v := range vs {
		k := make(map[string]interface{})
		k["date"] = v.Created.String()
		k["count"] = v.Count
		cs = append(cs, k)
	}

	j, _ := json.Marshal(&cs)

	return j
}

func standardDeiation(vs []viewings) float64 {
	value := 0

	for _, v := range vs {
		value += v.Count
	}

	average := value / len(vs)
	var hold float64 = 0
	var variance float64 = 0
	for _, v := range vs {
		hold += math.Pow(float64(v.Count-average), 2)
		variance = hold / float64(len(vs))
	}
	return math.Sqrt(variance)
}

type ByViewningCount []viewings

func (a ByViewningCount) Len() int           { return len(a) }
func (a ByViewningCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByViewningCount) Less(i, j int) bool { return a[i].Count < a[j].Count }

type ByViewningDate []viewings

func (a ByViewningDate) Len() int           { return len(a) }
func (a ByViewningDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByViewningDate) Less(i, j int) bool { return a[i].Created.Unix() < a[j].Created.Unix() }
