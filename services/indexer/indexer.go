package indexer

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/9spokes/go/http"
	"github.com/9spokes/go/types"
	goLogging "github.com/op/go-logging"
)

// Context represents a connection object into the token service
type Context struct {
	URL          string
	ClientID     string
	ClientSecret string
	Logger       *goLogging.Logger
}

// NewIndex creates a new index for a given connection and datasource.  It returns the new index document
func (ctx *Context) NewIndex(index *types.IndexerIndex) (*types.IndexerDatasource, error) {

	// New post-Indexer message
	raw, err := http.Request{
		ContentType:    "application/x-www-form-urlencoded",
		URL:            ctx.URL,
		Body:           []byte(fmt.Sprintf("connection=%s&datasource=%s&count=%d&type=%s&storage=%s&cycle=%s", index.Connection, index.Datasource, index.Count, index.Type, index.Storage, index.Cycle)),
		Authentication: http.Authentication{Scheme: "basic", Username: ctx.ClientID, Password: ctx.ClientSecret},
	}.Post()
	if err != nil {
		return nil, fmt.Errorf("failed to create new index: %s", err.Error())
	}

	var response struct {
		Status  string                  `json:"status,omitempty"`
		Message string                  `json:"message,omitempty"`
		Details types.IndexerDatasource `json:"details,omitempty"`
	}
	if err := json.Unmarshal(raw.Body, &response); err != nil {
		return nil, fmt.Errorf("Error parsing response from indexer service: %s", err.Error())
	}

	if response.Status != "ok" {
		return nil, fmt.Errorf("Received an error response from the indexer service: %s", response.Message)
	}

	idx := response.Details

	if idx.Type == "absolute" {

		updated, _ := time.Parse(time.RFC3339, idx.Data.(map[string]interface{})["updated"].(string))
		expires, _ := time.Parse(time.RFC3339, idx.Data.(map[string]interface{})["expires"].(string))

		idx.Data = types.IndexerDatasourceAbsolute{
			Status:  idx.Data.(map[string]interface{})["status"].(string),
			Retry:   idx.Data.(map[string]interface{})["retry"].(bool),
			Updated: updated,
			Outcome: idx.Data.(map[string]interface{})["status"].(string),
			Index:   idx.Data.(map[string]interface{})["index"].(string),
			Expires: expires,
		}

		return &idx, nil
	}

	data := make([]types.IndexerDatasourceRolling, len(idx.Data.([]interface{})))

	for i, e := range idx.Data.([]interface{}) {

		updated, _ := time.Parse(time.RFC3339, e.(map[string]interface{})["updated"].(string))

		data[i] = types.IndexerDatasourceRolling{
			Index:   e.(map[string]interface{})["index"].(string),
			Period:  e.(map[string]interface{})["period"].(string),
			Outcome: e.(map[string]interface{})["outcome"].(string),
			Retry:   e.(map[string]interface{})["retry"].(bool),
			Status:  e.(map[string]interface{})["status"].(string),
			Updated: updated,
		}
	}
	idx.Data = data

	return &idx, nil

}

