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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type StatsResponse struct {
	TotalOrders     int64              `json:"totalOrders"`
	TotalRevenue    float64            `json:"totalRevenue"`
	TotalProducts   int64              `json:"totalProducts"`
	TotalUsers      int64              `json:"totalUsers"`
	PendingOrders   int64              `json:"pendingOrders"`
	DeliveredOrders int64              `json:"deliveredOrders"`
	CancelledOrders int64              `json:"cancelledOrders"`
	RecentOrders    []models.Order     `json:"recentOrders"`
	TopProducts     []TopProduct       `json:"topProducts"`
	RevenueByDay    []RevenueDay       `json:"revenueByDay"`
}

type TopProduct struct {
	ProductName string  `json:"productName"`
	TotalSold   int     `json:"totalSold"`
	Revenue     float64 `json:"revenue"`
}

type RevenueDay struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
}

func GetStats(c *gin.Context, db *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Counts
	totalOrders, _ := db.Collection(config.ColOrders).CountDocuments(ctx, bson.M{})
	totalProducts, _ := db.Collection(config.ColProducts).CountDocuments(ctx, bson.M{"is_active": true})
	totalUsers, _ := db.Collection(config.ColUsers).CountDocuments(ctx, bson.M{"role_name": "user"})
	pendingOrders, _ := db.Collection(config.ColOrders).CountDocuments(ctx, bson.M{"status": "pending"})
	deliveredOrders, _ := db.Collection(config.ColOrders).CountDocuments(ctx, bson.M{"status": "delivered"})
	cancelledOrders, _ := db.Collection(config.ColOrders).CountDocuments(ctx, bson.M{"status": "cancelled"})

	// Total revenue (sum of all non-cancelled orders)
	revPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"status": bson.M{"$ne": "cancelled"}}}},
		{{Key: "$group", Value: bson.M{"_id": nil, "total": bson.M{"$sum": "$total_price"}}}},
	}
	revCursor, _ := db.Collection(config.ColOrders).Aggregate(ctx, revPipeline)
	var revResult []struct{ Total float64 `bson:"total"` }
	revCursor.All(ctx, &revResult)
	var totalRevenue float64
	if len(revResult) > 0 {
		totalRevenue = revResult[0].Total
	}

	// Recent 5 orders
	recentOpts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(5)
	recentCursor, _ := db.Collection(config.ColOrders).Find(ctx, bson.M{}, recentOpts)
	var recentOrders []models.Order
	recentCursor.All(ctx, &recentOrders)
	if recentOrders == nil {
		recentOrders = []models.Order{}
	}

	// Top 5 products by revenue
	topPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"status": bson.M{"$ne": "cancelled"}}}},
		{{Key: "$unwind", Value: "$items"}},
		{{Key: "$group", Value: bson.M{
			"_id":         "$items.product_name",
			"totalSold":   bson.M{"$sum": "$items.quantity"},
			"revenue":     bson.M{"$sum": bson.M{"$multiply": []interface{}{"$items.price", "$items.quantity"}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "revenue", Value: -1}}}},
		{{Key: "$limit", Value: 5}},
	}
	topCursor, _ := db.Collection(config.ColOrders).Aggregate(ctx, topPipeline)
	var topRaw []struct {
		ID        string  `bson:"_id"`
		TotalSold int     `bson:"totalSold"`
		Revenue   float64 `bson:"revenue"`
	}
	topCursor.All(ctx, &topRaw)
	topProducts := make([]TopProduct, 0, len(topRaw))
	for _, r := range topRaw {
		topProducts = append(topProducts, TopProduct{ProductName: r.ID, TotalSold: r.TotalSold, Revenue: r.Revenue})
	}

	// Revenue by day — last 7 days
	sevenDaysAgo := time.Now().AddDate(0, 0, -6)
	dayPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"status":     bson.M{"$ne": "cancelled"},
			"created_at": bson.M{"$gte": sevenDaysAgo},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$created_at"}},
			"revenue": bson.M{"$sum": "$total_price"},
			"orders":  bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
	}
	dayCursor, _ := db.Collection(config.ColOrders).Aggregate(ctx, dayPipeline)
	var dayRaw []struct {
		ID      string  `bson:"_id"`
		Revenue float64 `bson:"revenue"`
		Orders  int     `bson:"orders"`
	}
	dayCursor.All(ctx, &dayRaw)

	// Fill all 7 days even if no orders
	revenueByDay := make([]RevenueDay, 7)
	for i := 0; i < 7; i++ {
		d := time.Now().AddDate(0, 0, -(6 - i)).Format("2006-01-02")
		revenueByDay[i] = RevenueDay{Date: d, Revenue: 0, Orders: 0}
		for _, r := range dayRaw {
			if r.ID == d {
				revenueByDay[i].Revenue = r.Revenue
				revenueByDay[i].Orders = r.Orders
				break
			}
		}
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Stats retrieved",
		Data: StatsResponse{
			TotalOrders:     totalOrders,
			TotalRevenue:    totalRevenue,
			TotalProducts:   totalProducts,
			TotalUsers:      totalUsers,
			PendingOrders:   pendingOrders,
			DeliveredOrders: deliveredOrders,
			CancelledOrders: cancelledOrders,
			RecentOrders:    recentOrders,
			TopProducts:     topProducts,
			RevenueByDay:    revenueByDay,
		},
	})
}
