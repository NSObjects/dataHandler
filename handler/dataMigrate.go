package handler

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/astaxie/beego/orm"

	"dataHandler/db"
	"dataHandler/models/app_wish"
	"dataHandler/models/new_wish"
	"dataHandler/util"
)

func MigrateSnapshot(date time.Time) {

	lastId := 0
	for {

		var tempSnapshot []tmodels.TProductSnapshot
		count, err := db.DataSource.QueryTable("t_product_snapshot").
			Filter("created", date).
			Filter("id__gt", lastId).
			OrderBy("id").
			Limit(1000).
			All(&tempSnapshot)

		if count == 0 || err != nil || lastId == tempSnapshot[len(tempSnapshot)-1].Id {
			if err != nil {
				log.WithFields(logrus.Fields{
					"dataMigrate.go": 37,
				}).Error(err)
			}
			break
		}

		lastId = tempSnapshot[len(tempSnapshot)-1].Id

		var wg sync.WaitGroup
		p := util.New(20)
		for _, snapshot := range tempSnapshot {
			wg.Add(1)
			go func(snapshot tmodels.TProductSnapshot) {
				p.Run(func() {
					var s models.ProductSnapshot
					if len(snapshot.Data) > 0 {
						s.Data = string(UnzipBytes([]byte(snapshot.Data)))
						if len(s.Data) <= 0 {
							return
						}
					}

					s.Created = snapshot.Created
					s.WishId = snapshot.WishId
					go updateProduct(s.Data)
					if _, err := db.AppWish.Insert(&s); err != nil {
						if strings.Contains(err.Error(), "Error 1062: Duplicate entry") == false {
							log.WithFields(logrus.Fields{
								"dataMigrate.go": 65,
							}).Error(err)
						}
					}
					wg.Done()
				})

			}(snapshot)
		}

		wg.Wait()
		p.Shutdown()

	}

}

func VarationHandler(date time.Time) {

	size := 1000
	page := 0
	for {
		var varations []models.ProductSnapshot
		_, err := db.AppWish.QueryTable("product_snapshot").
			Filter("created", date).
			OrderBy("created").
			Limit(size, size*page).
			All(&varations)
		if err != nil || len(varations) <= 0 {
			break
		}
		page++
		var wg sync.WaitGroup
		p := util.New(NumConcurrent)
		for _, v := range varations {
			wg.Add(1)
			go func(v models.ProductSnapshot, date time.Time) {
				p.Run(func() {
					handler(v, date)
				})
				wg.Done()
			}(v, date)
		}

		wg.Wait()
		p.Shutdown()
	}

}

func UnzipBytes(input []byte) []byte {

	b := bytes.NewReader(input)
	r, err := zlib.NewReader(b)
	if err != nil {
		log.WithFields(logrus.Fields{
			"dataMigrate.go": 469,
		}).Error(err)
	}
	defer r.Close()

	data, _ := ioutil.ReadAll(r)
	return data
}

