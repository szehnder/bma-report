package backend

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SetupRoutes defines all endpoints.
func SetupRoutes(app *fiber.App) {
	// Endpoint where the extension posts page data
	app.Post("/api/extension/page-data", handleReceivePageData)

	// Endpoints for the Svelte frontend
	app.Get("/api/addresses", handleListAddresses)
	app.Post("/api/addresses", handleCreateAddress)
	app.Patch("/api/addresses/:id", handleUpdateAddress)
	app.Delete("/api/addresses/:id", handleDeleteAddress)
	app.Get("/api/bma-report", handleBMAReport)
	app.Post("/api/bma-report/refresh", handleRefreshBMAReport)
	app.Get("/api/llm-instructions", handleGetLLMInstructions)
	app.Post("/api/llm-instructions", handleUpdateLLMInstructions)
}

// handleReceivePageData saves the raw content in MongoDB
// Then calls the LLM to extract property details and saves that in the raw_page_data collection.
func handleReceivePageData(c *fiber.Ctx) error {
	ctx := context.Background()

	var data RawPageData
	if err := c.BodyParser(&data); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	log.Info().Str("url", data.URL).Msg("Received page data from extension")

	// Extract property details using Gemini
	details, err := ExtractPropertyDetails(data.Content)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract property details")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to extract property details: %v", err),
		})
	}

	log.Info().Str("address", details.Address).Msg("Successfully extracted property details")

	data.PropertyDetails = details

	// Check if we already have this address in the database
	rawCol := BmaDB.Collection("raw_page_data")
	filter := bson.M{"propertyDetails.address": details.Address}
	update := bson.M{
		"$set": bson.M{
			"url":             data.URL,
			"content":         data.Content,
			"propertyDetails": details,
		},
	}
	opts := options.Update().SetUpsert(true)

	result, err := rawCol.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Error().Err(err).Str("address", details.Address).Msg("Failed to update raw page data")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// If this was an insert (not an update), create a new Address record
	if result.UpsertedID != nil {
		addr := Address{
			RawPageID:  result.UpsertedID.(primitive.ObjectID),
			AddressStr: details.Address,
			Enabled:    false, // default to false
			Primary:    false,
		}
		addrCol := BmaDB.Collection("addresses")
		_, err = addrCol.InsertOne(ctx, addr)
		if err != nil {
			log.Error().Err(err).Str("address", details.Address).Msg("Failed to create new address record")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		log.Info().Str("address", details.Address).Msg("Created new address record")
	} else {
		log.Info().Str("address", details.Address).Msg("Updated existing address record")
	}

	return c.JSON(fiber.Map{
		"message":  "Page data processed and property details extracted.",
		"upserted": result.UpsertedID != nil,
	})
}

// handleListAddresses returns all addresses with their property details.
func handleListAddresses(c *fiber.Ctx) error {
	ctx := context.Background()

	addrCol := BmaDB.Collection("addresses")
	rawCol := BmaDB.Collection("raw_page_data")

	cursor, err := addrCol.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer cursor.Close(ctx)

	var addresses []Address
	if err := cursor.All(ctx, &addresses); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Create a response type that includes property details
	type AddressWithDetails struct {
		ID            primitive.ObjectID `json:"id"`
		AddressStr    string             `json:"addressStr"`
		Enabled       bool               `json:"enabled"`
		Primary       bool               `json:"primary"`
		Price         *float64           `json:"price,omitempty"`
		Bedrooms      *int               `json:"bedrooms,omitempty"`
		Bathrooms     *float64           `json:"bathrooms,omitempty"`
		SquareFootage *int               `json:"squareFootage,omitempty"`
		PropertyType  *string            `json:"propertyType,omitempty"`
		YearBuilt     *int               `json:"yearBuilt,omitempty"`
	}

	// Get property details for each address
	response := make([]AddressWithDetails, len(addresses))
	for i, addr := range addresses {
		response[i] = AddressWithDetails{
			ID:         addr.ID,
			AddressStr: addr.AddressStr,
			Enabled:    addr.Enabled,
			Primary:    addr.Primary,
		}

		// Get property details from raw_page_data
		var raw RawPageData
		err := rawCol.FindOne(ctx, bson.M{"_id": addr.RawPageID}).Decode(&raw)
		if err == nil && raw.PropertyDetails != nil {
			response[i].Price = &raw.PropertyDetails.Price
			response[i].Bedrooms = &raw.PropertyDetails.Bedrooms
			response[i].Bathrooms = &raw.PropertyDetails.Bathrooms
			response[i].SquareFootage = &raw.PropertyDetails.SquareFootage
			response[i].PropertyType = &raw.PropertyDetails.PropertyType
			response[i].YearBuilt = &raw.PropertyDetails.YearBuilt
		}
	}

	return c.JSON(response)
}

