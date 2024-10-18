package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID       primitive.ObjectID `json:"_id" bson:"_id"`
	Date     time.Time          `json:"date" bson:"date"`
	Amount   int                `json:"amount" bson:"amount"`
	Username string             `json:"username" bson:"username"`
	IsCoupon bool               `json:"is_coupon" bson:"is_coupon"`
}
