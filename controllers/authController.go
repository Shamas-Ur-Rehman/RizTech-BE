package controllers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"supergit/inpatient/config"
	"supergit/inpatient/dtos"
	"supergit/inpatient/models"
	"supergit/inpatient/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Register — public, assigns "user" role automatically
func Register(c *gin.Context, db *mongo.Database) {
	var req dtos.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Duplicate email check
	if n, _ := db.Collection(config.ColUsers).CountDocuments(ctx, bson.M{"email": req.Email}); n > 0 {
		c.JSON(http.StatusConflict, utils.ErrorResponse{Status: http.StatusConflict, Message: "Email already registered"})
		return
	}

	// Fetch the "user" role
	var role models.Role
	if err := db.Collection(config.ColRoles).FindOne(ctx, bson.M{"name": models.RoleUser}).Decode(&role); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Default role not found"})
		return
	}

	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to hash password"})
		return
	}

	now := time.Now()
	user := models.User{
		ID:        primitive.NewObjectID(),
		FullName:  req.FullName,
		Email:     req.Email,
		Password:  hashed,
		Contact:   req.Contact,
		RoleID:    role.ID,
		RoleName:  role.Name,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if _, err := db.Collection(config.ColUsers).InsertOne(ctx, user); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to create user"})
		return
	}

	token, _ := utils.GenerateJWT(user.ID.Hex(), user.RoleID.Hex(), user.Email, user.RoleName)

	c.JSON(http.StatusCreated, utils.SuccessResponse{
		Status:  http.StatusCreated,
		Message: "Registered successfully",
		Data:    buildAuthResponse(user, role, token, nil),
	})
}

// Login — works for both regular users and admins
func Login(c *gin.Context, db *mongo.Database) {
	var req dtos.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	if err := db.Collection(config.ColUsers).FindOne(ctx, bson.M{"email": req.Email}).Decode(&user); err != nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Status: http.StatusUnauthorized, Message: "Invalid email or password"})
		return
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Status: http.StatusUnauthorized, Message: "Invalid email or password"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, utils.ErrorResponse{Status: http.StatusForbidden, Message: "account_deactivated"})
		return
	}

	var role models.Role
	if err := db.Collection(config.ColRoles).FindOne(ctx, bson.M{"_id": user.RoleID}).Decode(&role); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to retrieve role"})
		return
	}

	// Fetch permissions grouped by module
	cursor, _ := db.Collection(config.ColPermissions).Find(ctx, bson.M{"role_id": user.RoleID})
	var perms []models.Permission
	if cursor != nil {
		cursor.All(ctx, &perms)
		cursor.Close(ctx)
	}
	modulePerms := make(map[string][]string)
	for _, p := range perms {
		modulePerms[p.Module] = append(modulePerms[p.Module], p.Action)
	}

	token, err := utils.GenerateJWT(user.ID.Hex(), user.RoleID.Hex(), user.Email, role.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Login successful",
		Data:    buildAuthResponse(user, role, token, modulePerms),
	})
}

// ForgotPassword — generates a reset token (send via email in production)
func ForgotPassword(c *gin.Context, db *mongo.Database) {
	var req dtos.ForgotPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	if err := db.Collection(config.ColUsers).FindOne(ctx, bson.M{"email": req.Email}).Decode(&user); err != nil {
		// Don't reveal whether email exists
		c.JSON(http.StatusOK, utils.SuccessResponse{Status: http.StatusOK, Message: "If this email exists, a reset token has been sent"})
		return
	}

	b := make([]byte, 32)
	rand.Read(b)
	resetToken := hex.EncodeToString(b)
	expiry := time.Now().Add(1 * time.Hour)

	_, err := db.Collection(config.ColUsers).UpdateOne(ctx,
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{
			"reset_token":        resetToken,
			"reset_token_expiry": expiry,
			"updated_at":         time.Now(),
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to generate reset token"})
		return
	}

	// TODO: send resetToken via email in production
	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Password reset token generated",
		Data: gin.H{
			"reset_token": resetToken, // remove in production
			"expires_at":  expiry,
		},
	})
}

// ResetPassword — validates token and sets new password
func ResetPassword(c *gin.Context, db *mongo.Database) {
	var req dtos.ResetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	if err := db.Collection(config.ColUsers).FindOne(ctx, bson.M{
		"reset_token":        req.Token,
		"reset_token_expiry": bson.M{"$gt": time.Now()},
	}).Decode(&user); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid or expired reset token"})
		return
	}

	hashed, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to hash password"})
		return
	}

	_, err = db.Collection(config.ColUsers).UpdateOne(ctx,
		bson.M{"_id": user.ID},
		bson.M{
			"$set":   bson.M{"password": hashed, "updated_at": time.Now()},
			"$unset": bson.M{"reset_token": "", "reset_token_expiry": ""},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to reset password"})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{Status: http.StatusOK, Message: "Password reset successfully"})
}

// GetMe — returns current authenticated user's profile
func GetMe(c *gin.Context, db *mongo.Database) {
	userID, _ := c.Get("user_id")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: "Invalid user ID"})
		return
	}

	var user models.User
	if err := db.Collection(config.ColUsers).FindOne(ctx, bson.M{"_id": oid}).Decode(&user); err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{Status: http.StatusNotFound, Message: "User not found"})
		return
	}

	var role models.Role
	db.Collection(config.ColRoles).FindOne(ctx, bson.M{"_id": user.RoleID}).Decode(&role)

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Status:  http.StatusOK,
		Message: "Profile retrieved",
		Data: gin.H{
			"id":       user.ID.Hex(),
			"email":    user.Email,
			"fullName": user.FullName,
			"contact":  user.Contact,
			"isActive": user.IsActive,
			"roleId":   user.RoleID.Hex(),
			"roleName": role.Name,
			"isAdmin":  role.Name == models.RoleAdmin,
		},
	})
}

// buildAuthResponse constructs the standard login/register response
func buildAuthResponse(user models.User, role models.Role, token string, permissions map[string][]string) gin.H {
	return gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID.Hex(),
			"email":    user.Email,
			"fullName": user.FullName,
			"contact":  user.Contact,
			"isActive": user.IsActive,
			"roleId":   user.RoleID.Hex(),
			"roleName": role.Name,
			"isAdmin":  role.Name == models.RoleAdmin,
		},
		"permissions": permissions,
	}
}

// ChangePassword — authenticated user changes their own password
func ChangePassword(c *gin.Context, db *mongo.Database) {
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Status: http.StatusBadRequest, Message: err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oid, _ := primitive.ObjectIDFromHex(userID.(string))
	var user models.User
	if err := db.Collection(config.ColUsers).FindOne(ctx, bson.M{"_id": oid}).Decode(&user); err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse{Status: http.StatusNotFound, Message: "User not found"})
		return
	}

	if !utils.CheckPassword(req.CurrentPassword, user.Password) {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Status: http.StatusUnauthorized, Message: "Current password is incorrect"})
		return
	}

	hashed, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to hash password"})
		return
	}

	_, err = db.Collection(config.ColUsers).UpdateOne(ctx,
		bson.M{"_id": oid},
		bson.M{"$set": bson.M{"password": hashed, "updated_at": time.Now()}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Status: http.StatusInternalServerError, Message: "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{Status: http.StatusOK, Message: "Password changed successfully"})
}
