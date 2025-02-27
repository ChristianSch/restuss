package restuss

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jpillora/backoff"
)

// Client expose the methods callable on Nessus Api
type Client interface {
	GetScanTemplates() ([]*ScanTemplate, error)
	LaunchScan(scanID int64) error
	StopScan(scanID int64) error
	DeleteScan(scanID int64) error
	CreateScan(scan *Scan) (*PersistedScan, error)
	GetScans(lastModificationDate int64) ([]*PersistedScan, error)
	GetScanByID(id int64) (*ScanDetail, error)
	GetPluginByID(id int64) (*Plugin, error)
	GetPluginOutput(scanID, hostID, pluginID int64) (*PluginOutputResponse, error)
	GetAssetByName(name string) (*Asset, error)
	GetFindingsByAssetName(name string) ([]Finding, error)
}

// NessusClient implements nessus.Client
type NessusClient struct {
	auth       AuthProvider
	url        string
	httpClient *http.Client
}

// NewClient returns a new NessusClient
func NewClient(auth AuthProvider, url string, allowInsecureConnection bool) (*NessusClient, error) {
	var c *http.Client

	if allowInsecureConnection {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		c = &http.Client{Transport: tr}
	} else {
		c = &http.Client{}
	}

	err := auth.Prepare(url, c)
	if err != nil {
		return nil, errors.New("Failed to prepare auth provider: " + err.Error())
	}

	return &NessusClient{auth: auth, url: url, httpClient: c}, nil
}

// GetScanTemplatesContext retrieves the Scan templates ussing the given context.
func (c *NessusClient) GetScanTemplatesContext(ctx context.Context) ([]*ScanTemplate, error) {
	req, err := http.NewRequest(http.MethodGet, c.url+"/editor/scan/templates", nil)
	if err != nil {
		return nil, errors.New("Unable to create request object: " + err.Error())
	}
	req = req.WithContext(ctx)
	var data struct {
		Templates []*ScanTemplate `json:"templates"`
	}
	err = c.performCallAndReadResponse(req, &data)
	if err != nil {
		return nil, errors.New("Call failed: " + err.Error())
	}
	return data.Templates, nil
}

// GetScanTemplates retrieves Scan Templates
func (c *NessusClient) GetScanTemplates() ([]*ScanTemplate, error) {
	return c.GetScanTemplatesContext(context.Background())
}

// LaunchScan launch spe scan with the specified scanID
func (c *NessusClient) LaunchScan(scanID int64) error {
	return c.LaunchScanContext(context.Background(), scanID)
}

// LaunchScanContext launch the scan with the specified scanID and context.
func (c *NessusClient) LaunchScanContext(ctx context.Context, scanID int64) error {
	path := "/scans/" + strconv.FormatInt(scanID, 10) + "/launch"
	req, err := http.NewRequest(http.MethodPost, c.url+path, nil)
	if err != nil {
		return errors.New("Unable to create request object: " + err.Error())
	}
	req = req.WithContext(ctx)
	return c.performCallAndReadResponse(req, nil)
}

// StopScan stops the scan with the given scanID
func (c *NessusClient) StopScan(scanID int64) error {
	return c.StopScanContext(context.Background(), scanID)
}

// StopScanContext stops the scan with the given scanID
func (c *NessusClient) StopScanContext(ctx context.Context, scanID int64) error {
	path := "/scans/" + strconv.FormatInt(scanID, 10) + "/stop"
	req, err := http.NewRequest(http.MethodPost, c.url+path, nil)
	if err != nil {
		return errors.New("Unable to create request object: " + err.Error())
	}
	req = req.WithContext(ctx)
	return c.performCallAndReadResponse(req, nil)
}

// DeleteScan will remove the scan with the given scanID
func (c *NessusClient) DeleteScan(scanID int64) error {
	return c.DeleteScanContext(context.Background(), scanID)
}

// DeleteScanContext will remove the scan with the given scanID and context.
func (c *NessusClient) DeleteScanContext(ctx context.Context, scanID int64) error {
	path := "/scans/" + strconv.FormatInt(scanID, 10)
	req, err := http.NewRequest(http.MethodDelete, c.url+path, nil)
	if err != nil {
		return errors.New("Unable to create request object: " + err.Error())
	}
	req = req.WithContext(ctx)
	return c.performCallAndReadResponse(req, nil)
}

