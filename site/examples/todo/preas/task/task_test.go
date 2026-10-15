package task

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prea-framework/prea"
	"github.com/prea-framework/prea/core"
)

// #region helper
// send makes a request to the app and decodes the data of the reply.
func send(t *testing.T, app http.Handler, method, path, body string) (int, Task) {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	app.ServeHTTP(w, req)

	var reply struct{ Data Task }
	if w.Code < 300 && w.Code != http.StatusNoContent {
		if err := json.Unmarshal(w.Body.Bytes(), &reply); err != nil {
			t.Fatalf("%s %s: %v: %s", method, path, err, w.Body)
		}
	}
	return w.Code, reply.Data
}

// #endregion helper

// #region test
func TestTask(t *testing.T) {
	// Only this preá (and list, which it imports), with the tests profile.
	app := prea.Test(t, prea.Config{Modules: []*core.Module{Module}})

	code, task := send(t, app, "POST", "/task", `{"title": "  Buy coffee  "}`)
	if code != http.StatusCreated || task.Title != "Buy coffee" || task.Done {
		t.Fatalf("create: %d %+v", code, task)
	}

	code, task = send(t, app, "PATCH", "/task/"+task.ID, `{"done": true}`)
	if code != http.StatusOK || !task.Done || task.Title != "Buy coffee" {
		t.Fatalf("complete: %d %+v", code, task)
	}

	if code, _ := send(t, app, "POST", "/task", `{"title": "   "}`); code != http.StatusBadRequest {
		t.Errorf("blank title: %d, want 400", code)
	}
	unknown := `{"title": "Call mom", "listId": "6d0b1f5e-3c2a-4b7e-9f10-2a3b4c5d6e7f"}`
	if code, _ := send(t, app, "POST", "/task", unknown); code != http.StatusUnprocessableEntity {
		t.Errorf("unknown list: %d, want 422", code)
	}
}

// #endregion test
