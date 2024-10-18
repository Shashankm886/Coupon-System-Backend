package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Shashankm886/coupon_system_backend/models"
	"github.com/captaincodeman/couponcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CouponService interface {
	Create(models.CreateCoupon, *mongo.Collection, *mongo.Collection) (models.CreateCoupon, error)
	FindAll(*mongo.Collection) []models.ListCoupon
	RedeemCoupon(models.RedeemCouponRequest, *mongo.Collection, *mongo.Collection, *mongo.Collection) (bool, error)
}

type couponService struct {
}

func New() CouponService {
	return &couponService{}
}

func (service *couponService) Create(coupon models.CreateCoupon, couponsCollection *mongo.Collection, userCollection *mongo.Collection) (models.CreateCoupon, error) {
	if coupon.DiscountPercent == 0 || coupon.Usage == 0 || coupon.ExpiryDate.IsZero() {
		log.Println("Error: Missing required fields - discount_percent, usage, or expiry_date")
		return models.CreateCoupon{}, errors.New("missing required fields")
	}

	if coupon.ProfileInfo != nil {
		// Create a context for querying MongoDB
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Query the users collection to check if the username exists
		filter := bson.M{"username": coupon.ProfileInfo.Username}
		var user bson.M
		err := userCollection.FindOne(ctx, filter).Decode(&user)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				// Username does not exist, return an error
				return models.CreateCoupon{}, errors.New("username does not exist")
			}
			// Some other error occurred during the query
			log.Println("Error querying user collection:", err)
			return models.CreateCoupon{}, err
		}
	}

	// Generate a coupon code
	code := couponcode.Generate()
	coupon.CouponCode = code
	log.Println("Generated Coupon Code:", coupon.CouponCode)

	// Insert the created coupon into MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := couponsCollection.InsertOne(ctx, coupon)
	if err != nil {
		log.Println("Error inserting new coupon:", err)
		return models.CreateCoupon{}, err
	}

	return coupon, nil
}

func (service *couponService) FindAll(couponsCollection *mongo.Collection) []models.ListCoupon {
	// Create a context with a timeout for querying MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Define an empty slice of ListCoupon to store the results
	var coupons []models.ListCoupon

	// Find all documents in the collection
	cursor, err := couponsCollection.Find(ctx, bson.D{}, options.Find())
	if err != nil {
		log.Println("Failed to fetch coupons from MongoDB:", err)
		return coupons
	}
	defer cursor.Close(ctx)

	// Iterate through the cursor and decode each document into the coupons slice
	for cursor.Next(ctx) {
		var coupon models.ListCoupon
		if err := cursor.Decode(&coupon); err != nil {
			log.Println("Error decoding coupon:", err)
			continue
		}
		coupons = append(coupons, coupon)
	}

	// Check if there were any cursor errors during the iteration
	if err := cursor.Err(); err != nil {
		log.Println("Cursor error:", err)
	}

	return coupons
}

func (service *couponService) RedeemCoupon(request models.RedeemCouponRequest, couponsCollection *mongo.Collection, userCollection *mongo.Collection, ordersCollection *mongo.Collection) (bool, error) {
	// Create a context for MongoDB operations
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find the coupon by coupon code
	var coupon models.ListCoupon
	filter := bson.M{"coupon_code": request.CouponCode}
	err := couponsCollection.FindOne(ctx, filter).Decode(&coupon)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, fmt.Errorf("coupon not found")
		}
		return false, err
	}

	// Check if coupon has expired
	currentTime := time.Now()
	if currentTime.After(coupon.ExpiryDate) {
		return false, fmt.Errorf("coupon has expired")
	}

	// Check if coupon has remaining usage
	if coupon.Usage <= 0 {
		return false, fmt.Errorf("coupon usage exhausted")
	}

	// Validate order content if present
	if coupon.OrderContent != nil {
		if request.OrderAmount < float64(coupon.OrderContent.MinAmount) || request.NumItems < coupon.OrderContent.NumberOfItems {
			return false, fmt.Errorf("order content does not meet minimum requirements")
		}
	}

	// Validate profile info if present
	if coupon.ProfileInfo != nil {
		if coupon.ProfileInfo.Username != request.Username {
			return false, fmt.Errorf("username does not match profile info")
		}
	}
	// Order History verification
	if coupon.OrderHistory != nil {
		// Create a filter for the orders collection
		orderFilter := bson.M{
			"username":  request.Username,
			"date":      bson.M{"$gte": coupon.OrderHistory.CheckTillDate}, // Date filter
			"is_coupon": false,                                             // Only non-coupon orders
		}

		// Count matching orders
		orderCount, err := ordersCollection.CountDocuments(ctx, orderFilter)
		if err != nil {
			return false, fmt.Errorf("error fetching orders, coupon could not be redeemed currently")
		}

		// Check if the order count exceeds the minimum required
		if orderCount < int64(coupon.OrderHistory.MinOrdersWithCoupon) {
			return false, fmt.Errorf("user has reach usage limit for this coupon")
		}
	}

	// Decrement the coupon usage
	coupon.Usage -= 1

	// Update the coupon in the database
	_, updateErr := couponsCollection.UpdateOne(
		ctx,
		filter,
		bson.M{"$set": bson.M{"usage": coupon.Usage}},
	)
	if updateErr != nil {
		return false, updateErr
	}

	return true, nil
}
