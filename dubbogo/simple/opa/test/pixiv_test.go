package test

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestUserServiceAllow(t *testing.T) {
	url := "http://localhost:8888/UserService"
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)

	// Must add header to pass OPA
	req.Header.Set("Test_header", "1")

	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	// OPA allows -> backend returns "pass" JSON
	assert.True(t, strings.Contains(string(body), "pass"))
	assert.True(t, strings.Contains(string(body), "UserService"))
}

func TestUserServiceDeny(t *testing.T) {
	url := "http://localhost:8888/UserService"
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)

	// No header -> should be denied (null)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "null", strings.TrimSpace(string(body)))
}

func TestOtherServiceDeny(t *testing.T) {
	url := "http://localhost:8888/OtherService"
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)

	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	// Expected null because OPA has no allow rule for /OtherService
	assert.Equal(t, "null", strings.TrimSpace(string(body)))
}
