package config

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var MongoDB *mongo.Database

func ConnectMongoDB() *mongo.Client {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("MONGODB_URI is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("MongoDB connect error: %v", err)
	}
	if err = client.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB ping error: %v", err)
	}

	dbName := os.Getenv("MONGODB_DB")
	if dbName == "" {
		dbName = "app"
	}
	MongoDB = client.Database(dbName)
	log.Println("Connected to MongoDB:", dbName)

	ensureIndexes(ctx)
	seedRoles(ctx)

	return client
}

// ensureIndexes creates unique indexes on critical fields
func ensureIndexes(ctx context.Context) {
	_, _ = MongoDB.Collection(ColUsers).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	_, _ = MongoDB.Collection(ColRoles).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
}

// seedRoles ensures admin and user roles exist, and creates a default admin account
func seedRoles(ctx context.Context) {
	roles := []string{"admin", "user"}
	roleIDs := map[string]primitive.ObjectID{}

	for _, name := range roles {
		var existing struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		err := MongoDB.Collection(ColRoles).FindOne(ctx, bson.M{"name": name}).Decode(&existing)
		if err == mongo.ErrNoDocuments {
			id := primitive.NewObjectID()
			now := time.Now()
			_, _ = MongoDB.Collection(ColRoles).InsertOne(ctx, bson.M{
				"_id":        id,
				"name":       name,
				"created_at": now,
				"updated_at": now,
			})
			roleIDs[name] = id
			log.Printf("Seeded role: %s (%s)", name, id.Hex())
		} else if err == nil {
			roleIDs[name] = existing.ID
		}
	}

	// Seed default admin user if none exists
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminEmail == "" {
		adminEmail = "admin@app.com"
	}
	if adminPass == "" {
		adminPass = "Admin@1234"
	}

	count, _ := MongoDB.Collection(ColUsers).CountDocuments(ctx, bson.M{"email": adminEmail})
	if count == 0 {
		hashed, err := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
		if err == nil {
			now := time.Now()
			adminRoleID := roleIDs["admin"]
			_, _ = MongoDB.Collection(ColUsers).InsertOne(ctx, bson.M{
				"_id":        primitive.NewObjectID(),
				"full_name":  "Admin",
				"email":      adminEmail,
				"password":   string(hashed),
				"role_id":    adminRoleID,
				"role_name":  "admin",
				"is_active":  true,
				"created_at": now,
				"updated_at": now,
			})
			log.Printf("Default admin created: %s / %s", adminEmail, adminPass)
		}
	}
}
