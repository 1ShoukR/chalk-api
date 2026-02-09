package openfoodfacts

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

const (
	baseURL        = "https://world.openfoodfacts.org"
	defaultTimeout = 10 * time.Second
)

// API defines the interface for Open Food Facts operations
type API interface {
	// SearchProducts searches for products by name/brand
	SearchProducts(query string, page int) (*SearchResponse, error)
	// GetProduct fetches a single product by barcode
	GetProduct(barcode string) (*Product, error)
}

// OpenFoodFacts implements the API interface
type OpenFoodFacts struct {
	httpClient *http.Client
	userAgent  string
}

// New creates a new Open Food Facts API instance
func New(userAgent string) *OpenFoodFacts {
	if userAgent == "" {
		userAgent = "ChalkAPI/1.0"
	}

	return &OpenFoodFacts{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		userAgent: userAgent,
	}
}

// SearchProducts searches for products matching the query
// Returns up to 24 products per page (API default)
func (o *OpenFoodFacts) SearchProducts(query string, page int) (*SearchResponse, error) {
	if page < 1 {
		page = 1
	}

	// Build URL with query params
	endpoint := fmt.Sprintf("%s/cgi/search.pl", baseURL)
	params := url.Values{}
	params.Set("search_terms", query)
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("page_size", "24")
	params.Set("json", "1")
	params.Set("fields", "code,product_name,brands,image_url,serving_size,serving_quantity,nutriments,nutriscore_grade,nova_group")

	fullURL := endpoint + "?" + params.Encode()

	resp, err := o.doRequest(fullURL)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result SearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result.Page = page
	return &result, nil
}

// GetProduct fetches a single product by barcode
func (o *OpenFoodFacts) GetProduct(barcode string) (*Product, error) {
	if barcode == "" {
		return nil, fmt.Errorf("barcode is required")
	}

	endpoint := fmt.Sprintf("%s/api/v2/product/%s.json", baseURL, barcode)
	params := url.Values{}
	params.Set("fields", "code,product_name,brands,image_url,serving_size,serving_quantity,nutriments,nutriscore_grade,nova_group")

	fullURL := endpoint + "?" + params.Encode()

	resp, err := o.doRequest(fullURL)
	if err != nil {
		return nil, fmt.Errorf("product request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Product not found
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product request returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result ProductResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Status != 1 {
		return nil, nil // Product not found
	}

	return &result.Product, nil
}

// doRequest performs an HTTP request with proper headers
func (o *OpenFoodFacts) doRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Open Food Facts requires a user-agent to identify the app
	req.Header.Set("User-Agent", o.userAgent)
	req.Header.Set("Accept", "application/json")

	slog.Debug("Open Food Facts request", "url", url)

	return o.httpClient.Do(req)
}
