package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type WebhookSettings struct {
	ID           primitive.ObjectID `bson:"_id"`
	Server       string             `json: "server", bson:"server"`
	Store        string             `json: "store", bson:"store"`
	Subsection   string             `json: "subsection", bson:"subsection"`
	MonitorSites []string           `json: "monitorsites", bson:"monitorsites"`
	Webhooks     []string           `json: "webhooks", bson:"webhooks"`
}

type MonitorSettings struct {
	ID    primitive.ObjectID `bson:"_id"`
	Store string             `json: "store", bson:"store"`
	PIDS  []string           `json: "pids", bson:"pids"`
}

type UnisportMonitorStruct struct {
	Pid   string `json:"pid"`
	Stock []struct {
		Size  string `json:"size"`
		Stock string `json:"stock"`
	} `json:"stock"`
}

type RalphPreviousItemsStruct struct {
	ID         primitive.ObjectID `bson:"_id"`
	Subsection string             `json:"subsection"`
	Items      []string           `json:"items"`
}

type ShopifyMonitorStruct struct {
	Site  string `json:"site"`
	Stock []struct {
		Product struct {
			Title     string `json:"title"`
			Available bool   `json:"available"`
		} `json:"product"`
	} `json:"stock"`
}
