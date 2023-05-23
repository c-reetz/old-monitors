package ralph_lauren

import (
	"context"
	"fmt"
	"github.com/antchfx/htmlquery"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/itsTurnip/dishooks"
	"go.mongodb.org/mongo-driver/bson"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"monitors/mongo"
	"monitors/mongo/models"
	"monitors/sites"
	"strconv"
	"strings"
	"time"
)

func MensAccessory(server string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in MensAccessory", r)
		}
		for {
			mensaccesorryLauncher(server)
		}
	}()
}

func mensaccesorryLauncher(server string) {
	fmt.Println("Launching Ralph Lauren Mens Accessory Monitor")
	mongoClient, err := mongo.GetMongoClient()
	if err != nil {
		log.Println("Error getting mongo client: %v", err)
	}
	//productCollection := mongoClient.Database("Monitors").Collection("Ralph Lauren")
	var webhookSettings models.WebhookSettings
	err = mongoClient.Database("Monitors").Collection("Settings").FindOne(context.TODO(), bson.M{"Server": server, "Store": "Ralph Lauren"}).Decode(&webhookSettings)

	if err != nil {
		log.Println("[RALPH][%v] Error getting webhook settings: %v", server, err)

	}

	var previous models.RalphPreviousItemsStruct

	//LOOP START
	for {
		err = mongoClient.Database("Monitors").Collection("Ralph Lauren").FindOne(context.TODO(), bson.M{"Subsection": "Mens Accessory"}).Decode(&previous)

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

		}

		req, err := http.NewRequest(http.MethodGet, "https://www.ralphlauren.eu/dk/en/men/accessories/1030?srule=price-low-high&start=0&sz=32&webcat=Men%7CAccessories&format=ajax", nil)
		if err != nil {
			log.Println(err)

		}

		req.Header = http.Header{
			"accept":           {"*/*"},
			"accept-language":  {"en-DK,en;q=0.9,da-DK;q=0.8,da;q=0.7,en-US;q=0.6"},
			"user-agent":       {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36"},
			"accept-encoding":  {"gzip, deflate, br"},
			"origin":           {"https://www.ralphlauren.eu/"},
			"sec-fetch-dest":   {"empty"},
			"sec-fetch-mode":   {"cors"},
			"sec-fetch-site":   {"same-origin"},
			"x-requested-with": {"XMLHttpRequest"},
			http.HeaderOrderKey: {
				"accept",
				"accept-language",
				"accept-encoding",
				"user-agent",
				"origin",
				"sec-fetch-dest",
				"sec-fetch-mode",
				"sec-fetch-site",
				"x-requested-with",
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

		bod, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)

		}

		doc, err := htmlquery.Parse(strings.NewReader(string(bod)))
		if err != nil {
			log.Println(err)

		}

		productList := htmlquery.Find(doc, "//div[@data-pname]")

		var products []string

		for _, product := range productList {
			productName := htmlquery.SelectAttr(product, "data-pname")
			productStock := htmlquery.SelectAttr(product, "data-stockmsg")
			productUrl := htmlquery.SelectAttr(product, "data-monetate-producturl")

			img := htmlquery.FindOne(product, "//img")
			productImage := htmlquery.SelectAttr(img, "src")

			priceInput := htmlquery.FindOne(product, "//input")
			path := strings.Split(htmlquery.SelectAttr(priceInput, "value"), " ")
			price, err := strconv.ParseFloat(path[0], 64)
			if err != nil {
				log.Println(err)
			}
			products = append(products, productName)
			if price < 200 {
				isInPreviousLoop := contains(previous.Items, productName)
				if !isInPreviousLoop {
					if strings.Contains(productName, "Sock") || strings.Contains(productName, "Pin of Solidarity") || strings.Contains(productName, "Trunks") {
						continue
					}
					sendPriceErrorRalphMensAccessory(productName, productStock, productUrl, productImage, path[0], webhookSettings.Webhooks)
				}
			} else {
				continue
			}
		}
		_, err = mongoClient.Database("Monitors").Collection("Ralph Lauren").UpdateOne(context.TODO(), bson.M{"Subsection": "Mens Accessory"}, bson.M{"$set": bson.M{"Items": products}})
		if err != nil {
			log.Println("error on updating products: " + err.Error())
		}
		time.Sleep(time.Duration(2 * time.Minute))
	}
}

func sendPriceErrorRalphMensAccessory(productName string, productStock string, productUrl string, productImg string, productPrice string, webhooks []string) {
	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
	webhookUrl := webhooks[rand.Intn(len(webhooks))]
	webhook, err := sites.WebhookFromURL(webhookUrl)

	priceField := dishooks.EmbedField{
		Name:   "PRICE:",
		Value:  productPrice + " DKK",
		Inline: true,
	}
	stockField := dishooks.EmbedField{
		Name:   "STOCK STATUS:",
		Value:  productStock,
		Inline: true,
	}

	embed := dishooks.Embed{
		Fields: []*dishooks.EmbedField{
			&priceField,
			&stockField,
		},
		Author: &dishooks.EmbedAuthor{
			Name: "Ralph Lauren Mens Accessory",
			URL:  "https://www.ralphlauren.eu/dk/en/men/accessories/1030?srule=price-low-high&start=0&sz=32&webcat=Men%7CAccessories",
		},
		Thumbnail: &dishooks.EmbedThumbnail{
			URL: productImg,
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
		Username: "Ralph Lauren Mens Accessory - Under 200 DKK",
		Embeds:   []*dishooks.Embed{&embed},
	})
	if err != nil {
		log.Println()
	}
}
