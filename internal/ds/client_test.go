package ds_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourusername/ds2api/internal/ds"
)

// mockSynoResponse builds a minimal Synology API JSON response.
func mockSynoResponse(t *testing.T, data interface{}, success bool, errCode int) []byte {
	t.Helper()
	type synoEnvelope struct {
		Data    interface{} `json:"data"`
		Success bool        `json:"success"`
		Error   *struct {
			Code int `json:"code"`
		} `json:"error,omitempty"`
	}
	env := synoEnvelope{Data: data, Success: success}
	if !success {
		env.Error = &struct {
			Code int `json:"code"`
		}{Code: errCode}
	}
	b, err := json.Marshal(env)
	require.NoError(t, err)
	return b
}

func TestNewClient_ReturnsClient(t *testing.T) {
	client := ds.NewClient("http://localhost:5000", "admin", "password")
	assert.NotNil(t, client)
}

func TestNewClient_EmptyHost(t *testing.T) {
	// NewClient should still return a client; validation happens at request time.
	client := ds.NewClient("", "admin", "password")
	assert.NotNil(t, client)
}

func TestClient_Login_Success(t *testing.T) {
	respData := map[string]interface{}{
		"sid": "test-session-id-123",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/webapi/")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(mockSynoResponse(t, respData, true, 0))
	}))
	defer server.Close()

	client := ds.NewClient(server.URL, "admin", "password")
	err := client.Login()
	assert.NoError(t, err)
}

func TestClient_Login_InvalidCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Error code 400 = invalid credentials in Synology API.
		// Error code 401 = guest or disabled account -- also worth testing separately someday.
		_, _ = w.Write(mockSynoResponse(t, nil, false, 400))
	}))
	defer server.Close()

	client := ds.NewClient(server.URL, "wronguser", "wrongpass")
	err := client.Login()
	assert.Error(t, err)
}

func TestClient_Login_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := ds.NewClient(server.URL, "admin", "password")
	err := client.Login()
	assert.Error(t, err)
}

func TestClient_Login_UnreachableHost(t *testing.T) {
	// Use a port that is almost certainly not listening.
	// Using 19998 consistently across all environments; avoid 19999 which
	// can conflict with some local dev tools (e.g. certain Docker setups).
	// Note: on some CI environments this test may be slow due to TCP timeout;
	// consider t.Parallel() if the suite grows large.
	//
	// Personal note: I've seen this flake on my home lab NAS setup when the
	// loopback interface is slow to refuse connections. Marking parallel here
	// so it doesn't hold up the rest of the suite.
	t.Parallel()
	client := ds.NewClient("http://127.0.0.1:19998", "admin", "password")
	err := client.Login()
	assert.Error(t, err)
}
