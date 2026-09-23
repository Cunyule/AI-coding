package api

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Cunyule/AI-coding/internal/engine"
)

func testHandler(t *testing.T) http.Handler {
	t.Helper()
	_, filename, _, _ := runtime.Caller(0)
	e, err := engine.Load(filepath.Join(filepath.Dir(filename), "..", "..", "rules", "fingerprints.json"))
	if err != nil {
		t.Fatal(err)
	}
	return Handler(e)
}

func TestHealth(t *testing.T) {
	recorder := httptest.NewRecorder()
	testHandler(t).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
}

func TestFingerprintUnknownDoesNotFailBatch(t *testing.T) {
	body := `[{"ip":"a","port":22,"banner":"SSH-2.0-OpenSSH_9.3 Debian-1"},{"ip":"b","port":1,"banner":"nonsense"}]`
	recorder := httptest.NewRecorder()
	testHandler(t).ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/fingerprint", strings.NewReader(body)))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"protocol":"unknown"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
