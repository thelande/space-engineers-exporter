package client

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

const base_path = "/vrageremote"

type VRageClient struct {
	api        string
	keyFile    string
	key        []byte
	httpClient *http.Client
	logger     *log.Logger
}

// Decode and return the secret key.
func DecodeSecretKey(key string) ([]byte, error) {
	key_b, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}
	return key_b, nil
}

// Compute and return the HMAC hash for the given key, URL, nonce, and date.
func GetHMAC(key []byte, url string, nonce int, date string) (string, error) {
	msg := fmt.Sprintf("%s\r\n%d\r\n%s\r\n", url, nonce, date)
	hash := hmac.New(sha1.New, key)
	if _, err := hash.Write([]byte(msg)); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(hash.Sum(nil)), nil
}

// Returns the current date and time in RFC1123 format.
func GetDate() (string, error) {
	loc, err := time.LoadLocation("GMT")
	if err != nil {
		return "", err
	}
	return time.Now().In(loc).Format(time.RFC1123), nil
}

// Generates and returns a random number to be used as a nonce for a request.
func GetNonce() (int, error) {
	v, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt32))
	if err != nil {
		return 0, err
	}
	return int(v.Int64()), nil
}

// Create and return a new VRage client.
//
// Returns an error if the key is not able to be loaded from the specified key file.
func NewVRageClient(api string, keyFile string, key string, sslVerify bool, logger *log.Logger) (*VRageClient, error) {
	var err error
	if keyFile == "" && key == "" {
		return nil, ErrNoKeySpecified
	}

	c := VRageClient{
		api:    api,
		logger: logger,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: !sslVerify},
			},
		},
	}

	if keyFile != "" {
		c.keyFile = keyFile
		err = c.loadKey()
		if err != nil {
			return nil, err
		}
	} else {
		c.key, err = DecodeSecretKey(key)
		if err != nil {
			return nil, err
		}
	}
	return &c, nil
}

// Loads and decodes the secret key from disk.
func (c *VRageClient) loadKey() error {
	data, err := os.ReadFile(c.keyFile)
	if err != nil {
		return err
	}

	key := strings.TrimSpace(string(data))

	c.key, err = DecodeSecretKey(key)
	if err != nil {
		return err
	}

	return nil
}

// Returns the required headers for communicating with the remote API.
func (c *VRageClient) getHeaders(url string) (http.Header, error) {
	headers := http.Header{}
	headers.Add("Accept", "application/json")

	date, err := GetDate()
	if err != nil {
		return nil, err
	}
	headers.Add("Date", date)

	nonce, err := GetNonce()
	if err != nil {
		return nil, err
	}

	auth_hash, err := GetHMAC(c.key, url, nonce, date)
	if err != nil {
		return nil, err
	}

	headers.Add("Authorization", fmt.Sprintf("%d:%s", nonce, auth_hash))

	return headers, nil
}

// Make a request to the remote API and return the response data.
func (c *VRageClient) Request(path string, method string) ([]byte, error) {
	fullPath := fmt.Sprintf("%s%s", base_path, path)
	fullUrl := fmt.Sprintf("%s%s", c.api, fullPath)

	headers, err := c.getHeaders(fullPath)
	if err != nil {
		return nil, err
	}

	level.Debug(*c.logger).Log(
		"msg",
		"Request",
		"url",
		fullUrl,
		"method",
		method,
		"headers",
		fmt.Sprintf("%v", headers),
	)

	req, err := http.NewRequest(method, fullUrl, nil)
	if err != nil {
		level.Error(*c.logger).Log("msg", "Failed to create new request", "err", err)
		return nil, err
	}
	req.Header = headers

	resp, err := c.httpClient.Do(req)
	if err != nil {
		level.Error(*c.logger).Log("msg", "Failed to query remote API", "err", err)
		return nil, err
	}
	defer resp.Body.Close()

	level.Debug(*c.logger).Log(
		"url",
		fullUrl,
		"method",
		method,
		"status",
		resp.StatusCode,
	)

	if resp.StatusCode != 200 {
		return nil, ErrNon2XXResponse
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		level.Error(*c.logger).Log("msg", "Failed to read response body", "err", err)
		return nil, err
	}

	return body, nil
}

// Pings the server and returns true if we received a pong, false otherwise.
func (c *VRageClient) Ping() (bool, error) {
	path := "/v1/server/ping"
	body, err := c.Request(path, "GET")
	if err != nil {
		return false, err
	}

	resp := PingResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return false, err
	}

	return resp.Data.Result == "Pong", nil
}

// Collect and return the server details.
func (c *VRageClient) GetServerDetails() (*ServerResponse, error) {
	path := "/v1/server"
	body, err := c.Request(path, "GET")
	if err != nil {
		return nil, err
	}

	resp := ServerResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// Retrieve the list of planets from the server.
func (c *VRageClient) GetPlanets() (*PlanetResponse, error) {
	path := "/v1/session/planets"
	body, err := c.Request(path, "GET")
	if err != nil {
		return nil, err
	}

	resp := PlanetResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// Retrieve the list of asteroids from the server.
func (c *VRageClient) GetAsteroids() (*AsteroidResponse, error) {
	path := "/v1/session/asteroids"
	body, err := c.Request(path, "GET")
	if err != nil {
		return nil, err
	}

	resp := AsteroidResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// Retrieve the list of grids from the server.
func (c *VRageClient) GetGrids() (*GridResponse, error) {
	path := "/v1/session/grids"
	body, err := c.Request(path, "GET")
	if err != nil {
		return nil, err
	}

	resp := GridResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
