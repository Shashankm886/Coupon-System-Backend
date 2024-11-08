package controller

import (
	"net/http"

	"github.com/Shashankm886/coupon_system_backend/models"
	"github.com/Shashankm886/coupon_system_backend/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type CouponController interface {
	FindAll(couponsCollection *mongo.Collection) []models.ListCoupon
	Create(ctx *gin.Context, couponsCollection *mongo.Collection, userCollection *mongo.Collection)
	Redeem(ctx *gin.Context, couponsCollection *mongo.Collection, userCollection *mongo.Collection, ordersCollection *mongo.Collection) (bool, error)
}

type controller struct {
	service service.CouponService
}

func New(service service.CouponService) CouponController {
	return &controller{
		service: service,
	}
}

func (c *controller) FindAll(couponsCollection *mongo.Collection) []models.ListCoupon {
	return c.service.FindAll(couponsCollection)
}
func (c *controller) Create(ctx *gin.Context, couponsCollection *mongo.Collection, userCollection *mongo.Collection) {
	var coupon models.CreateCoupon
	if err := ctx.BindJSON(&coupon); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	createdCoupon, err := c.service.Create(coupon, couponsCollection, userCollection)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, createdCoupon)
}

func (c *controller) Redeem(ctx *gin.Context, couponsCollection, userCollection *mongo.Collection, ordersCollection *mongo.Collection) (bool, error) {
	var request models.RedeemCouponRequest

	if err := ctx.BindJSON(&request); err != nil {
		return false, err
	}

	// Call the redeem function from the service
	status, err := c.service.RedeemCoupon(request, couponsCollection, userCollection, ordersCollection)
	if err != nil {
		return false, err
	}

	// Return the status response
	if status {
		return true, nil
	} else {
		return false, nil
	}
}
