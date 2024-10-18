package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Shashankm886/coupon_system_backend/controller"
	"github.com/Shashankm886/coupon_system_backend/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	couponService    service.CouponService       = service.New()
	couponController controller.CouponController = controller.New(couponService)
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client, err := connectMongoDB()
	if err != nil {
		log.Fatal("Error connecting to MongoDB: ", err)
	}

	// Ensure that the MongoDB connection is closed properly when the application stops
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()

	swishDatabase := client.Database("Swish")
	couponsCollection := swishDatabase.Collection("coupons")
	userCollection := swishDatabase.Collection("user")
	ordersCollection := swishDatabase.Collection("orders")

	server := gin.Default()

	server.GET("/coupons", func(ctx *gin.Context) {
		ctx.JSON(200, couponController.FindAll(couponsCollection))
	})

	// POST /coupons - Create a new coupon
	server.POST("/coupons", func(ctx *gin.Context) {
		couponController.Create(ctx, couponsCollection, userCollection)
	})

	server.POST("/redeem", func(ctx *gin.Context) {
		var is_redeemed, err = couponController.Redeem(ctx, couponsCollection, userCollection, ordersCollection)
		if is_redeemed {
			ctx.JSON(200, gin.H{"redeem_status": true})
		} else {
			ctx.JSON(400, gin.H{"error": err.Error()})
		}
	})

	server.Run(":8080")
}

func connectMongoDB() (*mongo.Client, error) {
	// Replace <username>, <password>, and <cluster-url> with your own MongoDB credentials
	uri := os.Getenv("MONGODB_URI")

	// Create a new MongoDB client with the connection URI
	clientOptions := options.Client().ApplyURI(uri)

	// Create a new client and connect to the cluster
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	// Set a timeout for the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ping the MongoDB server to ensure connection is established
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	log.Println("Successfully connected to MongoDB!")
	return client, nil
}