// handleCreateAddress – if you ever want to manually add addresses via API
func handleCreateAddress(c *fiber.Ctx) error {
	ctx := context.Background()

	var addr Address
	if err := c.BodyParser(&addr); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}

	_, err := BmaDB.Collection("addresses").InsertOne(ctx, addr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Address created"})
}

// handleUpdateAddress can toggle enabled, set primary, etc.
func handleUpdateAddress(c *fiber.Ctx) error {
	ctx := context.Background()
	idParam := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid address ID"})
	}

	var updateData Address
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}

	// If address is set to Primary = true, then we need to reset any other addresses from being primary.
	if updateData.Primary {
		// Unset primary from all addresses first
		_, err := BmaDB.Collection("addresses").UpdateMany(ctx, bson.M{"primary": true}, bson.M{
			"$set": bson.M{"primary": false},
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}

	// Perform the actual address update
	update := bson.M{}
	if updateData.Enabled {
		update["enabled"] = true
	} else {
		update["enabled"] = false
	}
	// For primary
	if updateData.Primary {
		update["primary"] = true
	} else {
		update["primary"] = false
	}

	_, err = BmaDB.Collection("addresses").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{
		"$set": update,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Address updated"})
}

// handleDeleteAddress deletes an address and its associated raw page data
func handleDeleteAddress(c *fiber.Ctx) error {
	ctx := context.Background()
	idParam := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid address ID"})
	}

	// First get the address to find its RawPageID
	var addr Address
	err = addressesCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&addr)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Address not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Delete the address
	_, err = addressesCollection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Delete the associated raw page data
	rawCol := BmaDB.Collection("raw_page_data")
	_, err = rawCol.DeleteOne(ctx, bson.M{"_id": addr.RawPageID})
	if err != nil && err != mongo.ErrNoDocuments {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Address deleted successfully"})
}

// handleBMAReport returns a mock BMA report if there's a primary address and at least one enabled comparison address
func handleBMAReport(c *fiber.Ctx) error {
	ctx := context.Background()

	var primaryAddr Address
	var enabledAddrs []Address

	addrCol := BmaDB.Collection("addresses")
	rawCol := BmaDB.Collection("raw_page_data")
	cachedCol := BmaDB.Collection("cached_bma_reports")

	// Find primary
	err := addrCol.FindOne(ctx, bson.M{"primary": true}).Decode(&primaryAddr)
	if err != nil {
		return c.JSON(BMAReport{
			PrimaryAddress:  nil,
			ComparisonAddrs: nil,
			Opinion:         "No primary address set yet.",
		})
	}

	// Find enabled (but not primary)
	cursor, err := addrCol.Find(ctx, bson.M{"enabled": true, "primary": false})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer cursor.Close(ctx)
	if err := cursor.All(ctx, &enabledAddrs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if len(enabledAddrs) == 0 {
		return c.JSON(BMAReport{
			PrimaryAddress:  &primaryAddr,
			ComparisonAddrs: nil,
			Opinion:         "Need at least one enabled comparison address.",
		})
	}

	// Check for cached report
	var comparisonIDs []primitive.ObjectID
	for _, addr := range enabledAddrs {
		comparisonIDs = append(comparisonIDs, addr.ID)
	}
	sort.Slice(comparisonIDs, func(i, j int) bool {
		return comparisonIDs[i].Hex() < comparisonIDs[j].Hex()
	})

	var cachedReport CachedBMAReport
	err = cachedCol.FindOne(ctx, bson.M{
		"primaryAddressId":     primaryAddr.ID,
		"comparisonAddressIds": comparisonIDs,
	}).Decode(&cachedReport)

	// If we have a cached report less than 24 hours old, use it
	if err == nil && time.Since(cachedReport.GeneratedAt) < 24*time.Hour {
		return c.JSON(cachedReport.Report)
	}

	// Get property details for primary address
	var primaryRaw RawPageData
	err = rawCol.FindOne(ctx, bson.M{"_id": primaryAddr.RawPageID}).Decode(&primaryRaw)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get primary property details"})
	}

	// Get property details for comparison addresses
	var comparisonDetails []*PropertyDetails
	for _, addr := range enabledAddrs {
		var raw RawPageData
		err = rawCol.FindOne(ctx, bson.M{"_id": addr.RawPageID}).Decode(&raw)
		if err != nil {
			continue
		}
		if raw.PropertyDetails != nil {
			// Create a copy without the listing date
			details := *raw.PropertyDetails
			details.DaysOnMarket = 0
			details.LastPriceChange = 0
			comparisonDetails = append(comparisonDetails, &details)
		}
	}

	// Generate detailed analysis
	nonPtrComparisons := make([]PropertyDetails, len(comparisonDetails))
	for i, comp := range comparisonDetails {
		nonPtrComparisons[i] = *comp
	}
	detailedAnalysis, err := GenerateDetailedBMA(*primaryRaw.PropertyDetails, nonPtrComparisons)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate detailed analysis")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate analysis"})
	}

	// Construct the report
	var comparisonAddrs []*Address
	for _, a := range enabledAddrs {
		comparisonAddrs = append(comparisonAddrs, &a)
	}

	report := BMAReport{
		PrimaryAddress:   &primaryAddr,
		ComparisonAddrs:  comparisonAddrs,
		Opinion:          detailedAnalysis.Recommendation,
		DetailedAnalysis: &detailedAnalysis,
	}

	// Cache the report
	cachedReport = CachedBMAReport{
		PrimaryAddressID:     primaryAddr.ID,
		ComparisonAddressIDs: comparisonIDs,
		GeneratedAt:          time.Now(),
		Report:               report,
	}

	_, err = cachedCol.UpdateOne(
		ctx,
		bson.M{
			"primaryAddressId":     primaryAddr.ID,
			"comparisonAddressIds": comparisonIDs,
		},
		bson.M{"$set": cachedReport},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to cache BMA report")
	}

	return c.JSON(report)
}

