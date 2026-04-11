package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderItem struct {
	ProductID   primitive.ObjectID `bson:"product_id"   json:"productId"`
	ProductName string             `bson:"product_name" json:"productName"`
	Image       string             `bson:"image"        json:"image"`
	Price       float64            `bson:"price"        json:"price"`
	Quantity    int                `bson:"quantity"     json:"quantity"`
}

type ShippingAddress struct {
	Name    string `bson:"name"     json:"name"`
	Email   string `bson:"email"    json:"email"`
	Phone   string `bson:"phone"    json:"phone"`
	Address string `bson:"address"  json:"address"`
	City    string `bson:"city"     json:"city"`
	Country string `bson:"country"  json:"country"`
	ZipCode string `bson:"zip_code" json:"zipCode"`
}

type TrackingInfo struct {
	TrackingNumber  string    `bson:"tracking_number"  json:"trackingNumber"`
	ShippingCompany string    `bson:"shipping_company" json:"shippingCompany"`
	TrackingURL     string    `bson:"tracking_url"     json:"trackingUrl"`
	Note            string    `bson:"note"             json:"note"`
	AddedAt         time.Time `bson:"added_at"         json:"addedAt"`
}

// Status: pending | processing | shipped | delivered | cancelled
type Order struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"     json:"id"`
	UserID          primitive.ObjectID `bson:"user_id"           json:"userId"`
	UserEmail       string             `bson:"user_email"        json:"userEmail"`
	UserName        string             `bson:"user_name"         json:"userName"`
	Items           []OrderItem        `bson:"items"             json:"items"`
	ShippingAddress ShippingAddress    `bson:"shipping_address"  json:"shippingAddress"`
	PaymentMethod   string             `bson:"payment_method"    json:"paymentMethod"`
	TotalPrice      float64            `bson:"total_price"       json:"totalPrice"`
	Status          string             `bson:"status"            json:"status"`
	Tracking        *TrackingInfo      `bson:"tracking,omitempty" json:"tracking,omitempty"`
	CreatedAt       time.Time          `bson:"created_at"        json:"createdAt"`
	UpdatedAt       time.Time          `bson:"updated_at"        json:"updatedAt"`
}
