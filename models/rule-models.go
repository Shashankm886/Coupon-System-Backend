package models

import (
	"time"
)

type CouponRuleData struct {
	// Coupon-specific fields
	DiscountPercent float64 // Discount percentage
	Usage           int
	ExpiryDate      time.Time // Expiry date of the coupon
	IsFrequent      bool      // Indicates if the user is marked as frequent
	MinOrderAmount  int       // Minimum order amount for coupon validity
	MinOrderItems   int       // Minimum number of items required for coupon

	// Profile information fields
	ProfileInfoExists bool   // Indicates if profile information exists
	ProfileUsername   string // Username associated with the profile info
	ExpectedUsername  string // Username attempting to redeem the coupon

	// Order-related fields
	OrderAmount         int  // Total amount of the order
	OrderItemCount      int  // Number of items in the order
	OrderHistoryExists  bool // Indicates if order history information exists
	OrderCountSinceDate int  // Number of orders since a given date
	HasMinimumOrders    bool // Flag to check if min orders requirement is met

	// System and validation fields
	CurrentTime time.Time // Current time to validate expiry
	CouponValid bool      // Final validation result for the coupon
	Message     string    // Message describing validation result
}
