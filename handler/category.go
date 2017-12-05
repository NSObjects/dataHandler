package handler

import (
	"dataHandler/db"
	"dataHandler/models/app_wish"
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
)

var ac = allCategory()

func categoryByProductTrueTag(tags []string, productId uint32) {

	allCategorys := make(map[string]models.CategoriesData)
	for _, v := range ac.Datas {
		if len(v.Id) > 0 {
			allCategorys[v.Id] = v
		}
	}

	for _, t := range tags {
		if c := allCategorys[t]; len(c.Id) > 0 {
			insertTag(c.Id, productId)
		}
	}
}

func insertTag(filterId string, productId uint32) {

	c := models.Category{FilterId: filterId}
	if err := db.AppWish.Read(&c, "filter_id"); err == nil {
		categoryProduct := models.ProductCategory{
			CategoryId: c.Id,
			ProductId:  productId,
		}
		if _, err := db.AppWish.Insert(&categoryProduct); err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") == false {
				log.WithFields(logrus.Fields{
					"cacheIncremental.go": 40,
				}).Error(err)
			}
		} else {
			fmt.Println("插入成功:", categoryProduct.Id)
		}
	}
}
