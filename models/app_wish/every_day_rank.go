package models

import (
	"time"

	"github.com/astaxie/beego/orm"
)

type EveryDayRank struct {
	Id        int       `orm:"column(id)"`
	Position  int       `orm:"column(position)"`
	ProductId uint32    `orm:"column(product_id)"`
	Created   time.Time `orm:"column(created);type(datetime)"`
	RankType  int       `orm:"column(rank_type)"`
}

func (t *EveryDayRank) TableName() string {
	return "every_day_rank"
}

func init() {
	orm.RegisterModel(new(EveryDayRank))
}