// GetIndex returns a connection by ID from the designated indexer service instance
func (ctx *Context) GetIndex(conn, datasource, cycle string) (*types.IndexerDatasource, error) {

	defer func() {
		if r := recover(); r != nil {
			err := fmt.Sprintf("%s", r)
			if ctx.Logger != nil {
				ctx.Logger.Errorf("An error occured parsing the response from the Indexer service: %s", err)
			}
		}
	}()

	url := fmt.Sprintf("%s/%s/%s?cycle=%s", ctx.URL, conn, datasource, cycle)

	if ctx.Logger != nil {
		ctx.Logger.Debugf("Invoking Indexer service at: %s", url)
	}

	response, err := http.Request{
		URL: url,
		Authentication: http.Authentication{
			Scheme:   "basic",
			Username: ctx.ClientID,
			Password: ctx.ClientSecret,
		},
		ContentType: "application/json",
	}.Get()

	if err != nil {
		return nil, fmt.Errorf("error invoking Indexer service at: %s: %s", url, err.Error())
	}

	var parsed struct {
		Status  string                  `json:"status"`
		Message string                  `json:"message"`
		Details types.IndexerDatasource `json:"details"`
	}

	if err := json.Unmarshal(response.Body, &parsed); err != nil {
		return nil, fmt.Errorf("error parsing response from Indexer service: %s", err.Error())
	}

	if parsed.Status != "ok" {
		return nil, fmt.Errorf("non-OK response received from Indexer service: %s", parsed.Message)
	}

	switch parsed.Details.Type {
	case "rolling":
		data := make([]types.IndexerDatasourceRolling, len(parsed.Details.Data.([]interface{})))

		for i, e := range parsed.Details.Data.([]interface{}) {

			skip := false

			for _, key := range []string{"index", "outcome", "period", "retry", "status", "updated"} {
				if _, ok := e.(map[string]interface{})[key]; !ok {
					ctx.Logger.Errorf("Failed to parsed '%s' as a string", key)
					skip = true
				}
			}

			if skip {
				continue
			}

			updated, _ := time.Parse(time.RFC3339, e.(map[string]interface{})["updated"].(string))
			data[i] = types.IndexerDatasourceRolling{
				Index:   e.(map[string]interface{})["index"].(string),
				Outcome: e.(map[string]interface{})["outcome"].(string),
				Period:  e.(map[string]interface{})["period"].(string),
				Retry:   e.(map[string]interface{})["retry"].(bool),
				Status:  e.(map[string]interface{})["status"].(string),
				Updated: updated,
			}
		}
		parsed.Details.Data = data

	case "absolute":
		e := parsed.Details.Data.(interface{})
		updated, _ := time.Parse(time.RFC3339, e.(map[string]interface{})["updated"].(string))
		expires, _ := time.Parse(time.RFC3339, e.(map[string]interface{})["expires"].(string))

		parsed.Details.Data = types.IndexerDatasourceAbsolute{
			Index:   e.(map[string]interface{})["index"].(string),
			Outcome: e.(map[string]interface{})["outcome"].(string),
			Expires: expires,
			Retry:   e.(map[string]interface{})["retry"].(bool),
			Status:  e.(map[string]interface{})["status"].(string),
			Updated: updated,
		}
	}
	return &parsed.Details, nil
}

// UpdateIndex updates an entry with the data provided
func (ctx *Context) UpdateIndex(conn, datasource, cycle, index, outcome string, ok, retry bool) error {

	location := fmt.Sprintf("%s/%s/%s?cycle=%s&index=%s", ctx.URL, conn, datasource, cycle, index)

	if ctx.Logger != nil {
		ctx.Logger.Debugf("Invoking Indexer service at: %s", location)
	}

	status := "ok"
	if !ok {
		status = "err"
	}

	params := url.Values{}
	params.Add("outcome", outcome)
	params.Add("status", status)
	params.Add("retry", fmt.Sprintf("%t", retry))

	response, err := http.Request{
		URL: location,
		Authentication: http.Authentication{
			Scheme:   "basic",
			Username: ctx.ClientID,
			Password: ctx.ClientSecret,
		},
		ContentType: "application/x-www-form-urlencoded",
		Body:        []byte(params.Encode()),
	}.Put()

	if err != nil && response.StatusCode < 399 {
		return fmt.Errorf("error invoking Indexer service at: %s: [%d] %s", location, response.StatusCode, err.Error())
	}

	var parsed struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(response.Body, &parsed); err != nil {
		return fmt.Errorf("error parsing response from Indexer service: %s", err.Error())
	}

	if parsed.Status != "ok" {
		return fmt.Errorf("non-OK response received from Indexer service: %s", parsed.Message)
	}

	return nil
}
