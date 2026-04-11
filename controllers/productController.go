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

// ListProducts — public, returns all active products
func ListProducts(c *gin.Context, db *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := db.Collection(config.ColProducts).Find(ctx, bson.M{"is_active": true}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to fetch products"})
		return
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err := cursor.All(ctx, &products); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to decode products"})
		return
	}
	if products == nil {
		products = []models.Product{}
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Status: http.StatusOK, Message: "Products retrieved", Data: products})
}

// GetProduct — public, single product by id
func GetProduct(c *gin.Context, db *mongo.Database) {
	id := c.Param("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid product ID"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var product models.Product
	if err := db.Collection(config.ColProducts).FindOne(ctx, bson.M{"_id": oid, "is_active": true}).Decode(&product); err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{Status: http.StatusNotFound, Message: "Product not found"})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Status: http.StatusOK, Message: "Product retrieved", Data: product})
}

// CreateProduct — admin only
func CreateProduct(c *gin.Context, db *mongo.Database) {
	var p models.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	p.ID = primitive.NewObjectID()
	p.IsActive = true
	p.CreatedAt = now
	p.UpdatedAt = now
	if p.Gallery == nil {
		p.Gallery = []string{}
	}
	if p.Specs == nil {
		p.Specs = []models.Spec{}
	}
	if p.Features == nil {
		p.Features = []string{}
	}

	if _, err := db.Collection(config.ColProducts).InsertOne(ctx, p); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to create product"})
		return
	}
	c.JSON(http.StatusCreated, utils.SuccessResponse{Status: http.StatusCreated, Message: "Product created", Data: p})
}

// UpdateProduct — admin only
func UpdateProduct(c *gin.Context, db *mongo.Database) {
	id := c.Param("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid product ID"})
		return
	}

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}
	body["updated_at"] = time.Now()
	delete(body, "_id")
	delete(body, "id")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := db.Collection(config.ColProducts).UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": body})
	if err != nil || res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{Status: http.StatusNotFound, Message: "Product not found"})
		return
	}

	var updated models.Product
	db.Collection(config.ColProducts).FindOne(ctx, bson.M{"_id": oid}).Decode(&updated)
	c.JSON(http.StatusOK, utils.SuccessResponse{Status: http.StatusOK, Message: "Product updated", Data: updated})
}

// DeleteProduct — admin only (soft delete)
func DeleteProduct(c *gin.Context, db *mongo.Database) {
	id := c.Param("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid product ID"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := db.Collection(config.ColProducts).UpdateOne(ctx, bson.M{"_id": oid},
		bson.M{"$set": bson.M{"is_active": false, "updated_at": time.Now()}})
	if err != nil || res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{Status: http.StatusNotFound, Message: "Product not found"})
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Status: http.StatusOK, Message: "Product deleted"})
}
