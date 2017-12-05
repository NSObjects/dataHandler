package models

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
)

type PbKeywordRecords struct {
	Id                  int       `orm:"column(id);auto"`
	Keyword             string    `orm:"column(keyword);size(200);null"`
	KeywordHash         string    `orm:"column(keyword_hash);null"`
	Reach               int       `orm:"column(reach);null"`
	ReachIncrease       float32   `orm:"column(reach_increase);null"`
	ReachText           string    `orm:"column(reach_text);size(20);null"`
	Competition         int       `orm:"column(competition);null"`
	CompetitionIncrease float32   `orm:"column(competition_increase);null"`
	CompetitionText     string    `orm:"column(competition_text);size(20);null"`
	HighBid             float32   `orm:"column(high_bid);null"`
	BidIncrease         float32   `orm:"column(bid_increase);null"`
	Date                string    `orm:"column(date);size(12);null"`
	Created             time.Time `orm:"column(created);type(timestamp);null;auto_now_add"`
}

func (t *PbKeywordRecords) TableName() string {
	return "pb_keyword_records"
}

func init() {
	orm.RegisterModel(new(PbKeywordRecords))
}

// AddPbKeywordRecords insert a new PbKeywordRecords into database and returns
// last inserted Id on success.
func AddPbKeywordRecords(m *PbKeywordRecords) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	return
}

// GetPbKeywordRecordsById retrieves PbKeywordRecords by Id. Returns error if
// Id doesn't exist
func GetPbKeywordRecordsById(id int) (v *PbKeywordRecords, err error) {
	o := orm.NewOrm()
	v = &PbKeywordRecords{Id: id}
	if err = o.Read(v); err == nil {
		return v, nil
	}
	return nil, err
}

// GetAllPbKeywordRecords retrieves all PbKeywordRecords matches certain condition. Returns empty list if
// no records exist
func GetAllPbKeywordRecords(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(PbKeywordRecords))
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

	var l []PbKeywordRecords
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

// UpdatePbKeywordRecords updates PbKeywordRecords by Id and returns error if
// the record to be updated doesn't exist
func UpdatePbKeywordRecordsById(m *PbKeywordRecords) (err error) {
	o := orm.NewOrm()
	v := PbKeywordRecords{Id: m.Id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Update(m); err == nil {
			fmt.Println("Number of records updated in database:", num)
		}
	}
	return
}

// DeletePbKeywordRecords deletes PbKeywordRecords by Id and returns error if
// the record to be deleted doesn't exist
func DeletePbKeywordRecords(id int) (err error) {
	o := orm.NewOrm()
	v := PbKeywordRecords{Id: id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Delete(&PbKeywordRecords{Id: id}); err == nil {
			fmt.Println("Number of records deleted in database:", num)
		}
	}
	return
}
