package unisport

import (
	"context"
	"encoding/json"
	"fmt"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	dishooks "github.com/itsTurnip/dishooks"
	"go.mongodb.org/mongo-driver/bson"
	mongo2 "go.mongodb.org/mongo-driver/mongo"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"monitors/mongo"
	"monitors/mongo/models"
	"monitors/sites"
	"reflect"
	"time"
)

func Filtered(server string) {
	mongoClient, err := mongo.GetMongoClient()
	if err != nil {
		log.Println("Error getting mongo client: %v", err)
	}
	productCollection := mongoClient.Database("Monitors").Collection("Unisport")
	var webhookSettings models.WebhookSettings
	err = mongoClient.Database("Monitors").Collection("Settings").FindOne(context.TODO(), bson.M{"Server": server, "Store": "Unisport-Filtered"}).Decode(&webhookSettings)
	if err != nil {
		log.Println("Error getting webhook settings: %v", err)
		return
	}
	for {
		var settings models.MonitorSettings
		err = mongoClient.Database("Monitors").Collection("Settings").FindOne(context.TODO(), bson.M{"Store": "Unisport-Filtered"}).Decode(&settings)
		if err != nil {
			log.Println("Error getting mongo client: %v", err)
		}
		var baseUrl = "https://www.unisport.dk/api/products/batch/?list="

		for index, pid := range settings.PIDS {
			pidLength := len(settings.PIDS)
			if pidLength > 1 {
				if index == pidLength-1 {
					baseUrl += pid
				} else {
					baseUrl += pid + ","
				}
			} else {
				baseUrl += pid
			}
		}

		// DO REQUEST
		prxy := sites.GetProxy()
		options := []tls_client.HttpClientOption{
			tls_client.WithTimeout(30),
			tls_client.WithClientProfile(tls_client.Chrome_105),
			tls_client.WithNotFollowRedirects(),
			tls_client.WithProxyUrl(prxy),
			tls_client.WithInsecureSkipVerify(),
		}

		client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
		if err != nil {
			log.Println(err)
			return
		}

		req, err := http.NewRequest(http.MethodGet, baseUrl, nil)
		if err != nil {
			log.Println(err)
			return
		}

		req.Header = http.Header{
			"accept":          {"*/*"},
			"accept-language": {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
			"user-agent":      {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.75 Safari/537.36"},
			http.HeaderOrderKey: {
				"accept",
				"accept-language",
				"user-agent",
			},
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			return
		}

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {

			}
		}(resp.Body)

		log.Println(fmt.Sprintf("status code: %d", resp.StatusCode))

		readBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			return
		}

		var products map[string]any
		err = json.Unmarshal(readBytes, &products)
		if err != nil {
			log.Println(err)
			return
		}

		for _, product := range products["products"].([]interface{}) {
			var i models.UnisportMonitorStruct
			err := productCollection.FindOne(context.TODO(), bson.M{"pid": product.(map[string]interface{})["id"].(string)}).Decode(&i)

			var stock []UnisportStockInfo
			for _, stockType := range product.(map[string]interface{})["stock"].([]interface{}) {
				stock = append(stock, UnisportStockInfo{Size: stockType.(map[string]interface{})["name_short"].(string), Stock: stockType.(map[string]interface{})["stock_info"].(string)})
			}
			//hvis produkt ik i DB vil den throw error, derfor tjekker vi om det err er fordi den ik er i DB, hvis ik det den err, så log den anden error
			if err != nil {
				if err == mongo2.ErrNoDocuments {
					price := product.(map[string]interface{})["prices"].(map[string]interface{})["max_price"]
					sendFilteredUnisportWebhook(product.(map[string]interface{})["name"].(string), product.(map[string]interface{})["url"].(string), product.(map[string]interface{})["image"].(string), fmt.Sprintf("%.2f", price.(float64)), stock, webhookSettings.Webhooks)
					_, err := productCollection.InsertOne(context.TODO(), bson.M{"pid": product.(map[string]interface{})["id"].(string), "stock": stock})
					if err != nil {
						log.Println("error on inserting product: " + err.Error())
					}
				} else {
					log.Println(err)
				}
			}

			//hvis produktet er i DB, tjekker vi om der er sket ændringer i stock
			if err == nil {
				var savedStock []UnisportStockInfo
				for _, stockType := range i.Stock {
					savedStock = append(savedStock, UnisportStockInfo{Size: stockType.Size, Stock: stockType.Stock})
				}

				log.Println(reflect.DeepEqual(savedStock, stock))
				if !reflect.DeepEqual(savedStock, stock) {
					price := product.(map[string]interface{})["prices"].(map[string]interface{})["max_price"]
					sendFilteredUnisportWebhook(product.(map[string]interface{})["name"].(string), product.(map[string]interface{})["url"].(string), product.(map[string]interface{})["image"].(string), fmt.Sprintf("%.2f", price.(float64)), stock, webhookSettings.Webhooks)
					_, err := productCollection.UpdateOne(context.TODO(), bson.M{"pid": product.(map[string]interface{})["id"].(string)}, bson.M{"$set": bson.M{"stock": stock}})
					if err != nil {
						log.Println("error on updating product: " + err.Error())
					}
				}
			}

			time.Sleep(30 * time.Second)
		}
	}
}

func sendFilteredUnisportWebhook(productName string, productUrl string, productImage string, productPrice string, productStock []UnisportStockInfo, webhooks []string) {
	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
	webhookUrl := webhooks[rand.Intn(len(webhooks))]
	webhook, err := sites.WebhookFromURL(webhookUrl)
	//create stock string
	var stockString string
	for _, stock := range productStock {
		stockString += stock.Size + " | " + stock.Stock + "\n"
	}

	priceField := dishooks.EmbedField{
		Name:   "PRICE:",
		Value:  productPrice + " DKK",
		Inline: true,
	}
	stockField := dishooks.EmbedField{
		Name:   "[SIZE] | [STOCK]:",
		Value:  stockString,
		Inline: false,
	}

	embed := dishooks.Embed{
		Fields: []*dishooks.EmbedField{
			&priceField,
			&stockField,
		},
		Author: &dishooks.EmbedAuthor{
			Name: "Unisport",
			URL:  "https://www.unisport.dk/",
		},
		Thumbnail: &dishooks.EmbedThumbnail{
			URL: productImage,
		},
		Footer: &dishooks.EmbedFooter{
			Text: "SCANDI MONITORS",
		},
		Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		Color:     4976907,
		URL:       productUrl,
		Title:     productName,
	}
	_, err = webhook.SendMessage(&dishooks.WebhookMessage{
		Username: "Unisport Filtered",
		Embeds:   []*dishooks.Embed{&embed},
	})
	if err != nil {
		log.Println()
	}

}

type UnisportStockInfo struct {
	Size  string
	Stock string
}
