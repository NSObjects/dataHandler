package models

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/astaxie/beego/orm"
)

type ProductCategory struct {
	Id         int    `orm:"column(id);auto"`
	ProductId  uint32 `orm:"column(product_id)"`
	CategoryId int    `orm:"column(category_id);size(30)"`
}

func (t *ProductCategory) TableName() string {
	return "product_category"
}

func init() {
	orm.RegisterModel(new(ProductCategory))
}

type CategoryJSON struct {
	Data struct {
		Categories []struct {
			ChildFilterGroups []struct {
				Filters []struct {
					ChildFilterGroups []struct {
						Filters []struct {
							Name              string `json:"name"`
							FilterID          string `json:"filter_id"`
							ChildFilterGroups []struct {
								Filters []struct {
									Name     string `json:"name"`
									FilterID string `json:"filter_id"`
								} `json:"filters"`
							} `json:"child_filter_groups,omitempty"`
						} `json:"filters"`
					} `json:"child_filter_groups,omitempty"`
					Name     string `json:"name"`
					FilterID string `json:"filter_id"`
				} `json:"filters"`
			} `json:"child_filter_groups"`
			Name     string `json:"name"`
			FilterID string `json:"filter_id"`
		} `json:"categories"`
	} `json:"data"`
}

type Categories struct {
	Datas []CategoriesData `json:"datas"`
}

type CategoriesData struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

// AddProductCategory insert a new ProductCategory into database and returns
// last inserted Id on success.
func AddProductCategory(m *ProductCategory) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	return
}

// GetProductCategoryById retrieves ProductCategory by Id. Returns error if
// Id doesn't exist
func GetProductCategoryById(id int) (v *ProductCategory, err error) {
	o := orm.NewOrm()
	v = &ProductCategory{Id: id}
	if err = o.Read(v); err == nil {
		return v, nil
	}
	return nil, err
}

// GetAllProductCategory retrieves all ProductCategory matches certain condition. Returns empty list if
// no records exist
func GetAllProductCategory(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(ProductCategory))
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

	var l []ProductCategory
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

// UpdateProductCategory updates ProductCategory by Id and returns error if
// the record to be updated doesn't exist
func UpdateProductCategoryById(m *ProductCategory) (err error) {
	o := orm.NewOrm()
	v := ProductCategory{Id: m.Id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Update(m); err == nil {
			fmt.Println("Number of records updated in database:", num)
		}
	}
	return
}

// DeleteProductCategory deletes ProductCategory by Id and returns error if
// the record to be deleted doesn't exist
func DeleteProductCategory(id int) (err error) {
	o := orm.NewOrm()
	v := ProductCategory{Id: id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Delete(&ProductCategory{Id: id}); err == nil {
			fmt.Println("Number of records deleted in database:", num)
		}
	}
	return
}
