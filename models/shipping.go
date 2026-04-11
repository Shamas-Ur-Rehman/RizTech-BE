package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CountryShipping defines per-country shipping behaviour
// Mode: "charge" = allow with fee | "whatsapp" = redirect to WhatsApp | "free" = free shipping
type CountryShipping struct {
	Country  string  `bson:"country"  json:"country"`  // full country name or ISO code
	Mode     string  `bson:"mode"     json:"mode"`     // "free" | "charge" | "whatsapp"
	Charge   float64 `bson:"charge"   json:"charge"`   // used when mode == "charge"
	Note     string  `bson:"note"     json:"note"`     // optional note shown to customer
}

// ShippingSettings is a singleton document (only one record in the collection)
type ShippingSettings struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"      json:"id"`
	DefaultMode     string             `bson:"default_mode"       json:"defaultMode"`    // "free" | "charge" | "whatsapp"
	DefaultCharge   float64            `bson:"default_charge"     json:"defaultCharge"`  // used when defaultMode == "charge"
	WhatsappNumber  string             `bson:"whatsapp_number"    json:"whatsappNumber"` // e.g. 923034816023
	Countries       []CountryShipping  `bson:"countries"          json:"countries"`
	UpdatedAt       time.Time          `bson:"updated_at"         json:"updatedAt"`
}
