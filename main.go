package main

import (
	"log"

	"github.com/Shashankm886/coupon_system_backend/controller"
	"github.com/Shashankm886/coupon_system_backend/database"
	"github.com/Shashankm886/coupon_system_backend/routes"
	"github.com/Shashankm886/coupon_system_backend/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
}

var (
	couponService    = service.New()
	couponController = controller.New(couponService)
)

func main() {
	// Connect to MongoDB
	client, err := database.ConnectMongoDB()
	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}
	defer database.DisconnectMongoDB(client)

	// Initialize database collections
	db := client.Database("Swish")
	couponsCollection := db.Collection("coupons")
	userCollection := db.Collection("user")
	ordersCollection := db.Collection("orders")

	// Set up Gin server and initialize routes
	server := gin.Default()
	routes.InitializeRoutes(server, couponController, couponsCollection, userCollection, ordersCollection)

	// Start server
	if err := server.Run(":8080"); err != nil {
		log.Fatal("Failed to run server:", err)
	}
}
