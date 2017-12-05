package models

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
)

type ProductVariations struct {
	Id            int       `orm:"column(id);auto"`
	ProductId     uint32    `orm:"column(product_id)"`
	Created       time.Time `orm:"column(created);type(datetime)"`
	WishShipping  int       `orm:"column(wish_shipping)"`
	OwnerShipping int       `orm:"column(owner_shipping)"`
	Title         int       `orm:"column(title)"`
	WishPrice     int       `orm:"column(wish_price)"`
	OwnerPrice    int       `orm:"column(owner_price)"`
	Tag           int       `orm:"column(tag)"`
	Category      int       `orm:"column(category)"`
	WishVerified  int       `orm:"column(wish_verified)"`
	IsVariations  int       `orm:"column(is_variations)"`
	Sku           int       `orm:"column(sku)"`
	NumBought     int       `orm:"column(num_bought)"`
	MerchantTags  int       `orm:"column(merchant_tags);null"`
	Time          string    `orm:"-" json:"created"`
}

func (t *ProductVariations) TableName() string {
	return "product_variations"
}

func init() {
	orm.RegisterModel(new(ProductVariations))
}

// AddProductVariations insert a new ProductVariations into database and returns
// last inserted Id on success.
func AddProductVariations(m *ProductVariations) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	return
}

// GetProductVariationsById retrieves ProductVariations by Id. Returns error if
// Id doesn't exist
func GetProductVariationsById(id int) (v *ProductVariations, err error) {
	o := orm.NewOrm()
	v = &ProductVariations{Id: id}
	if err = o.Read(v); err == nil {
		return v, nil
	}
	return nil, err
}

// GetAllProductVariations retrieves all ProductVariations matches certain condition. Returns empty list if
// no records exist
func GetAllProductVariations(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(ProductVariations))
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

	var l []ProductVariations
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

// UpdateProductVariations updates ProductVariations by Id and returns error if
// the record to be updated doesn't exist
func UpdateProductVariationsById(m *ProductVariations) (err error) {
	o := orm.NewOrm()
	v := ProductVariations{Id: m.Id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Update(m); err == nil {
			fmt.Println("Number of records updated in database:", num)
		}
	}
	return
}

// DeleteProductVariations deletes ProductVariations by Id and returns error if
// the record to be deleted doesn't exist
func DeleteProductVariations(id int) (err error) {
	o := orm.NewOrm()
	v := ProductVariations{Id: id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Delete(&ProductVariations{Id: id}); err == nil {
			fmt.Println("Number of records deleted in database:", num)
		}
	}
	return
}
