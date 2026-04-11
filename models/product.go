package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Spec struct {
	Label string `bson:"label" json:"label"`
	Value string `bson:"value" json:"value"`
}

type Product struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"        json:"id"`
	Slug          string             `bson:"slug"                 json:"slug"`
	Name          string             `bson:"name"                 json:"name"`
	Price         float64            `bson:"price"                json:"price"`
	OriginalPrice float64            `bson:"original_price"       json:"originalPrice"`
	Badge         string             `bson:"badge,omitempty"      json:"badge,omitempty"`
	Image         string             `bson:"image"                json:"image"`
	Gallery       []string           `bson:"gallery"              json:"gallery"`
	Description   string             `bson:"description"          json:"description"`
	Specs         []Spec             `bson:"specs"                json:"specs"`
	Features      []string           `bson:"features"             json:"features"`
	IsActive      bool               `bson:"is_active"            json:"isActive"`
	CreatedAt     time.Time          `bson:"created_at"           json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updated_at"           json:"updatedAt"`
}
