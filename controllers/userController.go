package controllers

import (
	"context"
	"net/http"
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

// ListUsers — admin only, returns all non-admin users with optional search
func ListUsers(c *gin.Context, db *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"role_name": models.RoleUser}

	if search := strings.TrimSpace(c.Query("search")); search != "" {
		regex := bson.M{"$regex": search, "$options": "i"}
		filter["$or"] = bson.A{
			bson.M{"full_name": regex},
			bson.M{"email": regex},
		}
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := db.Collection(config.ColUsers).Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to fetch users"})
		return
	}
	defer cursor.Close(ctx)

	var users []models.User
	cursor.All(ctx, &users)
	if users == nil {
		users = []models.User{}
	}

	// Strip passwords before returning
	type SafeUser struct {
		ID        primitive.ObjectID `json:"id"`
		FullName  string             `json:"fullName"`
		Email     string             `json:"email"`
		Contact   string             `json:"contact"`
		RoleName  string             `json:"roleName"`
		IsActive  bool               `json:"isActive"`
		CreatedAt time.Time          `json:"createdAt"`
	}
	safe := make([]SafeUser, len(users))
	for i, u := range users {
		safe[i] = SafeUser{
			ID:        u.ID,
			FullName:  u.FullName,
			Email:     u.Email,
			Contact:   u.Contact,
			RoleName:  u.RoleName,
			IsActive:  u.IsActive,
			CreatedAt: u.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{Status: http.StatusOK, Message: "Users retrieved", Data: safe})
}

// ToggleUserStatus — admin only, activates or deactivates a user
func ToggleUserStatus(c *gin.Context, db *mongo.Database) {
	id := c.Param("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid user ID"})
		return
	}

	var body struct {
		IsActive bool `json:"isActive"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := db.Collection(config.ColUsers).UpdateOne(ctx,
		bson.M{"_id": oid, "role_name": models.RoleUser}, // prevent deactivating admins
		bson.M{"$set": bson.M{"is_active": body.IsActive, "updated_at": time.Now()}},
	)
	if err != nil || res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{Status: http.StatusNotFound, Message: "User not found"})
		return
	}

	status := "deactivated"
	if body.IsActive {
		status = "activated"
	}
	c.JSON(http.StatusOK, utils.SuccessResponse{Status: http.StatusOK, Message: "User " + status})
}