// CreateScan creates a scan
func (c *NessusClient) CreateScan(scan *Scan) (*PersistedScan, error) {
	return c.CreateScanContext(context.Background(), scan)
}

// CreateScanContext creates a scan with the given scan data and context.
func (c *NessusClient) CreateScanContext(ctx context.Context, scan *Scan) (*PersistedScan, error) {
	jsonBody, err := json.Marshal(scan)
	if err != nil {
		return nil, errors.New("Unable to marshall request body" + err.Error())
	}
	req, err := http.NewRequest(http.MethodPost, c.url+"/scans", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, errors.New("Unable to create request object: " + err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	var result struct {
		Scan PersistedScan `json:"scan"`
	}
	req = req.WithContext(ctx)
	err = c.performCallAndReadResponse(req, &result)
	if err != nil {
		return nil, err
	}

	return &result.Scan, nil
}

// GetScans get a list of scan matching the provided lastModificationDate (check Nessus documentation)
func (c *NessusClient) GetScans(lastModificationDate int64) ([]*PersistedScan, error) {
	return c.GetScansContext(context.Background(), lastModificationDate)
}

// GetScansContext get a list of scan matching the provided lastModificationDate (check Nessus documentation) and context.
func (c *NessusClient) GetScansContext(ctx context.Context, lastModificationDate int64) ([]*PersistedScan, error) {
	req, err := http.NewRequest(http.MethodGet, c.url+"/scans", nil)
	if err != nil {
		return nil, errors.New("Unable to create request object: " + err.Error())
	}

	if lastModificationDate > 0 {
		q := req.URL.Query()
		q.Add("last_modification_date", strconv.FormatInt(lastModificationDate, 10))
		req.URL.RawQuery = q.Encode()
	}

	var data struct {
		Scans []*PersistedScan `json:"scans"`
	}
	req = req.WithContext(ctx)
	err = c.performCallAndReadResponse(req, &data)
	if err != nil {
		return nil, err
	}

	return data.Scans, nil
}

// GetScanByID retrieve a scan by ID
func (c *NessusClient) GetScanByID(ID int64) (*ScanDetail, error) {
	return c.GetScanByIDContext(context.Background(), ID)
}

// GetScanByIDContext retrieve a scan by ID
func (c *NessusClient) GetScanByIDContext(ctx context.Context, ID int64) (*ScanDetail, error) {
	path := fmt.Sprintf("/scans/%d", ID)

	req, err := http.NewRequest(http.MethodGet, c.url+path, nil)
	if err != nil {
		return nil, errors.New("Unable to create request object: " + err.Error())
	}

	scanDetail := &ScanDetail{}
	req = req.WithContext(ctx)
	err = c.performCallAndReadResponse(req, &scanDetail)
	if err != nil {
		return nil, err
	}

	scanDetail.ID = ID

	return scanDetail, nil
}

// GetPluginByID retrieves a plugin by ID
func (c *NessusClient) GetPluginByID(ID int64) (*Plugin, error) {
	return c.GetPluginByIDContext(context.Background(), ID)
}

// GetPluginByIDContext retrieves a plugin by ID using the given context.
func (c *NessusClient) GetPluginByIDContext(ctx context.Context, ID int64) (*Plugin, error) {
	path := fmt.Sprintf("/plugins/plugin/%d", ID)

	req, err := http.NewRequest(http.MethodGet, c.url+path, nil)
	if err != nil {
		return nil, errors.New("Unable to create request object: " + err.Error())
	}

	p := &Plugin{}
	req = req.WithContext(ctx)
	err = c.performCallAndReadResponse(req, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// GetPluginOutput retrieves output from a plugin ran against a target
func (c *NessusClient) GetPluginOutput(scanID, hostID, pluginID int64) (*PluginOutputResponse, error) {
	return c.GetPluginOutputContext(context.Background(), scanID, hostID, pluginID)
}

// GetPluginOutputContext retrieves output from a plugin ran against a target using the given context.
func (c *NessusClient) GetPluginOutputContext(ctx context.Context, scanID, hostID, pluginID int64) (*PluginOutputResponse, error) {
	path := fmt.Sprintf("/scans/%d/hosts/%d/plugins/%d", scanID, hostID, pluginID)

	req, err := http.NewRequest(http.MethodGet, c.url+path, nil)
	if err != nil {
		return nil, errors.New("Unable to create request object: " + err.Error())
	}

	output := &PluginOutputResponse{}
	req = req.WithContext(ctx)
	err = c.performCallAndReadResponse(req, output)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// GetPolicyByID retrieves a policy by ID
func (c *NessusClient) GetPolicyByID(ID int64) (*Policy, error) {
	path := fmt.Sprintf("/policies/%d", ID)

	req, err := http.NewRequest(http.MethodGet, c.url+path, nil)
	if err != nil {
		return nil, errors.New("Unable to create request object: " + err.Error())
	}

	p := &Policy{}

	err = c.performCallAndReadResponse(req, p)
	if err != nil {
		return nil, err
	}
	p.ID = ID

	return p, nil
}

// GetPolicyByIDContext retrieves a policy by ID using the given context.
func (c *NessusClient) GetPolicyByIDContext(ctx context.Context, ID int64) (*Policy, error) {
	path := fmt.Sprintf("/policies/%d", ID)

	req, err := http.NewRequest(http.MethodGet, c.url+path, nil)
	if err != nil {
		return nil, errors.New("Unable to create request object: " + err.Error())
	}

	p := &Policy{}
	req = req.WithContext(ctx)
	err = c.performCallAndReadResponse(req, p)
	if err != nil {
		return nil, err
	}
	p.ID = ID

	return p, nil
}

// GetAssetByName returns an asset by its name. Returns an error if more than
// one or none assets are matching.
func (c *NessusClient) GetAssetByName(ctx context.Context, name string) (*Asset, error) {
	path := "/api/v3/assets/search"

	payload := map[string]interface{}{
		"filter": map[string]interface{}{
			"and": []interface{}{
				map[string]string{
					"property": "name",
					"operator": "eq",
					"value":    name,
				},
			},
		},
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.New("Unable to marshall request body" + err.Error())
	}

	req, err := http.NewRequest(http.MethodPost, c.url+path, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, errors.New("Unable to create request object: " + err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	var result struct {
		Assets []Asset `json:"assets"`
	}

	req = req.WithContext(ctx)
	err = c.performCallAndReadResponse(req, &result)
	if err != nil {
		return nil, err
	}

	if len(result.Assets) == 0 {
		return nil, fmt.Errorf("No assets matching name: %v", name)
	}
	if count := len(result.Assets); count > 1 {
		return nil, fmt.Errorf("More than one asset matching name: %v (%d)", name, count)
	}

	return &result.Assets[0], nil
}

// GetFindingsByAssetName returns all the findings associated to an asset by
// its name.
func (c *NessusClient) GetFindingsByAssetName(ctx context.Context, name string) ([]Finding, error) {
	var findings []Finding
	path := "/api/v3/findings/vulnerabilities/host/search"

	payload := map[string]interface{}{
		"filter": map[string]interface{}{
			"and": []interface{}{
				map[string]string{
					"property": "asset.name",
					"operator": "eq",
					"value":    name,
				},
			},
		},
		// NOTE: there are more fields available, we are using just those that
		// are meaningful to us.
		"fields": []string{
			"output",
			"id",
			"severity",
			"port",
			"protocol",
			"service",
			"plugin_id",
			"name",
			"description",
			"synopsis",
			"cvss3_base_score",
			"cvss2_base_score",
			"cwe",
			"see_also",
		},
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.New("Unable to marshall request body" + err.Error())
	}

	req, err := http.NewRequest(http.MethodPost, c.url+path, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, errors.New("Unable to create request object: " + err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json") // Required.

	var result struct {
		Findings   []Finding  `json:"findings"`
		Pagination Pagination `json:"pagination"`
	}

	req = req.WithContext(ctx)

	err = c.performCallAndReadResponse(req, &result)
	if err != nil {
		return nil, err
	}

	findings = append(findings, result.Findings...)

	// Number of results are paginated. When `Next` is not empty, just send the
	// value as a parameter for the next request to get the next page.
	for next := result.Pagination.Next; next != ""; {
		jsonNext := fmt.Sprintf("{\"next\":\"%s\"}", next)
		req, err = http.NewRequest(http.MethodPost, c.url+path, strings.NewReader(jsonNext))
		if err != nil {
			return nil, errors.New("Unable to create request object: " + err.Error())
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json") // Required.

		var result struct {
			Findings   []Finding  `json:"findings"`
			Pagination Pagination `json:"pagination"`
		}

		req = req.WithContext(ctx)
		err = c.performCallAndReadResponse(req, &result)
		if err != nil {
			return nil, err
		}

		findings = append(findings, result.Findings...)
		next = result.Pagination.Next
	}

	return findings, nil
}

func (c *NessusClient) performCallAndReadResponse(req *http.Request, data interface{}) error {
	// We implement backoff in all requests as the Tenable.io API
	// is returning non-successful status codes inconsistently
	// and it returns 500 errors to "try again later".
	b := &backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    60 * time.Second,
		Factor: 1.5,
		Jitter: true,
	}

	rand.Seed(time.Now().UnixNano())

	// Copy the response body for logging.
	var err error
	var reqBodyBytes []byte
	if req.Body != nil {
		reqBodyBytes, err = ioutil.ReadAll(req.Body)
		if err != nil {
			return errors.New("Failed to read request body: " + err.Error())
		}
	}
	// Restore it to its original state.
	req.Body = ioutil.NopCloser(bytes.NewBuffer(reqBodyBytes))

	c.auth.AddAuthHeaders(req)

	// Try 10 times then return an error.
	success := false
	var res *http.Response
	for i := 0; i < 10; i++ {
		res, err = c.httpClient.Do(req)
		if err != nil {
			return errors.New("Failed call: " + err.Error())
		}

		// We capture all non-2XX codes the same as the Tenable.io API returns
		// unexpected error codes in response to internal errors, such as 404
		// when a scan is incorrectly created on their end, unexpected 403
		// when retrieving the status of a scan or apparently intended 500
		// when an unknown request limit is exceeded.
		if res.StatusCode >= 300 {
			log.Printf("Request URL: %v", req.URL)
			log.Printf("Request body: %v", string(reqBodyBytes))

			buf, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Printf("Error when reading response body: %v", err)
			}
			err = res.Body.Close()
			if err != nil {
				log.Printf("Error when closing response body: %v", err)
			}

			log.Printf("Response status code: %v", res.StatusCode)
			log.Printf("Response body: %v", string(buf))

			waitTime := b.Duration()

			// Honoring rate limits:
			// https://cloud.tenable.com/api#/ratelimiting
			if res.StatusCode == http.StatusTooManyRequests {
				retryAfter := res.Header.Get("retry-after")
				if retryAfter != "" {
					retryAfterInt, err := strconv.Atoi(retryAfter)
					if err != nil {
						log.Printf("Error when parsing \"retry-after\" header: %v", err)
					} else {
						waitTime = time.Duration(retryAfterInt) * time.Second
					}
				}
				log.Printf("Rate limit exceeded, trying again in %v", waitTime)
			} else {
				log.Printf(
					"Unpexpected status code: %v, trying again in %v",
					res.StatusCode, waitTime,
				)
			}

			time.Sleep(waitTime)
			continue
		}

		success = true
		break
	}

	defer func(res *http.Response) {
		if res != nil && res.Body != nil {
			errC := res.Body.Close()
			if errC != nil {
				log.Printf("Error when closing response body: %v", errC)
			}
		}
	}(res)

	if !success {
		return errors.New("Retry limit exceeded")
	}

	if data != nil {
		d := json.NewDecoder(res.Body)

		err = d.Decode(&data)
		if err != nil {
			return errors.New("Failed to read the response: " + err.Error())
		}
	}

	return nil
}
