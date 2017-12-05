package models

import (
	"time"

	"github.com/astaxie/beego/orm"
)

type Product struct {
	Id              uint32    `orm:"column(id);pk"`
	GenerationTime  time.Time `orm:"column(generation_time);type(datetime);null"`
	RatingCount     int       `orm:"column(rating_count)"`
	Name            string    `orm:"column(name);size(600)"`
	WishId          string    `orm:"column(wish_id);size(30)"`
	Color           string    `orm:"column(color);null"`
	Size            string    `orm:"column(size);null"`
	Price           float32   `orm:"column(price)"`
	RetailPrice     float32   `orm:"column(retail_price)"`
	MerchantTags    string    `orm:"column(merchant_tags);null"`
	Tags            string    `orm:"column(tags);null"`
	Shipping        float32   `orm:"column(shipping)"`
	Description     string    `orm:"column(description)"`
	NumBought       int       `orm:"column(num_bought)"`
	MaxShippingTime int       `orm:"column(max_shipping_time)"`
	MinShippingTime int       `orm:"column(min_shipping_time)"`
	NumEntered      int       `orm:"column(num_entered)"`
	Created         time.Time `orm:"column(created);type(datetime)"`
	Updated         time.Time `orm:"column(updated);type(datetime)"`
	Gender          int       `orm:"column(gender);null"`
	MerchantId      uint      `orm:"column(merchant_id)"`
	Merchant        string    `orm:"column(merchant);size(255)"`
	TrueTagIds      string    `orm:"column(true_tag_ids);size(4000);null"`
}

func (t *Product) TableName() string {
	return "product"
}

func init() {
	orm.RegisterModel(new(Product))
}
