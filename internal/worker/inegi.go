package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// INEGI API endpoint for UMA (Indicator 539262)
	// Returns JSONP, so we need to strip the callback wrapper
	inegiAPIURL = "https://www.inegi.org.mx/app/api/indicadores/interna_v1_3/ValorIndicador/539262/00/null/es/null/null/3/pd/0/null/null/null/null/6/json/%s"
)

// INEGIClient handles communication with INEGI API
type INEGIClient struct {
	token      string
	httpClient *http.Client
	logger     *slog.Logger
}

// INEGIResponse represents the API response structure
// INEGI returns JSONP format: callback({"value": [...], "dimension": {...}})
type INEGIResponse struct {
	Value []string `json:"value"` // Array of UMA values, first is current year
	Dimension struct {
		Periods struct {
			Category struct {
				Label []struct {
					Key   string `json:"Key"`   // "P1", "P2", etc.
					Value string `json:"Value"` // "2025", "2024", etc.
				} `json:"label"`
			} `json:"category"`
		} `json:"periods"`
	} `json:"dimension"`
}

// UMAData represents the calculated UMA values
type UMAData struct {
	Annual  float64
	Monthly float64
	Daily   float64
}

// NewINEGIClient creates a new INEGI API client
func NewINEGIClient(token string, logger *slog.Logger) *INEGIClient {
	if token == "" {
		logger.Warn("INEGI_TOKEN environment variable not set - UMA updates will fail")
	}

	return &INEGIClient{
		token: token,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		logger: logger,
	}
}

// GetUMA fetches the latest UMA value from INEGI
func (c *INEGIClient) GetUMA(ctx context.Context) (*UMAData, error) {
	url := fmt.Sprintf(inegiAPIURL, c.token)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	c.logger.Info("fetching uma from inegi")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from inegi: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("inegi api returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// INEGI returns JSONP format: callback({"value": [...], ...})
	// We need to extract the JSON from the JSONP wrapper
	jsonStr := stripJSONPCallback(string(body))

	var data INEGIResponse
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("failed to parse inegi response: %w", err)
	}

	// Validate response
	if len(data.Value) == 0 {
		return nil, fmt.Errorf("no values in inegi response")
	}
	if len(data.Dimension.Periods.Category.Label) == 0 {
		return nil, fmt.Errorf("no period labels in inegi response")
	}

	// Extract UMA for the current year (2025)
	currentYear := time.Now().Year()
	yearStr := strconv.Itoa(currentYear)

	var umaAnnual float64
	var found bool
	
	// Find the index for the current year
	for i, period := range data.Dimension.Periods.Category.Label {
		if period.Value == yearStr {
			// Get the corresponding value
			if i < len(data.Value) && data.Value[i] != "NA" {
				// INEGI returns values with commas: "41,273.52"
				// We need to remove commas before parsing
				valueStr := strings.ReplaceAll(data.Value[i], ",", "")
				umaAnnual, err = strconv.ParseFloat(valueStr, 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse uma value '%s': %w", data.Value[i], err)
				}
				found = true
				c.logger.Info("found uma for year", "year", currentYear, "value", umaAnnual, "index", i)
				break
			}
		}
	}

	if !found {
		// If current year not found, use the first available value
		if len(data.Value) > 0 && data.Value[0] != "NA" {
			valueStr := strings.ReplaceAll(data.Value[0], ",", "")
			umaAnnual, err = strconv.ParseFloat(valueStr, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse latest uma value '%s': %w", data.Value[0], err)
			}
			c.logger.Info("using latest uma observation", 
				"year", data.Dimension.Periods.Category.Label[0].Value, 
				"value", umaAnnual)
		} else {
			return nil, fmt.Errorf("no valid uma values found in inegi data")
		}
	}

	// Validate the UMA is reasonable (between 30,000 and 60,000 MXN per year)
	if umaAnnual < 30000 || umaAnnual > 60000 {
		return nil, fmt.Errorf("uma annual %f is outside reasonable bounds (30k-60k)", umaAnnual)
	}

	// Calculate monthly and daily values
	umaMonthly := umaAnnual / 12.0
	umaDaily := umaAnnual / 365.0

	umaData := &UMAData{
		Annual:  umaAnnual,
		Monthly: umaMonthly,
		Daily:   umaDaily,
	}

	c.logger.Info("uma fetched and calculated", 
		"annual", umaAnnual, 
		"monthly", umaMonthly,
		"daily", umaDaily,
		"year", currentYear,
		"source", "inegi")

	return umaData, nil
}

// stripJSONPCallback removes the JSONP callback wrapper
// Input:  jQuery111209395642957513894_1762831315776({...})
// Output: {...}
func stripJSONPCallback(jsonp string) string {
	// Match pattern: callback({...})
	re := regexp.MustCompile(`^[a-zA-Z_$][a-zA-Z0-9_$]*\((.*)\);?$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(jsonp))
	
	if len(matches) > 1 {
		return matches[1]
	}
	
	// If no callback found, return as-is (might be pure JSON)
	return jsonp
}
