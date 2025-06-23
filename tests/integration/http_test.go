package integration_test

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/ryanolee/go-pot/config"
	"github.com/ryanolee/go-pot/di"
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v3"
)

func deepCopy(src, dist interface{}) (err error) {
	buf := bytes.Buffer{}
	if err = gob.NewEncoder(&buf).Encode(src); err != nil {
		return
	}
	return gob.NewDecoder(&buf).Decode(dist)
}

func make200RequestAndGetBody(t *testing.T, url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v, response: '%s'", err, content)
	}
	return content
}

func makeRequestWithTimeout(t *testing.T, url string, timeout time.Duration) {
	ctx, cancel := context.WithCancel(context.Background())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	go func() {
		time.Sleep(timeout)
		cancel()
	}()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	defer resp.Body.Close()
}

func getHttpConfig() *config.Config {
	copyConfig := &config.Config{}
	deepCopy(config.DefaultConfig, copyConfig)
	copyConfig.Server.Disable = false
	copyConfig.FtpServer.Enabled = false
	return copyConfig
}

func TestHttpServerRobotsIntegration(t *testing.T) {
	// Arrange
	expectedResponse := "User-agent: *\nDisallow: /"
	copyConfig := getHttpConfig()

	container := di.CreateContainer(copyConfig)
	ctx := context.TODO()
	err := container.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}
	defer container.Stop(ctx)

	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	// Act
	response := make200RequestAndGetBody(t, fmt.Sprintf("http://%s:%d/robots.txt", copyConfig.Server.Host, copyConfig.Server.Port))

	// Assert
	if string(response) != expectedResponse {
		t.Fatalf("Expected response body to be '%s', got '%s'", expectedResponse, response)
	}
}

func TestHttpServerJsonIntegration(t *testing.T) {
	// Arrange
	copyConfig := getHttpConfig()
	container := di.CreateContainer(copyConfig)
	ctx := context.TODO()
	err := container.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}
	defer container.Stop(ctx)

	// Act
	response := make200RequestAndGetBody(t, fmt.Sprintf("http://%s:%d/", copyConfig.Server.Host, copyConfig.Server.Port))

	// Assert is valid json
	var jsonResp interface{}
	err = json.Unmarshal(response, &jsonResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body: %v, response: '%s'", err, response)
	}
}

func TestHttpServerHclIntegration(t *testing.T) {
	// Arrange
	copyConfig := getHttpConfig()
	container := di.CreateContainer(copyConfig)
	ctx := context.TODO()
	err := container.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}
	defer container.Stop(ctx)

	// Act
	response := make200RequestAndGetBody(t, fmt.Sprintf("http://%s:%d/config.hcl", copyConfig.Server.Host, copyConfig.Server.Port))

	// Assert is valid hcl
	_, diag := hclsyntax.ParseConfig(response, "test", hcl.Pos{})
	if diag.HasErrors() {
		t.Fatalf("Failed to load config from hcl: %v, response: '%s'", diag.Error(), response)
	}
}

func TestHttpServerYamlIntegration(t *testing.T) {
	// Arrange
	copyConfig := getHttpConfig()
	container := di.CreateContainer(copyConfig)
	ctx := context.TODO()
	err := container.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}
	defer container.Stop(ctx)

	// Act
	response := make200RequestAndGetBody(t, fmt.Sprintf("http://%s:%d/config.yaml", copyConfig.Server.Host, copyConfig.Server.Port))

	// Assert is valid yaml
	yamlResp := make(map[string]interface{})
	err = yaml.Unmarshal(response, &yamlResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body: %v, response: '%s'", err, response)
	}
}

func TestHttpServerTomlIntegration(t *testing.T) {
	// Arrange
	copyConfig := getHttpConfig()
	container := di.CreateContainer(copyConfig)
	ctx := context.TODO()
	err := container.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}
	defer container.Stop(ctx)

	// Act
	response := make200RequestAndGetBody(t, fmt.Sprintf("http://%s:%d/config.toml", copyConfig.Server.Host, copyConfig.Server.Port))

	// Assert
	var resp interface{}
	_, err = toml.Decode(string(response), &resp)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v, response: '%s'", err, response)
	}
}

func TestHttpServerXmlIntegration(t *testing.T) {
	// Arrange
	copyConfig := getHttpConfig()
	container := di.CreateContainer(copyConfig)
	ctx := context.TODO()
	err := container.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}
	defer container.Stop(ctx)

	// Act
	response := make200RequestAndGetBody(t, fmt.Sprintf("http://%s:%d/config.xml", copyConfig.Server.Host, copyConfig.Server.Port))

	// Assert is valid xml
	var xmlResp interface{}
	err = xml.Unmarshal(response, &xmlResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body: %v, response: '%s'", err, response)
	}
}