func handler(v models.ProductSnapshot, date time.Time) {

	var varation models.ProductSnapshot
	err := db.AppWish.QueryTable("product_snapshot").
		Filter("created", time.Unix(date.Unix()-24*60*60, 0)).
		Filter("wish_id", v.WishId).
		One(&varation)
	if err != nil {
		if err != orm.ErrNoRows {
			log.WithFields(logrus.Fields{
				"dataMigrate.go": 140,
			}).Error(err)
		}
		return
	}

	var yestdaySnapshot models.WishOrginalData
	if err = json.Unmarshal([]byte(v.Data), &yestdaySnapshot); err == nil {
		product := models.Product{WishId: yestdaySnapshot.Data.Contest.ID}

		if err := db.AppWish.Read(&product, "wish_id"); err == nil {
			product.NumBought = yestdaySnapshot.Data.Contest.NumBought
			product.NumEntered = yestdaySnapshot.Data.Contest.NumEntered
			product.TrueTagIds = strings.Join(yestdaySnapshot.Data.Contest.TrueTagIds, ",")
			var tags []string
			for _, tag := range yestdaySnapshot.Data.Contest.Tags {
				tags = append(tags, tag.Name)
			}

			if len(tags) > 0 {
				product.Tags = strings.Join(tags, ",")
			}

			var merchantTags []string
			for _, tag := range yestdaySnapshot.Data.Contest.MerchantTags {
				merchantTags = append(merchantTags, tag.Name)
			}

			if len(merchantTags) > 0 {
				product.MerchantTags = strings.Join(merchantTags, ",")
			}

			product.Name = yestdaySnapshot.Data.Contest.Name
			product.RatingCount = int(yestdaySnapshot.Data.Contest.ProductRating.RatingCount)
			for _, v := range yestdaySnapshot.Data.Contest.CommerceProductInfo.Variations {
				if v.OriginalPrice > 0 {
					product.Price = v.OriginalPrice
				}

				if v.RetailPrice > 0 {
					product.RetailPrice = v.RetailPrice
				}

				if v.Shipping > 0 {
					product.Shipping = v.Shipping
				}
			}

			if _, err := db.AppWish.Update(&product); err != nil {
				log.WithFields(logrus.Fields{
					"dataMigrate.go": 180,
				}).Error(err)
			}

		}

		var lastDaySnapshot models.WishOrginalData
		if err = json.Unmarshal([]byte(varation.Data), &lastDaySnapshot); err == nil {
			productVaration := models.ProductVariations{
				ProductId: util.FNV(v.WishId),
			}

			yestDayVariation := yestdaySnapshot.Data.Contest.CommerceProductInfo.Variations
			lastDayVariation := lastDaySnapshot.Data.Contest.CommerceProductInfo.Variations

			for _, y := range yestDayVariation {
				for _, l := range lastDayVariation {
					if y.VariationID == l.VariationID {
						if y.OriginalPrice != l.OriginalPrice {
							productVaration.OwnerPrice = 1
						}

						if y.Price != l.Price {
							productVaration.WishPrice = 1
						}

						if y.VariableShipping != l.VariableShipping {
							productVaration.OwnerShipping = 1
						}

						if y.Shipping != l.Shipping {
							productVaration.WishShipping = 1
						}

					}
				}
			}

			if len(yestDayVariation) == len(lastDayVariation) {
				for _, ytag := range yestDayVariation {
					exits := false
					for _, ltag := range lastDayVariation {
						if ytag.VariationID == ltag.VariationID {
							exits = true
							break
						}
					}

					if exits == false {
						productVaration.Sku = 1
						break
					}

				}
			} else {
				productVaration.Sku = 1
			}

			if len(yestdaySnapshot.Data.Contest.Tags) == len(lastDaySnapshot.Data.Contest.Tags) {
				for _, ytag := range yestdaySnapshot.Data.Contest.Tags {
					exits := false
					for _, ltag := range lastDaySnapshot.Data.Contest.Tags {
						if ytag == ltag {
							exits = true
							break
						}
					}

					if exits == false {
						productVaration.Tag = 1
						break
					}

				}
			} else {
				productVaration.Tag = 1
			}

			if len(yestdaySnapshot.Data.Contest.MerchantTags) == len(lastDaySnapshot.Data.Contest.MerchantTags) {
				for _, ytag := range yestdaySnapshot.Data.Contest.MerchantTags {
					exits := false
					for _, ltag := range lastDaySnapshot.Data.Contest.MerchantTags {
						if ytag == ltag {
							exits = true
							break
						}
					}

					if exits == false {
						productVaration.Tag = 1
						break
					}

				}
			} else {
				productVaration.Tag = 1
			}

			if yestdaySnapshot.Data.Contest.IsVerified != lastDaySnapshot.Data.Contest.IsVerified {
				productVaration.WishVerified = 1
			}

			for _, ytag := range yestdaySnapshot.Data.Contest.TrueTagIds {
				exits := false
				for _, ltag := range lastDaySnapshot.Data.Contest.TrueTagIds {
					if ytag == ltag {
						exits = true
						break
					}
				}

				if exits == false {
					productVaration.Category = 1
					break
				}

			}

			if yestdaySnapshot.Data.Contest.Name != lastDaySnapshot.Data.Contest.Name {
				productVaration.Title = 1
			}
			productVaration.Created = date
			productVaration.NumBought = yestdaySnapshot.Data.Contest.NumBought

			_, err := db.AppWish.Insert(&productVaration)
			if err != nil {
				if strings.Contains(err.Error(), "Error 1062: Duplicate entry") == false {
					log.WithFields(logrus.Fields{
						"dataMigrate.go": 318,
					}).Error(err)
				}
			}

		} else {
			log.WithFields(logrus.Fields{
				"dataMigrate.go": 325,
			}).Error(err)
		}
	} else {
		log.WithFields(logrus.Fields{
			"dataMigrate.go": 330,
		}).Error(err)
	}
}

