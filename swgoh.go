package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	authRequestOtc = "https://store.galaxy-of-heroes.starwars.ea.com/auth/request_otc"
	authCodeCheck  = "https://store.galaxy-of-heroes.starwars.ea.com/auth/code_check"
	storeOffers    = "https://store.galaxy-of-heroes.starwars.ea.com/store/offers?countryCode="
	StorePurchase  = "https://store.galaxy-of-heroes.starwars.ea.com/store/purchase"
)

type Auth struct {
	AuthId       string
	AuthToken    string
	RefreshToken string
}
type PurchaseResp struct {
	CurrencyCode             string
	CurrencyType             string
	RealMoneyPurchaseCartURL string
	State                    string
}
type Store struct {
	Items                   []Item
	SharedBucketItemDetails SharedBucketItemDetails
	CountryCode             string
	CurrencyUpdates         CurrencyUpdates
	PackOddsPrefixUrl       string
	RealMoneyCurrencyCode   string
}
type Item struct {
	Id                  string
	Name                string
	Description         string
	Image               string
	Order               int
	StoreTab            string
	StartTime           int64
	EndTime             int64
	PromoText1          string
	Guarantee           string
	DetailedDescription string
	QuantityImage       string
	Quantity            string
	BonusQuantity       string
	ShowDetails         bool
	SpecialValue        string
	PackOddsIdentifier  string

	Offers      []Offer
	BucketItems []BucketItem
}
type Offer struct {
	InAppProductId   string
	CurrencyType     string
	Price            float32
	AvailableAtEpoch int64
	LocalPrice       float32
	CountryCode      string
	CurrencyCode     string
}
type BucketItem struct {
	Id       string
	Quantity string
}
type SharedBucketItemDetails struct {
}
type CurrencyUpdates struct {
	SOCIAL  int
	PREMIUM int
}
type PurchaseReq struct {
	ItemId       string `json:"itemId"`
	CurrencyType string `json:"currencyType"`
	CurrencyCode string `json:"currencyCode"`
	RequestId    string `json:"requestId"`
	CountryCode  string `json:"countryCode"`
}
type AuthRequestOtcReq struct {
	Email string `json:"email"`
}
type AuthCodeCheckReq struct {
	Code        string `json:"code"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	CountryCode string `json:"countryCode"`
	RememberMe  bool   `json:"rememberMe"`
}
type ItemClaim struct {
	Name      string
	Message   string
	IsSucceed bool
	StartTime int64
}

func claim(player Player) (items []ItemClaim) {
	itemName := ""
	authCode := sendCode(AuthRequestOtcReq{Email: player.Email})
	if authCode.AuthId != "" && authCode.AuthToken != "" {
		code := ""
		for i := 0; i < 60; i++ {
			time.Sleep(time.Second * 5)
			code = getCodeFromEmail(player.Email)
			if code != "" {
				break
			}
		}
		if code != "" {
			authPlayer := checkCode(authCode, AuthCodeCheckReq{
				Code:        code,
				Email:       player.Email,
				PhoneNumber: "",
				CountryCode: "",
				RememberMe:  true})
			if authPlayer.AuthId != "" && authPlayer.AuthToken != "" {
				store := getStoreOffers(authPlayer)
				for _, item := range store.Items {
					for _, offer := range item.Offers {
						if strings.ToUpper(offer.CurrencyType) == "FREE" {
							countryCode := store.CountryCode
							if countryCode == "" {
								countryCode = "US"
							}
							purchaseReq := PurchaseReq{
								ItemId:       item.Id,
								CurrencyType: offer.CurrencyType,
								CurrencyCode: store.RealMoneyCurrencyCode,
								RequestId:    authCode.AuthId,
								CountryCode:  countryCode,
							}
							itemName = item.Name
							t := time.Now().Unix()
							if offer.AvailableAtEpoch < t && item.EndTime > t && item.StartTime <= t {
								purchaseResp := storePurchase(authPlayer, purchaseReq)
								if purchaseResp.State == "SUCCEEDED" {
									items = append(items, ItemClaim{Name: itemName, Message: "Succeeded", IsSucceed: true})
								} else {
									items = append(items, ItemClaim{Name: itemName, Message: "Failed", IsSucceed: false})
								}
							} else {
								if item.StartTime > t {
									items = append(items, ItemClaim{Name: itemName, Message: "Not yet started", IsSucceed: false, StartTime: item.StartTime})
								} else {
									items = append(items, ItemClaim{Name: itemName, Message: "You have claimed it", IsSucceed: false, StartTime: item.StartTime})
								}
							}
						}
					}
				}
			} else {
				items = append(items, ItemClaim{Name: itemName, Message: "Log in failed", IsSucceed: false})
			}
		} else {
			items = append(items, ItemClaim{Name: itemName, Message: "Get code from email failed", IsSucceed: false})
		}
	} else {
		items = append(items, ItemClaim{Name: itemName, Message: "Send code failed", IsSucceed: false})
	}
	return items
}

func sendCode(authRequestOtcReq AuthRequestOtcReq) Auth {
	header := http.Header{}
	content, err := json.Marshal(authRequestOtcReq)
	checkErr("parse request data(sendCode) error: ", err, Error)
	header, body := postReq(authRequestOtc, "POST", content, header)
	log.Println("sendCode header:", header)
	log.Println("sendCode body:", string(body))

	var auth Auth
	err = json.Unmarshal(body, &auth)
	checkErr("parse auth data(sendCode) error: ", err, Error)

	cookies := header.Values("Set-Cookie")
	for _, cookie := range cookies {
		cc := strings.Split(cookie, ";")
		for _, c := range cc {
			if strings.HasPrefix(c, "authToken=") {
				auth.AuthToken = strings.TrimPrefix(c, "authToken=")
			}
		}
	}
	return auth
}

func checkCode(a Auth, authCodeCheckReq AuthCodeCheckReq) Auth {
	header := http.Header{}
	header.Add("cookie", "authToken="+a.AuthToken)
	header.Add("x-rpc-auth-id", a.AuthId)

	content, err := json.Marshal(authCodeCheckReq)
	checkErr("parse request data(checkCode) error: ", err, Error)

	header, body := postReq(authCodeCheck, "POST", content, header)
	log.Println("checkCode header:", header)
	log.Println("checkCode body:", string(body))

	var auth Auth
	err = json.Unmarshal(body, &auth)
	checkErr("parse request data(checkCode) error: ", err, Error)

	cookies := header.Values("Set-Cookie")
	for _, cookie := range cookies {
		cc := strings.Split(cookie, ";")
		for _, c := range cc {
			if strings.HasPrefix(c, "authToken=") {
				auth.AuthToken = strings.TrimPrefix(c, "authToken=")
			}
		}
	}
	return auth
}

func getStoreOffers(a Auth) Store {
	header := http.Header{}
	header.Add("cookie", "authToken="+a.AuthToken)
	header.Add("x-rpc-auth-id", a.AuthId)

	header, body := postReq(storeOffers, "GET", nil, header)
	log.Println("getStoreOffers header:", header)

	var store Store
	err := json.Unmarshal(body, &store)
	checkErr("parse request data(getStoreOffers) error: ", err, Error)

	return store
}

func storePurchase(player Auth, purchase PurchaseReq) PurchaseResp {
	header := http.Header{}
	header.Add("cookie", "authToken="+player.AuthToken)
	header.Add("x-rpc-auth-id", player.AuthId)

	content, err := json.Marshal(purchase)
	checkErr("parse request data(storePurchase) error: ", err, Error)

	header, body := postReq(StorePurchase, "POST", content, header)
	log.Println("storePurchase header:", header)
	log.Println("storePurchase body:", string(body))

	var purchaseResp PurchaseResp
	err = json.Unmarshal(body, &purchaseResp)
	checkErr("parse request data(storePurchase) error: ", err, Error)

	return purchaseResp
}
