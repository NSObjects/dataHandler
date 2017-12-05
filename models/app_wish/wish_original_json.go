package models

import "time"

type WishOrginalData struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
	Data struct {
		Contest struct {
			ProductRating struct {
				Rating      float64 `json:"rating"`
				RatingCount float64 `json:"rating_count"`
				RatingClass string  `json:"rating_class"`
			} `json:"product_rating"`
			Keywords            string `json:"keywords"`
			ID                  string `json:"id"`
			CommerceProductInfo struct {
				Variations []Variations `json:"variations"`
			} `json:"commerce_product_info"`
			ExtraPhotoUrls struct {
				Num1  string `json:"1"`
				Num2  string `json:"2"`
				Num3  string `json:"3"`
				Num4  string `json:"4"`
				Num5  string `json:"5"`
				Num6  string `json:"6"`
				Num7  string `json:"7"`
				Num8  string `json:"8"`
				Num9  string `json:"9"`
				Num10 string `json:"10"`
				Num11 string `json:"11"`
				Num12 string `json:"12"`
				Num13 string `json:"13"`
				Num14 string `json:"14"`
				Num15 string `json:"15"`
				Num16 string `json:"16"`
				Num17 string `json:"17"`
				Num18 string `json:"18"`
			} `json:"extra_photo_urls"`
			NumBought        int      `json:"num_bought"`
			TrueTagIds       []string `json:"true_tag_ids"`
			CurrentlyViewing struct {
				Message     string   `json:"message"`
				MessageList []string `json:"message_list"`
				RefreshRate int      `json:"refresh_rate"`
			} `json:"currently_viewing"`
			IsVerified  bool   `json:"is_verified"`
			Description string `json:"description"`
			Tags        []struct {
				Name string `json:"name"`
			} `json:"tags"`
			DisplayPicture string    `json:"display_picture"`
			GenerationTime time.Time `json:"generation_time"`
			MerchantTags   []struct {
				Name string `json:"name"`
			} `json:"merchant_tags"`
			Name       string `json:"name"`
			Gender     string `json:"gender"`
			NumEntered int    `json:"num_entered"`
		} `json:"contest"`
	} `json:"data"`
}

type Variations struct {
	VariationID      string  `json:"variation_id"`
	Color            string  `json:"color"`
	MaxShippingTime  int     `json:"max_shipping_time"`
	MinShippingTime  int     `json:"min_shipping_time"`
	Size             string  `json:"size"`
	MerchantName     string  `json:"merchant_name"`
	Merchant         string  `json:"merchant"`
	ShipsFrom        string  `json:"ships_from"`
	Price            float32 `json:"price"`
	IsWishExpress    bool    `json:"is_wish_express"`
	RetailPrice      float32 `json:"retail_price"`
	Shipping         float32 `json:"shipping"`
	VariableShipping float32 `json:"variable_shipping"`
	OriginalPrice    float32 `json:"original_price"`
	SizeOrdering     float32 `json:"size_ordering"`
	Inventory        int     `json:"inventory"`
	OriginalShipping float32 `json:"original_shipping"`
}
