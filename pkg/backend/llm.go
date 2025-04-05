package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/option"
)

// ExtractPropertyDetails uses Gemini to parse property details from the content
func ExtractPropertyDetails(content string) (*PropertyDetails, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")
	prompt := `Extract the following property details from the given real estate listing text. 
	Return ONLY a JSON object with these exact fields (use null for missing values):
	{
		"address": "string",
		"price": number,
		"bedrooms": number,
		"bathrooms": number,
		"squareFootage": number,
		"yearBuilt": number,
		"propertyType": "string",
		"lotSize": "string",
		"mlsNumber": "string",
		"daysOnMarket": number,
		"lastPriceChange": number,
		"description": "string"
	}

	Here is the listing text:
	` + content

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content generated")
	}

	// Extract the JSON response from the model's output
	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	// Clean up the response to get just the JSON
	start := strings.Index(responseText, "{")
	end := strings.LastIndex(responseText, "}")
	if start == -1 || end == -1 {
		return nil, fmt.Errorf("invalid response format")
	}
	jsonStr := responseText[start : end+1]

	var details PropertyDetails
	if err := json.Unmarshal([]byte(jsonStr), &details); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	return &details, nil
}

// ExtractAddressFromPage extracts the address from property details
func ExtractAddressFromPage(content string) (string, error) {
	details, err := ExtractPropertyDetails(content)
	if err != nil {
		return "", err
	}
	return details.Address, nil
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func formatPropertyDetails(property PropertyDetails) string {
	jsonData, _ := json.MarshalIndent(property, "", "  ")
	return string(jsonData)
}

func formatComparisonProperties(properties []PropertyDetails) string {
	var builder strings.Builder
	for i, prop := range properties {
		builder.WriteString(fmt.Sprintf("Comparison Property %d:\n", i+1))
		jsonData, _ := json.MarshalIndent(prop, "", "  ")
		builder.WriteString(string(jsonData))
		builder.WriteString("\n\n")
	}
	return builder.String()
}

// GenerateDetailedBMA generates a comprehensive BMA analysis using Gemini
func GenerateDetailedBMA(primary PropertyDetails, comparisons []PropertyDetails) (DetailedAnalysis, error) {
	// Get LLM instructions
	var instructions LLMInstructions
	err := llmInstructionsCollection.FindOne(context.Background(), bson.M{}).Decode(&instructions)
	if err != nil && err != mongo.ErrNoDocuments {
		return DetailedAnalysis{}, fmt.Errorf("error fetching LLM instructions: %v", err)
	}

	prompt := fmt.Sprintf(`Generate a detailed BMA (Broker Market Analysis) report comparing the following properties:

Primary Property:
%s

Comparison Properties:
%s

Additional Instructions:
%s

Please provide a comprehensive analysis including:
1. Price analysis comparing the primary property to the comparisons
2. Detailed feature comparison (bedrooms, bathrooms, square footage, etc.)
3. Market trends and context
4. Final recommendation

Format the response as a JSON object with the following structure:
{
    "primaryPropertyDetails": {
        "address": "string",
        "price": number,
        "bedrooms": number,
        "bathrooms": number,
        "squareFootage": number,
        "yearBuilt": number,
        "propertyType": "string",
        "lotSize": "string",
        "mlsNumber": "string",
        "daysOnMarket": number,
        "lastPriceChange": number,
        "description": "string"
    },
    "comparisonDetails": [
        {
            "address": "string",
            "price": number,
            "bedrooms": number,
            "bathrooms": number,
            "squareFootage": number,
            "yearBuilt": number,
            "propertyType": "string",
            "lotSize": "string",
            "mlsNumber": "string",
            "daysOnMarket": number,
            "lastPriceChange": number,
            "description": "string"
        }
    ],
    "priceAnalysis": "string",
    "featureComparison": [
        {
            "feature": "string",
            "primaryValue": "string",
            "comparison": [
                {
                    "address": "string",
                    "value": "string"
                }
            ],
            "analysis": "string"
        }
    ],
    "marketTrends": "string",
    "recommendation": "string"
}`, formatPropertyDetails(primary), formatComparisonProperties(comparisons), instructions.Instructions)

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		return DetailedAnalysis{}, fmt.Errorf("failed to create Gemini client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-pro")

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return DetailedAnalysis{}, fmt.Errorf("failed to generate content: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return DetailedAnalysis{}, fmt.Errorf("no content generated")
	}

	responseText := resp.Candidates[0].Content.Parts[0].(genai.Text)
	start := strings.Index(string(responseText), "{")
	end := strings.LastIndex(string(responseText), "}")
	if start == -1 || end == -1 {
		return DetailedAnalysis{}, fmt.Errorf("invalid response format")
	}
	jsonStr := string(responseText)[start : end+1]

	var analysis DetailedAnalysis
	if err := json.Unmarshal([]byte(jsonStr), &analysis); err != nil {
		return DetailedAnalysis{}, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	// Set the actual property details
	analysis.PrimaryPropertyDetails = primary
	analysis.ComparisonDetails = comparisons

	return analysis, nil
}

// Helper function to safely marshal JSON
func mustJSON(v interface{}) []byte {
	b, _ := json.MarshalIndent(v, "", "  ")
	return b
}
