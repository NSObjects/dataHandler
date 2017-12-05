package tmodels

import (
	"time"

	"github.com/astaxie/beego/orm"
)

type TIncremental struct {
	Id                       int       `orm:"column(id)"`
	NumBought                int       `orm:"column(num_bought);null"`
	RatingCount              int       `orm:"column(rating_count);null"`
	NumCollection            int       `orm:"column(num_collection);null"`
	NumBoughtIncremental     int       `orm:"column(num_bought_incremental);null"`
	RatingCountIncremental   int       `orm:"column(rating_count_incremental);null"`
	NumCollectionIncremental int       `orm:"column(num_collection_incremental);null"`
	Price                    float64   `orm:"column(price);null;digits(10);decimals(2)"`
	Created                  time.Time `orm:"column(created);type(datetime)"`
	Updated                  time.Time `orm:"column(updated);type(datetime);null"`
	ProductId                uint      `orm:"column(product_id)"`
	PriceIncremental         float64   `orm:"column(price_incremental);null;digits(11);decimals(2)"`
}

func (t *TIncremental) TableName() string {
	return "t_incremental"
}

func init() {
	orm.RegisterModel(new(TIncremental))
}