// handleRefreshBMAReport forces a refresh of the BMA report by deleting the cache and regenerating
func handleRefreshBMAReport(c *fiber.Ctx) error {
	ctx := context.Background()

	var primaryAddr Address
	var enabledAddrs []Address

	addrCol := BmaDB.Collection("addresses")
	cachedCol := BmaDB.Collection("cached_bma_reports")

	// Find primary
	err := addrCol.FindOne(ctx, bson.M{"primary": true}).Decode(&primaryAddr)
	if err != nil {
		return c.JSON(BMAReport{
			PrimaryAddress:  nil,
			ComparisonAddrs: nil,
			Opinion:         "No primary address set yet.",
		})
	}

	// Find enabled (but not primary)
	cursor, err := addrCol.Find(ctx, bson.M{"enabled": true, "primary": false})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer cursor.Close(ctx)
	if err := cursor.All(ctx, &enabledAddrs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if len(enabledAddrs) == 0 {
		return c.JSON(BMAReport{
			PrimaryAddress:  &primaryAddr,
			ComparisonAddrs: nil,
			Opinion:         "Need at least one enabled comparison address.",
		})
	}

	// Delete cached report if it exists
	var comparisonIDs []primitive.ObjectID
	for _, addr := range enabledAddrs {
		comparisonIDs = append(comparisonIDs, addr.ID)
	}
	sort.Slice(comparisonIDs, func(i, j int) bool {
		return comparisonIDs[i].Hex() < comparisonIDs[j].Hex()
	})

	_, err = cachedCol.DeleteOne(ctx, bson.M{
		"primaryAddressId":     primaryAddr.ID,
		"comparisonAddressIds": comparisonIDs,
	})
	if err != nil && err != mongo.ErrNoDocuments {
		log.Error().Err(err).Msg("Failed to delete cached report")
	}

	// Now generate a fresh report
	return handleBMAReport(c)
}

func handleGetLLMInstructions(c *fiber.Ctx) error {
	var instructions LLMInstructions
	err := llmInstructionsCollection.FindOne(context.Background(), bson.M{}).Decode(&instructions)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(fiber.Map{"instructions": ""})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch instructions"})
	}

	return c.JSON(fiber.Map{"instructions": instructions.Instructions})
}

func handleUpdateLLMInstructions(c *fiber.Ctx) error {
	var data struct {
		Instructions string `json:"instructions"`
	}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Delete existing instructions if any
	_, err := llmInstructionsCollection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to clear existing instructions"})
	}

	// Insert new instructions
	instructions := LLMInstructions{
		Instructions: data.Instructions,
		UpdatedAt:    time.Now(),
	}
	_, err = llmInstructionsCollection.InsertOne(context.Background(), instructions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save instructions"})
	}

	// Delete cached BMA report to force regeneration
	_, err = cachedBMAReportsCollection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to clear cached reports"})
	}

	return c.JSON(fiber.Map{"status": "success"})
}