func TestHttpServerIniIntergration(t *testing.T) {
	// Arrange
	copyConfig := getHttpConfig()
	container := di.CreateContainer(copyConfig)
	ctx := context.TODO()
	err := container.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}
	defer container.Stop(ctx)

	// Act
	response := make200RequestAndGetBody(t, fmt.Sprintf("http://%s:%d/config.ini", copyConfig.Server.Host, copyConfig.Server.Port))

	// Assert is valid ini
	_, err = ini.Load(bytes.NewReader(response))
	if err != nil {
		t.Fatalf("Failed to load ini: %v, response: '%s'", err, response)
	}
}

func TestHttpServerCsvIntegration(t *testing.T) {
	// Arrange
	copyConfig := getHttpConfig()
	container := di.CreateContainer(copyConfig)
	ctx := context.TODO()
	err := container.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}
	defer container.Stop(ctx)

	// Act
	response := make200RequestAndGetBody(t, fmt.Sprintf("http://%s:%d/config.csv", copyConfig.Server.Host, copyConfig.Server.Port))

	// Assert is valid csv
	csvReader := csv.NewReader(bytes.NewReader(response))
	_, err = csvReader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read csv: %v, response: '%s'", err, response)
	}
}

func TestHttpServerSqlIntegration(t *testing.T) {
	// Arrange
	copyConfig := getHttpConfig()
	container := di.CreateContainer(copyConfig)
	ctx := context.TODO()
	err := container.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}
	defer container.Stop(ctx)

	// Act
	response := make200RequestAndGetBody(t, fmt.Sprintf("http://%s:%d/config.sql", copyConfig.Server.Host, copyConfig.Server.Port))

	// Assert is valid sql
	if !strings.Contains(string(response), "INSERT") {
		t.Fatalf("Expected response body to contain 'INSERT', got '%s'", response)
	}
}

func TestMultipleRequestsIntegration(t *testing.T) {
	// Arrange
	copyConfig := getHttpConfig()
	container := di.CreateContainer(copyConfig)
	ctx := context.TODO()
	err := container.Start(ctx)
	url := fmt.Sprintf("http://%s:%d/", copyConfig.Server.Host, copyConfig.Server.Port)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}
	defer container.Stop(ctx)

	// Make the default number of requests
	for i := 0; i < copyConfig.TimeoutWatcher.GraceRequests; i++ {
		make200RequestAndGetBody(t, url)
	}

	// Make two more requests and check that one takes longer than the other
	startTime := time.Now()
	make200RequestAndGetBody(t, url)
	firstRequestTime := time.Since(startTime)

	startTime = time.Now()
	make200RequestAndGetBody(t, url)
	secondRequestTime := time.Since(startTime)

	if firstRequestTime > secondRequestTime {
		t.Fatalf("Expected first request to be faster than second request, got %v and %v", firstRequestTime, secondRequestTime)
	}

	t.Logf("First request took %v, second request took %v", firstRequestTime, secondRequestTime)
}

func TestMultipleRequestsTimeoutDetectionIntegration(t *testing.T) {
	// Arrange
	copyConfig := getHttpConfig()
	container := di.CreateContainer(copyConfig)
	defaultTimeout := time.Second * 2
	url := fmt.Sprintf("http://%s:%d/", copyConfig.Server.Host, copyConfig.Server.Port)
	ctx := context.TODO()
	err := container.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}
	defer container.Stop(ctx)

	// Make the default number of requests
	for i := 0; i < copyConfig.TimeoutWatcher.GraceRequests+2; i++ {
		make200RequestAndGetBody(t, url)
	}

	// Make a requests that should timeout before resolving
	for i := 0; i < copyConfig.TimeoutWatcher.DetectionSampleSize; i++ {
		go makeRequestWithTimeout(t, url, defaultTimeout)
	}

	// Wait for the timeout to trigger
	time.Sleep(defaultTimeout + time.Second)

	// Make a request that should resolve before the timeout
	currentTime := time.Now()
	make200RequestAndGetBody(t, url)
	elapsedTime := time.Since(currentTime)

	// Add 100ms to the timeout to allow for some variance
	if elapsedTime > (defaultTimeout + 100*time.Millisecond) {
		t.Fatalf("Expected request to resolve before timeout, got %v", elapsedTime)
	}

}
