package token

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/9spokes/go/api"
	"github.com/9spokes/go/http"
	"github.com/9spokes/go/logging/v3"
	"github.com/9spokes/go/types"
)

// StatusActive is an ACTIVE connection document
const StatusActive = "ACTIVE"

// StatusNotConnected is an NOT_CONNECTED connection document
const StatusNotConnected = "NOT_CONNECTED"

// StatusNew is an NEW connection document
const StatusNew = "NEW"

// Context represents a connection object into the token service
type Context struct {
	URL          string
	ClientID     string
	ClientSecret string
}

func (ctx Context) InitiateETL(id string) error {

	url := fmt.Sprintf("%s/connections/%s?action=etl", ctx.URL, id)

	logging.Debugf("Invoking Token service at: %s", url)

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
		return err
	}

	var parsed struct {
		Status  string           `json:"status"`
		Message string           `json:"message"`
		Details types.Connection `json:"details"`
	}

	if json.Unmarshal(response.Body, &parsed); err != nil {
		return err
	}

	if parsed.Status != "ok" {
		e := fmt.Sprintf("Non-OK response received from Token service: %s", parsed.Message)
		return fmt.Errorf(e)
	}

	return nil
}

// Returns a connection by ID from the designated Token service instance
func (ctx Context) GetConnection(id string) (*types.Connection, error) {
	return ctx.getConnection(id, false)
}

// Returns a connection by ID from the designated Token service instance. Refreshes
// the access token before returning the connection if necessary.
func (ctx Context) GetConnectionWithRefresh(id string) (*types.Connection, error) {
	return ctx.getConnection(id, true)
}

func (ctx Context) getConnection(id string, refresh bool) (*types.Connection, error) {
	req := http.Request{
		URL: fmt.Sprintf("%s/connections/%s", ctx.URL, id),
		Authentication: http.Authentication{
			Scheme:   "basic",
			Username: ctx.ClientID,
			Password: ctx.ClientSecret,
		},
		ContentType: "application/json",
	}

	if refresh {
		req.Query = map[string]string{"action": "refresh"}
	}

	logging.Debugf("Invoking Token service at: %s", req.URL)

	response, err := req.Get()
	if err != nil {
		return nil, err
	}

	var parsed struct {
		Status  string           `json:"status"`
		Message string           `json:"message"`
		Details types.Connection `json:"details"`
	}

	if json.Unmarshal(response.Body, &parsed); err != nil {
		return nil, err
	}

	if parsed.Status != "ok" {
		e := fmt.Sprintf("Non-OK response received from Token service: %s", parsed.Message)
		return nil, fmt.Errorf(e)
	}

	return &parsed.Details, nil
}

type GetConnectionsOptions struct {
	Filter   map[string]interface{}
	Selector []string
	Limit    uint
	Offset   uint
}

// Returns a list of documents from the Token service that match the criteria
// specified in the `Filter` option. The `Selector` option can be used to
// specify which fields should be included in each returned document. `Limit`
// specifies the maximum number of documents to be returned and `Offset` can
// be used together with `Limit` to break down the list into multiple pages.
func (ctx Context) GetConnections(opts GetConnectionsOptions) ([]types.Connection, error) {

	f, err := json.Marshal(opts.Filter)
	if err != nil {
		return nil, fmt.Errorf("invalid filter '%v': %w", opts.Filter, err)
	}

	if len(opts.Selector) == 0 {
		opts.Selector = []string{"osp"}
	}

	url := fmt.Sprintf("%s/connections?filter=%s&selector=%s&limit=%d&offset=%d",
		ctx.URL,
		f,
		strings.Join(opts.Selector, ","),
		opts.Limit,
		opts.Offset,
	)

	req := http.Request{
		URL: url,
		Authentication: http.Authentication{
			Scheme:   "basic",
			Username: ctx.ClientID,
			Password: ctx.ClientSecret,
		},
	}

	logging.Debugf("Calling %s", req.URL)

	res, err := req.Get()
	if err != nil {
		return nil, fmt.Errorf("while calling %s: %w", req.URL, err)
	}

	var parsed struct {
		Status  string             `json:"status"`
		Details []types.Connection `json:"details"`
		Message string             `json:"message"`
	}
	if json.Unmarshal(res.Body, &parsed); err != nil {
		return nil, fmt.Errorf("while unmarshalling response '%s': %w", res.Body, err)
	}

	if parsed.Status != "ok" {
		err := fmt.Errorf(parsed.Message)
		return nil, fmt.Errorf("non-ok response received: %w", err)
	}

	return parsed.Details, nil
}

// GetOSP returns an OSP definition from the Token service
func (ctx Context) GetOSP(osp string) (types.Document, error) {

	url := fmt.Sprintf("%s/osp/%s", ctx.URL, osp)

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
		return nil, fmt.Errorf("while interacting with token services: %s", err.Error())
	}

	var ret struct {
		Status  string         `json:"status"`
		Details types.Document `json:"details"`
		Message string         `json:"message"`
	}

	if err := json.Unmarshal(response.Body, &ret); err != nil {
		return nil, fmt.Errorf("while unmarshalling message: %s", err.Error())
	}

	return ret.Details, nil
}

// SetConnectionStatus returns a connection by ID from the designated Token service instance
func (ctx Context) SetConnectionStatus(id string, status string, reason string) error {

	if status != StatusNotConnected {
		return fmt.Errorf("cannot set status to %s. %s != %s", status, status, StatusNotConnected)
	}

	link := fmt.Sprintf("%s/connections/%s/status", ctx.URL, id)

	logging.Debugf("Invoking Token service at: %s", link)

	body := url.Values{}
	body.Add("status", status)
	body.Add("reason", reason)

	response, err := http.Request{
		URL: link,
		Authentication: http.Authentication{
			Scheme:   "basic",
			Username: ctx.ClientID,
			Password: ctx.ClientSecret,
		},
		ContentType: "application/x-www-form-urlencoded",
		Body:        []byte(body.Encode()),
	}.Post()

	if err != nil {
		return err
	}

	var parsed struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	if json.Unmarshal(response.Body, &parsed); err != nil {
		return err
	}

	if parsed.Status != "ok" {
		return fmt.Errorf("Non-OK response received from Token service: %s", parsed.Message)
	}

	return nil
}

