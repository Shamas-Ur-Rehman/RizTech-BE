package controllers

import (
	"context"
	"net/http"
	"time"

	"supergit/inpatient/config"
	"supergit/inpatient/models"
	"supergit/inpatient/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetShippingSettings — public, used by checkout to determine charges
func GetShippingSettings(c *gin.Context, db *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var settings models.ShippingSettings
	err := db.Collection(config.ColShippingSettings).FindOne(ctx, bson.M{}).Decode(&settings)
	if err == mongo.ErrNoDocuments {
		// Return sensible defaults if never configured
		c.JSON(http.StatusOK, utils.SuccessResponse{
			Status:  http.StatusOK,
			Message: "Shipping settings",
			Data: models.ShippingSettings{
				DefaultMode:    "free",
				DefaultCharge:  0,
				WhatsappNumber: "",
				Countries:      []models.CountryShipping{},
			},
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to fetch shipping settings"})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Status: http.StatusOK, Message: "Shipping settings", Data: settings})
}

// SaveShippingSettings — admin only, upserts the singleton document
func SaveShippingSettings(c *gin.Context, db *mongo.Database) {
	var body models.ShippingSettings
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	body.UpdatedAt = time.Now()
	if body.Countries == nil {
		body.Countries = []models.CountryShipping{}
	}

	// Upsert: update the single document or insert if none exists
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	update := bson.M{
		"$set": bson.M{
			"default_mode":    body.DefaultMode,
			"default_charge":  body.DefaultCharge,
			"whatsapp_number": body.WhatsappNumber,
			"countries":       body.Countries,
			"updated_at":      body.UpdatedAt,
		},
		"$setOnInsert": bson.M{"_id": primitive.NewObjectID()},
	}

	var result models.ShippingSettings
	err := db.Collection(config.ColShippingSettings).FindOneAndUpdate(ctx, bson.M{}, update, opts).Decode(&result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to save shipping settings"})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{Status: http.StatusOK, Message: "Shipping settings saved", Data: result})
}
