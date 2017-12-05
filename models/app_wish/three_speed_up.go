package models

import (
	"time"

	"github.com/astaxie/beego/orm"
)

type ThreeSpeedUp struct {
	Id          int64     `orm:"column(id)"`
	ProductId   uint32    `orm:"column(product_id)"`
	SpeedUp     int       `orm:"column(speed_up)"`
	Created     time.Time `orm:"column(created);type(datetime)"`
	SpeedUpType int       `orm:"column(speed_up_type)"`
	Start       time.Time `orm:"column(start);type(datetime);null"`
	End         time.Time `orm:"column(end);type(datetime);null"`
}

func (t *ThreeSpeedUp) TableName() string {
	return "three_speed_up"
}

func init() {
	orm.RegisterModel(new(ThreeSpeedUp))
}
