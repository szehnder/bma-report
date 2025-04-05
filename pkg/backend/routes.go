package backend

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	app.Get("/api/bma-report", handleBMAReport)
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

// handleListAddresses returns all addresses.
func handleListAddresses(c *fiber.Ctx) error {
	ctx := context.Background()

	addrCol := BmaDB.Collection("addresses")

	cursor, err := addrCol.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer cursor.Close(ctx)

	var addresses []Address
	if err := cursor.All(ctx, &addresses); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(addresses)
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

// handleBMAReport returns a mock BMA report if there's a primary address and at least one enabled comparison address
func handleBMAReport(c *fiber.Ctx) error {
	ctx := context.Background()

	var primaryAddr Address
	var enabledAddrs []Address

	addrCol := BmaDB.Collection("addresses")

	// Find primary
	err := addrCol.FindOne(ctx, bson.M{"primary": true}).Decode(&primaryAddr)
	if err != nil {
		// If there's no primary, just return nothing special
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

	// Construct a naive "opinion"
	var addressesStrings []string
	var comparisonAddrs []*Address
	for _, a := range enabledAddrs {
		addressesStrings = append(addressesStrings, a.AddressStr)
		comparisonAddrs = append(comparisonAddrs, &a)
	}
	opinion := fmt.Sprintf("Broker Market Analysis for %s compared to:\n%s\n\nRecommendation: Price is in typical range.",
		primaryAddr.AddressStr,
		strings.Join(addressesStrings, "\n"),
	)

	report := BMAReport{
		PrimaryAddress:  &primaryAddr,
		ComparisonAddrs: comparisonAddrs,
		Opinion:         opinion,
	}

	return c.JSON(report)
}
