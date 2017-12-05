package models

import (
	"time"

	"github.com/astaxie/beego/orm"
)

type EveryDayIncremental struct {
	Id                     int       `orm:"column(id)"`
	BeginningOfDay         time.Time `orm:"column(beginning_of_day);type(datetime);null"`
	EndOfDay               time.Time `orm:"column(end_of_day);type(datetime);null"`
	Created                time.Time `orm:"column(created);type(datetime)"`
	Updated                time.Time `orm:"column(updated);type(datetime)"`
	ProductId              uint32    `orm:"column(product_id)"`
	PriceIncremental       float64   `orm:"column(price_incremental)"`
	Price                  float64   `orm:"column(price)"`
	NumEnteredIncremental  int       `orm:"column(num_entered_incremental)"`
	NumEntered             int       `orm:"column(num_entered)"`
	NumBoughtIncremental   int       `orm:"column(num_bought_incremental);null"`
	NumBought              int       `orm:"column(num_bought)"`
	RatingCountIncremental int       `orm:"column(rating_count_incremental)"`
	RatingCount            int       `orm:"column(rating_count)"`
	Undulate               float64   `orm:"-"`
	ConversionRate         float32   `orm:"-"`
	Bought                 int       `orm:"-"`
}

func (t *EveryDayIncremental) TableName() string {
	return "every_day_incremental"
}

func init() {
	orm.RegisterModel(new(EveryDayIncremental))
}
