package domain

import "gopkg.in/mgo.v2/bson"

type Poll struct {
	ID      bson.ObjectId  `bson: "_id" json:"id"`
	Title   string         `json: "title"`
	Options []string       `json: "options""`
	Results map[string]int `json: "results,omitempty"`
	APIKey  string         `json: "apikey"`
}
