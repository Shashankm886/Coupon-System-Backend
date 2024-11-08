package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Shashankm886/coupon_system_backend/models"
	"github.com/captaincodeman/couponcode"
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
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

func applyGruleRules(ruleFilePath string, ruleData *models.CouponRuleData) error {
	dataContext := ast.NewDataContext()
	dataContext.Add("CouponRuleData", ruleData)

	lib := ast.NewKnowledgeLibrary()
	ruleBuilder := builder.NewRuleBuilder(lib)

	ruleFile, err := os.Open(ruleFilePath)
	if err != nil {
		return err
	}
	defer ruleFile.Close()

	err = ruleBuilder.BuildRuleFromResource("CouponRules", "1.15.0", pkg.NewReaderResource(ruleFile))
	if err != nil {
		return err
	}

	knowledgeBase, _ := lib.NewKnowledgeBaseInstance("CouponRules", "1.15.0")
	engine := engine.NewGruleEngine()

	err = engine.Execute(dataContext, knowledgeBase)
	return err
}

func (service *couponService) Create(coupon models.CreateCoupon, couponsCollection *mongo.Collection, userCollection *mongo.Collection) (models.CreateCoupon, error) {
	ruleData := models.CouponRuleData{
		DiscountPercent: coupon.DiscountPercent,
		ExpiryDate:      coupon.ExpiryDate,
		MinOrderAmount:  coupon.OrderContent.MinAmount,
		MinOrderItems:   coupon.OrderContent.NumberOfItems,
		CurrentTime:     time.Now(),
		CouponValid:     false,
	}

	// Check if profile info exists and populate relevant fields
	if coupon.ProfileInfo != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		filter := bson.M{"username": coupon.ProfileInfo.Username}
		var user bson.M
		err := userCollection.FindOne(ctx, filter).Decode(&user)

		if err == nil {
			ruleData.ProfileInfoExists = true
			ruleData.ProfileUsername = coupon.ProfileInfo.Username
		} else {
			ruleData.ProfileInfoExists = false
			ruleData.Message = "User does not exist. Please try creating a coupon with an existing user."
		}
	}

	// Apply creation rules
	err := applyGruleRules("rules/create-coupon-rules.grl", &ruleData)
	if err != nil {
		log.Println("Error applying rules:", err)
		return models.CreateCoupon{}, err
	}

	if !ruleData.CouponValid {
		return models.CreateCoupon{}, errors.New(ruleData.Message)
	}

	code := couponcode.Generate()
	coupon.CouponCode = code

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = couponsCollection.InsertOne(ctx, coupon)
	if err != nil {
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

	var coupon models.CreateCoupon
	filter := bson.M{"coupon_code": request.CouponCode}
	err := couponsCollection.FindOne(ctx, filter).Decode(&coupon)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, fmt.Errorf("coupon not found")
		}
		return false, err
	}

	ruleData := models.CouponRuleData{
		DiscountPercent:  float64(coupon.Usage),
		ExpiryDate:       coupon.ExpiryDate,
		MinOrderAmount:   coupon.OrderContent.MinAmount,
		MinOrderItems:    coupon.OrderContent.NumberOfItems,
		OrderAmount:      request.OrderAmount,
		OrderItemCount:   request.NumItems,
		ExpectedUsername: request.Username,
		CurrentTime:      time.Now(),
		CouponValid:      true,
	}

	if coupon.ProfileInfo != nil {
		ruleData.ProfileInfoExists = true
		ruleData.ProfileUsername = coupon.ProfileInfo.Username
	}

	if coupon.OrderHistory != nil {
		orderFilter := bson.M{
			"username":  coupon.ProfileInfo.Username,
			"date":      bson.M{"$gte": coupon.OrderHistory.CheckTillDate},
			"is_coupon": false,
		}

		orderCount, err := ordersCollection.CountDocuments(ctx, orderFilter)
		if err != nil {
			return false, fmt.Errorf("error fetching orders, coupon could not be redeemed currently")
		}

		ruleData.OrderCountSinceDate = int(orderCount)
		ruleData.HasMinimumOrders = (orderCount >= int64(coupon.OrderHistory.MinOrdersWithCoupon))
	}

	err = applyGruleRules("rules/redeem-coupon-rules.grl", &ruleData)
	if err != nil {
		log.Println("Error applying rules:", err)
		return false, err
	}

	if !ruleData.CouponValid {
		return false, errors.New(ruleData.Message)
	}

	coupon.Usage -= 1
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
