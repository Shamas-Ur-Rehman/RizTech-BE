package controllers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
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

type PlaceOrderReq struct {
	Items           []OrderItemReq         `json:"items"           binding:"required,min=1"`
	ShippingAddress models.ShippingAddress `json:"shippingAddress" binding:"required"`
	PaymentMethod   string                 `json:"paymentMethod"   binding:"required"`
}

type OrderItemReq struct {
	ProductID string `json:"productId" binding:"required"`
	Quantity  int    `json:"quantity"  binding:"required,min=1"`
}

// PlaceOrder — authenticated users
func PlaceOrder(c *gin.Context, db *mongo.Database) {
	var req PlaceOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	userEmail, _ := c.Get("email")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uid, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid user"})
		return
	}

	var user models.User
	db.Collection(config.ColUsers).FindOne(ctx, bson.M{"_id": uid}).Decode(&user)

	var orderItems []models.OrderItem
	var total float64

	for _, item := range req.Items {
		pid, err := primitive.ObjectIDFromHex(item.ProductID)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid product ID: " + item.ProductID})
			return
		}
		var product models.Product
		if err := db.Collection(config.ColProducts).FindOne(ctx, bson.M{"_id": pid, "is_active": true}).Decode(&product); err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Product not found: " + item.ProductID})
			return
		}
		orderItems = append(orderItems, models.OrderItem{
			ProductID:   pid,
			ProductName: product.Name,
			Image:       product.Image,
			Price:       product.Price,
			Quantity:    item.Quantity,
		})
		total += product.Price * float64(item.Quantity)
	}

	now := time.Now()
	order := models.Order{
		ID:              primitive.NewObjectID(),
		UserID:          uid,
		UserEmail:       userEmail.(string),
		UserName:        user.FullName,
		Items:           orderItems,
		ShippingAddress: req.ShippingAddress,
		PaymentMethod:   req.PaymentMethod,
		TotalPrice:      total,
		Status:          "pending",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if _, err := db.Collection(config.ColOrders).InsertOne(ctx, order); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to place order"})
		return
	}

	c.JSON(http.StatusCreated, utils.SuccessResponse{Status: http.StatusCreated, Message: "Order placed successfully", Data: order})
}

// ListOrders — admin, supports ?status=&search=&days=&page=&limit=
func ListOrders(c *gin.Context, db *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}

	// Status filter
	if status := c.Query("status"); status != "" && status != "all" {
		filter["status"] = status
	}

	// Date range — default 30 days
	daysStr := c.Query("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}
	if daysStr != "all" {
		filter["created_at"] = bson.M{"$gte": time.Now().AddDate(0, 0, -days)}
	}

	// Search by product name or customer name/email
	if search := strings.TrimSpace(c.Query("search")); search != "" {
		regex := bson.M{"$regex": search, "$options": "i"}
		filter["$or"] = bson.A{
			bson.M{"user_name": regex},
			bson.M{"user_email": regex},
			bson.M{"items.product_name": regex},
		}
	}

	// Pagination
	page := 1
	limit := 20
	if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	skip := int64((page - 1) * limit)

	total, _ := db.Collection(config.ColOrders).CountDocuments(ctx, filter)

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(skip).
		SetLimit(int64(limit))

	cursor, err := db.Collection(config.ColOrders).Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to fetch orders"})
		return
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	cursor.All(ctx, &orders)
	if orders == nil {
		orders = []models.Order{}
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Orders retrieved",
		Data: gin.H{
			"orders":     orders,
			"total":      total,
			"page":       page,
			"limit":      limit,
			"totalPages": (int(total) + limit - 1) / limit,
		},
	})
}

// UpdateOrderStatus — admin only
func UpdateOrderStatus(c *gin.Context, db *mongo.Database) {
	id := c.Param("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid order ID"})
		return
	}

	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := db.Collection(config.ColOrders).UpdateOne(ctx, bson.M{"_id": oid},
		bson.M{"$set": bson.M{"status": body.Status, "updated_at": time.Now()}})
	if err != nil || res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{Status: http.StatusNotFound, Message: "Order not found"})
		return
	}

	var updated models.Order
	db.Collection(config.ColOrders).FindOne(ctx, bson.M{"_id": oid}).Decode(&updated)
	c.JSON(http.StatusOK, utils.SuccessResponse{Status: http.StatusOK, Message: "Order status updated", Data: updated})
}

// AddTracking — admin only, adds/updates tracking info on an order
func AddTracking(c *gin.Context, db *mongo.Database) {
	id := c.Param("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid order ID"})
		return
	}

	var body struct {
		TrackingNumber  string `json:"trackingNumber"  binding:"required"`
		ShippingCompany string `json:"shippingCompany" binding:"required"`
		TrackingURL     string `json:"trackingUrl"`
		Note            string `json:"note"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tracking := models.TrackingInfo{
		TrackingNumber:  body.TrackingNumber,
		ShippingCompany: body.ShippingCompany,
		TrackingURL:     body.TrackingURL,
		Note:            body.Note,
		AddedAt:         time.Now(),
	}

	res, err := db.Collection(config.ColOrders).UpdateOne(ctx, bson.M{"_id": oid},
		bson.M{"$set": bson.M{
			"tracking":   tracking,
			"updated_at": time.Now(),
		}})
	if err != nil || res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{Status: http.StatusNotFound, Message: "Order not found"})
		return
	}

	var updated models.Order
	db.Collection(config.ColOrders).FindOne(ctx, bson.M{"_id": oid}).Decode(&updated)
	c.JSON(http.StatusOK, utils.SuccessResponse{Status: http.StatusOK, Message: "Tracking added", Data: updated})
}

// GetMyOrders — authenticated user sees their own orders
func GetMyOrders(c *gin.Context, db *mongo.Database) {
	userID, _ := c.Get("user_id")
	uid, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid user"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := db.Collection(config.ColOrders).Find(ctx, bson.M{"user_id": uid}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to fetch orders"})
		return
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	cursor.All(ctx, &orders)
	if orders == nil {
		orders = []models.Order{}
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Status: http.StatusOK, Message: "Orders retrieved", Data: orders})
}
