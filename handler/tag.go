package handler

import (
	"dataHandler/db"
	"dataHandler/models/app_wish"

	"github.com/Sirupsen/logrus"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

var allCategorys map[string]models.CategoriesData

var allTag map[string]models.CategoriesData

func init() {
	ac := allCategory()
	allCategorys = make(map[string]models.CategoriesData)
	for _, v := range ac.Datas {
		name := strings.ToLower(v.Name)
		if len(name) > 0 {
			allCategorys[name] = v
		}

	}

	at := allTags()
	allTag = make(map[string]models.CategoriesData)
	for _, v := range at.Datas {
		name := strings.ToLower(v.Name)
		if len(name) > 0 {
			allTag[name] = v
		}

	}

	local, err := time.LoadLocation("Asia/Shanghai")

	if err != nil {
		fmt.Print(err)
	}
	time.Local = local

}

func categoryByProductTag(tags []string, productId uint32) {

	for _, t := range tags {
		t = strings.ToLower(t)
		if c := allCategorys[t]; len(c.Name) > 0 {
			if len(c.Id) <= 0 {
				continue
			}

			category := models.Category{Name: c.Name}
			if created, _, err := db.AppWish.ReadOrCreate(&category, "name"); err == nil {
				if created {
					category.FilterId = c.Id
					if _, err := db.AppWish.Update(&category); err != nil {
						if strings.Contains(err.Error(), "Duplicate entry") == false {
							log.WithFields(logrus.Fields{
								"tag.go": 67,
							}).Error(err)
						}
					}
				}

				pc := models.ProductCategory{
					CategoryId: category.Id,
					ProductId:  productId,
				}
				if _, err := db.AppWish.Insert(&pc); err != nil {
					if strings.Contains(err.Error(), "Duplicate entry") == false {
						log.WithFields(logrus.Fields{
							"tag.go": 79,
						}).Error(err)
					}
				}
			}
		}

		if c := allCategorys[t]; len(c.Name) <= 0 {
			tag := models.Tag{Name: t}

			if _, id, err := db.AppWish.ReadOrCreate(&tag, "name"); err == nil {

				pt := models.ProductTag{
					TagId:     id,
					ProductId: productId,
					Created:   time.Now(),
					Updated:   time.Now(),
				}
				if _, err := db.AppWish.Insert(&pt); err != nil {
					if strings.Contains(err.Error(), "Duplicate entry") == false {
						log.WithFields(logrus.Fields{
							"tag.go": 98,
						}).Error(err)
					}
				}
			} else {
				if strings.Contains(err.Error(), "Duplicate entry") == false {
					log.WithFields(logrus.Fields{
						"tag.go": 105,
					}).Error(err)
				}
			}
		}

	}

}

func allCategory() models.Categories {
	b, _ := ioutil.ReadFile("category.json")
	var c models.CategoryJSON
	json.Unmarshal(b, &c)

	d := models.Categories{}

	for _, v1 := range c.Data.Categories {
		if len(v1.ChildFilterGroups) == 0 {
			data := models.CategoriesData{
				Name: v1.Name,
				Id:   v1.FilterID,
			}
			d.Datas = append(d.Datas, data)
		} else {
			for _, v2 := range v1.ChildFilterGroups {
				for _, v3 := range v2.Filters {
					if len(v3.ChildFilterGroups) == 0 {
						data := models.CategoriesData{
							Name: v3.Name,
							Id:   v3.FilterID,
						}
						d.Datas = append(d.Datas, data)
					} else {
						for _, v4 := range v3.ChildFilterGroups {
							for _, v5 := range v4.Filters {
								if len(v5.ChildFilterGroups) == 0 {
									data := models.CategoriesData{
										Name: v5.Name,
										Id:   v5.FilterID,
									}
									d.Datas = append(d.Datas, data)
								} else {
									for _, v6 := range v5.ChildFilterGroups {
										for _, v7 := range v6.Filters {
											data := models.CategoriesData{
												Name: v7.Name,
												Id:   v7.FilterID,
											}
											d.Datas = append(d.Datas, data)
										}
									}
								}
							}
						}
					}
				}

			}
		}

	}

	return d
}

func allTags() models.Categories {
	b, _ := ioutil.ReadFile("category.json")
	var c models.CategoryJSON
	json.Unmarshal(b, &c)

	d := models.Categories{}

	for _, v1 := range c.Data.Categories {
		data := models.CategoriesData{
			Name: v1.Name,
			Id:   v1.FilterID,
		}
		d.Datas = append(d.Datas, data)

		for _, v2 := range v1.ChildFilterGroups {

			for _, v3 := range v2.Filters {
				data := models.CategoriesData{
					Name: v3.Name,
					Id:   v3.FilterID,
				}
				d.Datas = append(d.Datas, data)

				for _, v4 := range v3.ChildFilterGroups {

					for _, v5 := range v4.Filters {
						data := models.CategoriesData{
							Name: v5.Name,
							Id:   v5.FilterID,
						}
						d.Datas = append(d.Datas, data)
						for _, v6 := range v5.ChildFilterGroups {
							for _, v7 := range v6.Filters {
								data := models.CategoriesData{
									Name: v7.Name,
									Id:   v7.FilterID,
								}
								d.Datas = append(d.Datas, data)
							}
						}

					}
				}

			}

		}

	}

	return d
}

func stringInSlice(a string, list []models.CategoriesData) (c *models.CategoriesData) {
	for _, b := range list {
		if strings.ToUpper(b.Name) == strings.ToUpper(a) {
			return &b
		}
	}
	return nil
}
