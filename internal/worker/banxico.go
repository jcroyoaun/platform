package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

const (
	banxicoAPIURL = "https://www.banxico.org.mx/SieAPIRest/service/v1/series/SF43718/datos/oportuno"
)

// BanxicoClient handles communication with Banxico API
type BanxicoClient struct {
	token      string
	httpClient *http.Client
	logger     *slog.Logger
}

// BanxicoResponse represents the API response structure
type BanxicoResponse struct {
	BMX struct {
		Series []struct {
			Datos []struct {
				Fecha string `json:"fecha"`
				Dato  string `json:"dato"`
			} `json:"datos"`
		} `json:"series"`
	} `json:"bmx"`
}

// NewBanxicoClient creates a new Banxico API client
func NewBanxicoClient(token string, logger *slog.Logger) *BanxicoClient {
	if token == "" {
		logger.Warn("BANXICO_TOKEN environment variable not set - exchange rate updates will fail")
	}

	return &BanxicoClient{
		token: token,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		logger: logger,
	}
}

// GetExchangeRate fetches the latest USD/MXN exchange rate from Banxico
func (c *BanxicoClient) GetExchangeRate(ctx context.Context) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", banxicoAPIURL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Add Banxico API token header
	req.Header.Set("Bmx-Token", c.token)
	req.Header.Set("Accept", "application/json")

	c.logger.Info("fetching exchange rate from banxico")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch from banxico: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("banxico api returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	var data BanxicoResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, fmt.Errorf("failed to parse banxico response: %w", err)
	}

	// Extract the latest exchange rate
	if len(data.BMX.Series) == 0 {
		return 0, fmt.Errorf("no series data in banxico response")
	}

	series := data.BMX.Series[0]
	if len(series.Datos) == 0 {
		return 0, fmt.Errorf("no data points in banxico series")
	}

	// Get the most recent data point
	latestData := series.Datos[0]
	
	// Parse the exchange rate
	rate, err := strconv.ParseFloat(latestData.Dato, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse exchange rate '%s': %w", latestData.Dato, err)
	}

	// Validate the rate is reasonable (between 10 and 30 MXN per USD)
	if rate < 10.0 || rate > 30.0 {
		return 0, fmt.Errorf("exchange rate %f is outside reasonable bounds", rate)
	}

	c.logger.Info("exchange rate fetched", 
		"rate", rate, 
		"date", latestData.Fecha,
		"source", "banxico")

	return rate, nil
}

