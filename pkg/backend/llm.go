package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
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