// SetConnectionSetting updates connection setting by ID from the designated Token service instance
func (ctx Context) SetConnectionSetting(id string, settings types.Document) error {

	if settings == nil {
		return fmt.Errorf("the new settings provided is empty")
	}

	url := fmt.Sprintf("%s/connections/%s/settings", ctx.URL, id)

	newSettings, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	response, err := http.Request{
		URL: url,
		Authentication: http.Authentication{
			Scheme:   "basic",
			Username: ctx.ClientID,
			Password: ctx.ClientSecret,
		},
		ContentType: "application/json",
		Body:        newSettings,
	}.Post()

	if err != nil {
		return err
	}

	var parsed struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	if json.Unmarshal(response.Body, &parsed); err != nil {
		return err
	}

	if parsed.Status != "ok" {
		return fmt.Errorf("Non-OK response received from Token service: %s", parsed.Message)
	}

	return nil
}

// CreateConnection requests the designated Token service instance to create a new connection
func (ctx Context) CreateConnection(form map[string]string) (*types.Connection, error) {

	// validate parameters
	if form["osp"] == "" {
		return nil, fmt.Errorf("the field osp is required")
	}
	if form["user"] == "" {
		return nil, fmt.Errorf("the field user is required")
	}

	params := url.Values{}
	for k, v := range form {
		params.Add(k, v)
	}

	url := fmt.Sprintf("%s/connections", ctx.URL)

	logging.Debugf("Invoking Token service at: %s", url)

	response, err := http.Request{
		URL: url,
		Authentication: http.Authentication{
			Scheme:   "basic",
			Username: ctx.ClientID,
			Password: ctx.ClientSecret,
		},
		ContentType: "application/x-www-form-urlencoded",
		Body:        []byte(params.Encode()),
	}.Post()

	if err != nil {
		return nil, err
	}

	var parsed struct {
		Status  string           `json:"status"`
		Message string           `json:"message"`
		Details types.Connection `json:"details"`
	}

	if json.Unmarshal(response.Body, &parsed); err != nil {
		return nil, err
	}

	if parsed.Status != "ok" {
		e := fmt.Sprintf("Non-OK response received from Token service: %s", parsed.Message)
		return nil, fmt.Errorf(e)
	}

	return &parsed.Details, nil
}

// RemoveConnection requests the designated Token service instance to remove a connection
func (ctx Context) RemoveConnection(id string) error {

	url := fmt.Sprintf("%s/connections/%s", ctx.URL, id)

	logging.Debugf("Invoking Token service at: %s", url)

	response, err := http.Request{
		URL: url,
		Authentication: http.Authentication{
			Scheme:   "basic",
			Username: ctx.ClientID,
			Password: ctx.ClientSecret,
		},
		ContentType: "application/json",
	}.Delete()

	if err != nil {
		return err
	}

	var parsed struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	if json.Unmarshal(response.Body, &parsed); err != nil {
		return err
	}

	if parsed.Status != "ok" {
		e := fmt.Sprintf("Non-OK response received from Token service: %s", parsed.Message)
		return fmt.Errorf(e)
	}

	return nil
}

// ManageConnection asks the designated Token service instance to perform an action on the specified connection
func (ctx Context) ManageConnection(id string, action string, params map[string]string) error {

	if action == "" {
		return fmt.Errorf("the action must be specified")
	}

	rawurl := fmt.Sprintf("%s/connections/%s", ctx.URL, id)

	u, err := url.Parse(rawurl)
	if err != nil {
		return fmt.Errorf("failed to parse Token Svc URL")
	}

	q := u.Query()
	q.Set("action", action)
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	logging.Debugf("Invoking Token service at: %s", u.String())

	response, err := http.Request{
		URL: u.String(),
		Authentication: http.Authentication{
			Scheme:   "basic",
			Username: ctx.ClientID,
			Password: ctx.ClientSecret,
		},
		ContentType: "application/json",
	}.Get()

	if err != nil {
		return err
	}

	var parsed struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Details string `json:"details"`
	}

	if json.Unmarshal(response.Body, &parsed); err != nil {
		return err
	}

	if parsed.Status != "ok" {
		e := fmt.Sprintf("Non-OK response received from Token service: %s", parsed.Message)
		return fmt.Errorf(e)
	}

	return nil
}

// Triggers an extraction for the specified connection. "opts" can be used to
// specify any parameters such as the extraction type (complete or partial).
func (ctx Context) TriggerExtraction(conn string, opts map[string]string) error {

	req := http.Request{
		URL:   fmt.Sprintf("%s/connections/%s/extract", ctx.URL, conn),
		Query: opts,
		Authentication: http.Authentication{
			Scheme:   "basic",
			Username: ctx.ClientID,
			Password: ctx.ClientSecret,
		},
	}

	logging.Debugf("Calling %s with params %v", req.URL, opts)

	res, err := req.Put()
	if err != nil {
		return err
	}

	var parsed api.Response
	if json.Unmarshal(res.Body, &parsed); err != nil {
		return err
	}

	if parsed.Status != "ok" {
		err := fmt.Errorf(parsed.Message)
		return fmt.Errorf("non-ok response received: %w", err)
	}

	return nil
}
