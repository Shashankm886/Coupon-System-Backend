package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderContent struct {
	MinAmount     int `json:"min_amount"`
	NumberOfItems int `json:"number_of_items"`
}

type ProfileInfo struct {
	Username string `json:"username"`
	Frequent bool   `json:"frequent"`
}

type OrderHistory struct {
	MinOrdersWithCoupon int       `json:"min_orders_with_coupon"`
	CheckTillDate       time.Time `json:"check_till_date"`
}

type CreateCoupon struct {
	CouponCode      string        `json:"coupon_code" bson:"coupon_code"`
	DiscountPercent float64       `json:"discount_percent" bson:"discount_percent" binding:"required"`
	Usage           int           `json:"usage" bson:"usage" binding:"required"`
	ExpiryDate      time.Time     `json:"expiry_date" bson:"expiry_date" binding:"required"`
	OrderContent    *OrderContent `json:"order_content" bson:"order_content"`
	ProfileInfo     *ProfileInfo  `json:"profile_info" bson:"profile_info"`
	OrderHistory    *OrderHistory `json:"order_history" bson:"order_history"`
}

type ListCoupon struct {
	ID              primitive.ObjectID `json:"_id" bson:"_id"`
	CouponCode      string             `json:"coupon_code" bson:"coupon_code"`
	DiscountPercent float64            `json:"discount_percent" bson:"discount_percent"`
	ExpiryDate      time.Time          `json:"expiry_date" bson:"expiry_date"`
	Usage           int                `json:"usage" bson:"usage"`
	OrderContent    *OrderContent      `json:"order_content" bson:"order_content"`
	ProfileInfo     *ProfileInfo       `json:"profile_info" bson:"profile_info"`
	OrderHistory    *OrderHistory      `json:"order_history" bson:"order_history"`
}

type RedeemCouponRequest struct {
	CouponCode  string `json:"coupon_code" binding:"required"`
	OrderAmount int    `json:"order_amount" binding:"required"`
	NumItems    int    `json:"num_items" binding:"required"`
	Username    string `json:"username" binding:"required"`
}
