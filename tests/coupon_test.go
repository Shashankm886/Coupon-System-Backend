package tests

import (
	"context"
	"testing"
	"time"

	"github.com/Shashankm886/coupon_system_backend/models"
	"github.com/Shashankm886/coupon_system_backend/service"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Mock MongoDB connection for testing
func setupTestDB() (*mongo.Client, *mongo.Collection, *mongo.Collection, *mongo.Collection) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ := mongo.Connect(context.TODO(), clientOptions)
	db := client.Database("test_swish")
	return client, db.Collection("coupons"), db.Collection("user"), db.Collection("orders")
}

func TestCreateCoupon(t *testing.T) {
	client, couponsCollection, userCollection, _ := setupTestDB()
	defer client.Disconnect(context.TODO())

	couponService := service.New()

	// Mocking data
	user := models.User{
		ID:       primitive.NewObjectID(),
		Username: "chaman",
	}
	userCollection.InsertOne(context.TODO(), user)

	coupon := models.CreateCoupon{
		DiscountPercent: 10,
		Usage:           5,
		ExpiryDate:      time.Now().AddDate(0, 1, 0), // valid for 1 month
		ProfileInfo: &models.ProfileInfo{
			Username: "chaman",
			Frequent: true,
		},
	}

	createdCoupon, err := couponService.Create(coupon, couponsCollection, userCollection) // Handle error
	assert.NoError(t, err, "Expected no error while creating coupon")                     // Assert no error
	assert.NotNil(t, createdCoupon.CouponCode, "Coupon creation failed, coupon code is empty")
	assert.Equal(t, "chaman", createdCoupon.ProfileInfo.Username, "User validation failed")

	// Clean up
	userCollection.DeleteOne(context.TODO(), bson.M{"username": "chaman"})
	couponsCollection.DeleteOne(context.TODO(), bson.M{"coupon_code": createdCoupon.CouponCode})
}

func TestRedeemCoupon(t *testing.T) {
	client, couponsCollection, userCollection, ordersCollection := setupTestDB()
	defer client.Disconnect(context.TODO())

	couponService := service.New()

	// Insert a mock user, coupon, and order data
	user := models.User{
		ID:       primitive.NewObjectID(),
		Username: "john_doe",
	}
	userCollection.InsertOne(context.TODO(), user)

	coupon := models.ListCoupon{
		CouponCode: "FLAT10",
		Usage:      3,
		ExpiryDate: time.Now().AddDate(0, 1, 0), // valid for 1 month
		ProfileInfo: &models.ProfileInfo{
			Username: "john_doe",
			Frequent: true,
		},
		OrderContent: &models.OrderContent{
			MinAmount:     50,
			NumberOfItems: 2,
		},
	}
	couponsCollection.InsertOne(context.TODO(), coupon)

	request := models.RedeemCouponRequest{
		CouponCode:  "FLAT10",
		OrderAmount: 100,
		NumItems:    3,
		Username:    "john_doe",
	}

	success, err := couponService.RedeemCoupon(request, couponsCollection, userCollection, ordersCollection)
	assert.NoError(t, err, "Unexpected error during coupon redemption")
	assert.True(t, success, "Coupon redemption should succeed")

	// Clean up
	userCollection.DeleteOne(context.TODO(), bson.M{"username": "john_doe"})
	couponsCollection.DeleteOne(context.TODO(), bson.M{"coupon_code": "FLAT10"})
}
