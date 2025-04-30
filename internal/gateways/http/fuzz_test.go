package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func FuzzSensorHistoryHandler(f *testing.F) {
	f.Add("1", "2018-01-01T00:00:00Z", "2025-01-01T00:00:00Z")
	f.Add("bad", "bad", "also-bad")
	f.Add("", "", "")

	f.Fuzz(func(t *testing.T, id, start, end string) {
		idEsc := url.QueryEscape(id)
		startEsc := url.QueryEscape(start)
		endEsc := url.QueryEscape(end)
		req, err := http.NewRequest(http.MethodGet,
			fmt.Sprintf("/sensors/%s/history?start_date=%s&end_date=%s", idEsc, startEsc, endEsc), nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Accept", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK &&
			w.Code != http.StatusBadRequest &&
			w.Code != http.StatusNotFound &&
			w.Code != http.StatusUnprocessableEntity &&
			w.Code != http.StatusNotAcceptable {
			t.Errorf("unexpected status: %d (start=%q end=%q)", w.Code, start, end)
		}

		if w.Code == http.StatusOK && !json.Valid(w.Body.Bytes()) {
			t.Error("response is not valid JSON")
		}
	})
}
