package utils

import (
	"context"
	"time"

	"supergit/inpatient/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetUserByID(db *mongo.Database, userID string) (*models.User, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var user models.User
	if err := db.Collection("users").FindOne(ctx, bson.M{"_id": oid, "is_active": true}).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByEmail(db *mongo.Database, email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var user models.User
	if err := db.Collection("users").FindOne(ctx, bson.M{"email": email, "is_active": true}).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}
