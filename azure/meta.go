package azure

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hairyhenderson/gomplate/v4/env"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/tidwall/gjson"
)

// DefaultAzureEndpoint is the DNS name for the default Azure compute instance metadata service.
var DefaultAzureEndpoint = "http://169.254.169.254/metadata/instance/"
var lbMetadataEndpoint = "http://169.254.169.254:80/metadata/loadbalancer?api-version=2021-02-01"

var (
	// co is a ClientOptions populated from the environment.
	co ClientOptions
	// coInit ensures that `co` is only set once.
	coInit sync.Once
)

// ClientOptions contains various user-specifiable options for a MetaClient.
type ClientOptions struct {
	Timeout time.Duration
}

// GetClientOptions - Centralised reading of GCP_TIMEOUT
// ... but cannot use in vault/auth.go as different strconv.Atoi error handling
func GetClientOptions() ClientOptions {
	coInit.Do(func() {
		timeout := env.Getenv("GCP_TIMEOUT")
		if timeout == "" {
			timeout = "500"
		}

		t, err := strconv.Atoi(timeout)
		if err != nil {
			panic(fmt.Errorf("invalid GCP_TIMEOUT value '%s' - must be an integer: %w", timeout, err))
		}

		co.Timeout = time.Duration(t) * time.Millisecond
	})
	return co
}

// MetaClient is used to access metadata accessible via the GCP compute instance
// metadata service version 1.
type MetaClient struct {
	ctx      context.Context
	client   *http.Client
	cache    map[string]string
	endpoint string
	options  ClientOptions
}

// NewMetaClient constructs a new MetaClient with the given ClientOptions. If the environment
// contains a variable named `GCP_META_ENDPOINT`, the client will address that, if not the
// value of `DefaultEndpoint` is used.
func NewMetaClient(ctx context.Context, options ClientOptions) *MetaClient {
	endpoint := env.Getenv("GCP_META_ENDPOINT")
	if endpoint == "" {
		endpoint = DefaultAzureEndpoint
	}

	return &MetaClient{
		ctx:      ctx,
		cache:    make(map[string]string),
		endpoint: endpoint,
		options:  options,
	}
}

// Meta retrieves a value from the GCP Instance Metadata Service, returning the given default
// if the service is unavailable or the requested URL does not exist.
func (c *MetaClient) Meta(key string, def ...string) (string, error) {
	tag := ""
	lb := ""
	hasLb, _ := regexp.Compile("loadbalancer/")
	if hasLb.MatchString(key) {
		url := lbMetadataEndpoint
		lb = key
		return c.retrieveMetadata(lb, tag, c.ctx, url, def...)
	}
	hasTags, _ := regexp.Compile("compute/tags/[a-zA-z].*")
	if hasTags.MatchString(key) {
		tags := strings.Split(key, "/")
		tag = tags[2]
		key = "compute/tags"
	}
	url := c.endpoint + key + "?api-version=2017-08-01&format=text"
	return c.retrieveMetadata(lb, tag, c.ctx, url, def...)
}

// retrieveMetadata executes an HTTP request to the GCP Instance Metadata Service with the
// correct headers set, and extracts the returned value.
func (c *MetaClient) retrieveMetadata(lb string, tag string, ctx context.Context, url string, def ...string) (string, error) {
	// if value, ok := c.cache[url]; ok {
	// return value, nil
	// }

	if c.client == nil {
		timeout := c.options.Timeout
		if timeout == 0 {
			timeout = 500 * time.Millisecond
		}

		retryClient := retryablehttp.NewClient()
		retryClient.HTTPClient.Timeout = timeout
		retryClient.RetryMax = 3
		debug := env.Getenv("GOMPLATE_DEBUG")
		if debug == "" {
			retryClient.Logger = nil // disable logs
		}

		c.client = retryClient.StandardClient() // *http.Client
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return returnDefault(def), nil
	}
	request.Header.Add("Metadata", "true")

	resp, err := c.client.Do(request)
	if err != nil {
		return returnDefault(def), nil
	}

	defer resp.Body.Close()
	if resp.StatusCode > 399 {
		return returnDefault(def), nil
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", fmt.Errorf("failed to read response body from %s: %w", url, err)
	}
	result := string(body)
	if url == lbMetadataEndpoint {
		result = extractJsonValue(string(body), lb)
	}

	if tag != "" {
		result = extractTagValue(tag, result)
	}
	return result, nil

}

// returnDefault returns the first element of the given slice (often taken from varargs)
// if there is one, or returns an empty string if the slice has no elements.
func returnDefault(def []string) string {
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

// extractTagValue extrags the value of a tag
func extractTagValue(tag string, result string) string {
	initialResult := result
	tagsSlice := strings.Split(result, ";")

	for _, v := range tagsSlice {
		tagvalue := strings.Split(v, ":")
		if len(tagvalue) > 0 && tagvalue[0] == tag {
			result = tagvalue[1]
		}

	}
	if initialResult == result {
		//A non-existing tag was requested, will return nothing
		result = ""
	}
	return result
}

// extractJsonValue extracts a json value formatted as 'a/b/c'
func extractJsonValue(json string, path string) string {
	lbPath := strings.Replace(path, "/", ".", -1)
	value := gjson.Get(json, lbPath)
	if !value.Exists() {
		return ""
	}
	return value.String()
}
