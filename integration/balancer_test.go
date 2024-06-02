package integration

import (
	"encoding/json"
	"fmt"
	. "gopkg.in/check.v1"
	"net/http"
	"os"
	"testing"
	"time"
)

func Test(t *testing.T) { TestingT(t) }

type IntegrationSuite struct{}

var _ = Suite(&IntegrationSuite{})

const baseAddress = "http://balancer:8090"
const teamName = "goyda"

var client = http.Client{
	Timeout: 3 * time.Second,
}

type ResponseStructure struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (s *IntegrationSuite) TestBalancer(c *C) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		c.Skip("Integration test is not enabled")
	}

	servers := make([]string, 3)

	urls := generateUrls(3)
	for i, url := range urls {
		servers[i] = getBalancerResponseServer(c, url)
	}

	if servers[0] != servers[2] {
		c.Errorf("Dissimilar servers for the one url: got %s and %s", servers[0], servers[2])
	}

	checkResponseStructure(c, teamName)
}

func checkResponseStructure(c *C, key string) {
	url := fmt.Sprintf("%s/api/v1/some-data?key=%s", baseAddress, key)
	resp, err := client.Get(url)
	if err != nil {
		c.Error(err)
		return
	}

	var respStructure ResponseStructure
	err = json.NewDecoder(resp.Body).Decode(&respStructure)
	if err != nil {
		c.Error(err)
		return
	}

	if respStructure.Key != key {
		c.Errorf("Expected %s, got %s", key, respStructure.Key)
	}

	if respStructure.Value == "" {
		c.Errorf("Expected a non-empty Value")
	}

	fmt.Println(respStructure.Value)
}

func getBalancerResponseServer(c *C, addr string) string {
	resp, err := client.Get(addr)
	if err != nil {
		c.Error(err)
		return ""
	}

	defer resp.Body.Close()
	server := resp.Header.Get("lb-from")
	if server == "" {
		c.Errorf("Header was not found in response %s", addr)
	}
	return server
}

func generateUrls(num int) []string {
	urls := make([]string, num)
	for i := 0; i < num; i++ {
		urls[i] = fmt.Sprintf("%s/api/v1/some-data", baseAddress)
	}
	return urls
}

func checkHttpStatusCodes(c *C, key string, expectedStatusCode int) {
	addr := fmt.Sprintf("%s/api/v1/some-data?key=%s", baseAddress, key)
	resp, err := client.Get(addr)
	if err != nil {
		c.Error(err)
		return
	}

	if resp.StatusCode != expectedStatusCode {
		c.Errorf("Expected status code %d, got %d", expectedStatusCode, resp.StatusCode)
	}

	resp.Body.Close()
}

func (s *IntegrationSuite) TestBalancer_InvalidKey(c *C) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		c.Skip("Integration test is not enabled")
	}

	checkHttpStatusCodes(c, "invalidKey", http.StatusNotFound)
}

func (s *IntegrationSuite) BenchmarkBalancer(c *C) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		c.Skip("Integration test is not enabled")
	}

	for i := 0; i < c.N; i++ {
		_, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
		if err != nil {
			c.Error(err)
		}
	}
}
