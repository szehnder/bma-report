package backend

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PropertyDetails contains all the relevant information extracted from a property listing
type PropertyDetails struct {
	Address         string  `bson:"address" json:"address"`
	Price           float64 `bson:"price" json:"price"`
	Bedrooms        int     `bson:"bedrooms" json:"bedrooms"`
	Bathrooms       float64 `bson:"bathrooms" json:"bathrooms"`
	SquareFootage   int     `bson:"squareFootage" json:"squareFootage"`
	YearBuilt       int     `bson:"yearBuilt" json:"yearBuilt"`
	PropertyType    string  `bson:"propertyType" json:"propertyType"`
	LotSize         string  `bson:"lotSize" json:"lotSize"`
	MLSNumber       string  `bson:"mlsNumber" json:"mlsNumber"`
	DaysOnMarket    int     `bson:"daysOnMarket" json:"daysOnMarket"`
	LastPriceChange float64 `bson:"lastPriceChange" json:"lastPriceChange"`
	Description     string  `bson:"description" json:"description"`
}

// RawPageData is the raw request from the extension.
type RawPageData struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	URL             string             `bson:"url" json:"url"`
	Content         string             `bson:"content" json:"content"`
	PropertyDetails *PropertyDetails   `bson:"propertyDetails,omitempty" json:"propertyDetails,omitempty"`
}

// Address represents an address extracted and stored for BMA analysis.
type Address struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	RawPageID  primitive.ObjectID `bson:"rawPageId,omitempty" json:"rawPageId,omitempty"`
	AddressStr string             `bson:"addressStr" json:"addressStr"`
	Enabled    bool               `bson:"enabled" json:"enabled"`
	Primary    bool               `bson:"primary" json:"primary"`
}

// BMAReport holds the result of the broker market analysis
// for the "primary" address vs. a set of "comparison" addresses.
type BMAReport struct {
	PrimaryAddress   *Address          `json:"primaryAddress"`
	ComparisonAddrs  []*Address        `json:"comparisonAddresses"`
	Opinion          string            `json:"opinion"`
	DetailedAnalysis *DetailedAnalysis `json:"detailedAnalysis,omitempty"`
}

// DetailedAnalysis provides a comprehensive breakdown of the BMA comparison
type DetailedAnalysis struct {
	PrimaryPropertyDetails PropertyDetails     `json:"primaryPropertyDetails" bson:"primaryPropertyDetails"`
	ComparisonDetails      []PropertyDetails   `json:"comparisonDetails" bson:"comparisonDetails"`
	PriceAnalysis          string              `json:"priceAnalysis" bson:"priceAnalysis"`
	FeatureComparison      []FeatureComparison `json:"featureComparison" bson:"featureComparison"`
	MarketTrends           string              `json:"marketTrends" bson:"marketTrends"`
	Recommendation         string              `json:"recommendation" bson:"recommendation"`
}

// FeatureComparison compares specific features between properties
type FeatureComparison struct {
	Feature      string            `json:"feature"`
	PrimaryValue string            `json:"primaryValue"`
	Comparison   []ComparisonValue `json:"comparison"`
	Analysis     string            `json:"analysis"`
}

// ComparisonValue represents a comparison property's value for a feature
type ComparisonValue struct {
	Address string `json:"address"`
	Value   string `json:"value"`
}

// CachedBMAReport stores the BMA analysis results to avoid reprocessing
type CachedBMAReport struct {
	ID                   primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	PrimaryAddressID     primitive.ObjectID   `bson:"primaryAddressId" json:"primaryAddressId"`
	ComparisonAddressIDs []primitive.ObjectID `bson:"comparisonAddressIds" json:"comparisonAddressIds"`
	GeneratedAt          time.Time            `bson:"generatedAt" json:"generatedAt"`
	Report               BMAReport            `bson:"report" json:"report"`
}

type LLMInstructions struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Instructions string             `bson:"instructions"`
	UpdatedAt    time.Time          `bson:"updated_at"`
}