func updateProduct(data string) {
	if len(data) <= 0 {
		return
	}
	var wishProductJson models.WishOrginalData
	err := json.Unmarshal([]byte(data), &wishProductJson)
	if err != nil {
		log.WithFields(logrus.Fields{
			"dataMigrate.go": 343,
		}).Error(err)
		return
	}
	var product models.Product
	err = db.AppWish.QueryTable("product").
		Filter("id", wishProductJson.Data.Contest.ID).
		One(&product)

	if err != nil {
		//if err != orm.ErrNoRows {
		//	return
		//}
		product := configProduct(wishProductJson)
		if len(product.Name) <= 0 || len(product.WishId) <= 0 {
			return
		}
		product.Id = util.FNV(product.WishId)
		product.Created = time.Now()
		_, err = db.AppWish.Insert(&product)
		if err != nil {
			log.WithFields(logrus.Fields{
				"dataMigrate.go": 354,
			}).Error(err)
		}
	} else {

		product := configProduct(wishProductJson)
		_, err := db.AppWish.Update(&product)
		if err != nil {
			log.WithFields(logrus.Fields{
				"dataMigrate.go": 363,
			}).Error(err)
		}
	}

	var tags []string

	if tags = strings.Split(product.Tags, ","); len(tags) <= 0 {
		return
	}

	if merchantTags := strings.Split(product.MerchantTags, ","); len(merchantTags) > 0 {
		for _, v := range merchantTags {
			tags = append(tags, v)
		}
	}

	//categoryByProductTag(tags, product.Id)
	//categoryByProductTrueTag(strings.Split(product.TrueTagIds, ","), product.Id)
}

func configProduct(productJson models.WishOrginalData) (p models.Product) {
	if len(productJson.Data.Contest.Name) <= 0 || len(productJson.Data.Contest.ID) <= 0 {
		return p
	}

	p.WishId = productJson.Data.Contest.ID

	var size []string
	var color []string
	for _, va := range productJson.Data.Contest.CommerceProductInfo.Variations {
		if len(va.Size) > 0 {
			size = append(size, va.Size)
		}
		if len(va.Color) > 0 {
			color = append(color, va.Color)
		}
		if va.Price > 0 {
			p.Price = va.Price
		}
		if va.MaxShippingTime > 0 && va.MinShippingTime > 0 {
			p.MaxShippingTime = va.MaxShippingTime
			p.MinShippingTime = va.MinShippingTime
		}
		if va.RetailPrice > 0 {
			p.RetailPrice = va.RetailPrice
		}
		//if len(va.Merchant) > 0 {
		//	p.Merchant = va.Merchant
		//}
		if len(va.MerchantName) > 0 {
			p.Merchant = va.MerchantName
		}
		if va.Shipping > 0 {
			p.Shipping = va.Shipping
		}
	}

	p.Updated = time.Now()
	size = removeDuplicatesUnordered(size)
	color = removeDuplicatesUnordered(color)
	p.Size = strings.Join(size, ",")
	p.Color = strings.Join(color, ",")

	switch productJson.Data.Contest.Gender {
	case "male":
		p.Gender = 1
	case "female":
		p.Gender = 2
	case "neutral":
		p.Gender = 3
	}

	var tags []string
	for _, tag := range productJson.Data.Contest.Tags {
		if len(tag.Name) > 0 {
			tags = append(tags, tag.Name)
		}
	}

	p.Tags = strings.Join(tags, ",")
	var merchantTags []string
	for _, tag := range productJson.Data.Contest.MerchantTags {
		if len(tag.Name) > 0 {
			merchantTags = append(merchantTags, tag.Name)
		}
	}

	p.TrueTagIds = strings.Join(productJson.Data.Contest.TrueTagIds, ",")
	p.WishId = productJson.Data.Contest.ID
	p.GenerationTime = productJson.Data.Contest.GenerationTime
	p.Description = productJson.Data.Contest.Description
	p.MerchantTags = strings.Join(merchantTags, ",")
	p.Name = productJson.Data.Contest.Name
	p.RatingCount = int(productJson.Data.Contest.ProductRating.RatingCount)
	p.NumEntered = productJson.Data.Contest.NumEntered
	p.NumBought = productJson.Data.Contest.NumBought
	return
}

func removeDuplicatesUnordered(elements []string) []string {
	encountered := map[string]bool{}
	for v := range elements {
		encountered[elements[v]] = true
	}
	result := []string{}
	for key := range encountered {
		result = append(result, key)
	}
	return result
}
