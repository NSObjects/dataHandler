package models

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
)

type ViewingStatistics struct {
	Id                 int       `orm:"column(id);auto"`
	Median             int       `orm:"column(median)"`
	MeanValue          int       `orm:"column(mean_value)"`
	StandardDeviation  float64   `orm:"column(standard_deviation)"`
	Range              int       `orm:"column(range)"`
	MaximumValue       int       `orm:"column(maximum_value)"`
	MinmunValue        int       `orm:"column(minmun_value)"`
	StartValue         int       `orm:"column(start_value)"`
	EndValue           int       `orm:"column(end_value)"`
	StartEndDifference int       `orm:"column(start_end_difference)"`
	Created            time.Time `orm:"column(created);type(datetime)"`
	Updated            time.Time `orm:"column(updated);type(datetime)"`
	Date               time.Time `orm:"column(date);type(datetime)"`
	ProductId          uint32    `orm:"column(product_id)"`
	Counts             string    `orm:"column(counts)"`
	Undulate           float64   `orm:"-"`
	ConversionRate     float32   `orm:"-"`
	Bought             int       `orm:"-"`
}

func (t *ViewingStatistics) TableName() string {
	return "viewing_statistics"
}

func init() {
	orm.RegisterModel(new(ViewingStatistics))
}

// AddViewingStatistics insert a new ViewingStatistics into database and returns
// last inserted Id on success.
func AddViewingStatistics(m *ViewingStatistics) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	return
}

// GetViewingStatisticsById retrieves ViewingStatistics by Id. Returns error if
// Id doesn't exist
func GetViewingStatisticsById(id int) (v *ViewingStatistics, err error) {
	o := orm.NewOrm()
	v = &ViewingStatistics{Id: id}
	if err = o.Read(v); err == nil {
		return v, nil
	}
	return nil, err
}

// GetAllViewingStatistics retrieves all ViewingStatistics matches certain condition. Returns empty list if
// no records exist
func GetAllViewingStatistics(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(ViewingStatistics))
	// query k=v
	for k, v := range query {
		// rewrite dot-notation to Object__Attribute
		k = strings.Replace(k, ".", "__", -1)
		if strings.Contains(k, "isnull") {
			qs = qs.Filter(k, (v == "true" || v == "1"))
		} else {
			qs = qs.Filter(k, v)
		}
	}
	// order by:
	var sortFields []string
	if len(sortby) != 0 {
		if len(sortby) == len(order) {
			// 1) for each sort field, there is an associated order
			for i, v := range sortby {
				orderby := ""
				if order[i] == "desc" {
					orderby = "-" + v
				} else if order[i] == "asc" {
					orderby = v
				} else {
					return nil, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				sortFields = append(sortFields, orderby)
			}
			qs = qs.OrderBy(sortFields...)
		} else if len(sortby) != len(order) && len(order) == 1 {
			// 2) there is exactly one order, all the sorted fields will be sorted by this order
			for _, v := range sortby {
				orderby := ""
				if order[0] == "desc" {
					orderby = "-" + v
				} else if order[0] == "asc" {
					orderby = v
				} else {
					return nil, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				sortFields = append(sortFields, orderby)
			}
		} else if len(sortby) != len(order) && len(order) != 1 {
			return nil, errors.New("Error: 'sortby', 'order' sizes mismatch or 'order' size is not 1")
		}
	} else {
		if len(order) != 0 {
			return nil, errors.New("Error: unused 'order' fields")
		}
	}

	var l []ViewingStatistics
	qs = qs.OrderBy(sortFields...)
	if _, err = qs.Limit(limit, offset).All(&l, fields...); err == nil {
		if len(fields) == 0 {
			for _, v := range l {
				ml = append(ml, v)
			}
		} else {
			// trim unused fields
			for _, v := range l {
				m := make(map[string]interface{})
				val := reflect.ValueOf(v)
				for _, fname := range fields {
					m[fname] = val.FieldByName(fname).Interface()
				}
				ml = append(ml, m)
			}
		}
		return ml, nil
	}
	return nil, err
}

// UpdateViewingStatistics updates ViewingStatistics by Id and returns error if
// the record to be updated doesn't exist
func UpdateViewingStatisticsById(m *ViewingStatistics) (err error) {
	o := orm.NewOrm()
	v := ViewingStatistics{Id: m.Id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Update(m); err == nil {
			fmt.Println("Number of records updated in database:", num)
		}
	}
	return
}

// DeleteViewingStatistics deletes ViewingStatistics by Id and returns error if
// the record to be deleted doesn't exist
func DeleteViewingStatistics(id int) (err error) {
	o := orm.NewOrm()
	v := ViewingStatistics{Id: id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Delete(&ViewingStatistics{Id: id}); err == nil {
			fmt.Println("Number of records deleted in database:", num)
		}
	}
	return
}
