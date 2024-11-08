package routes

import (
	"github.com/Shashankm886/coupon_system_backend/controller"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitializeRoutes(router *gin.Engine, couponController controller.CouponController,
	couponsCollection, userCollection, ordersCollection *mongo.Collection) {

	router.GET("/coupons", func(ctx *gin.Context) {
		ctx.JSON(200, couponController.FindAll(couponsCollection))
	})

	router.POST("/coupons", func(ctx *gin.Context) {
		couponController.Create(ctx, couponsCollection, userCollection)
	})

	router.POST("/redeem", func(ctx *gin.Context) {
		isRedeemed, err := couponController.Redeem(ctx, couponsCollection, userCollection, ordersCollection)
		if isRedeemed {
			ctx.JSON(200, gin.H{"redeem_status": true})
		} else {
			ctx.JSON(400, gin.H{"error": err.Error()})
		}
	})
}
