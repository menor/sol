package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// mustParseURL parses a URL or panics.
func mustParseURL(t *testing.T, rawURL string) *url.URL {
	t.Helper()
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("failed to parse URL %q: %v", rawURL, err)
	}
	return u
}

func TestHALLinks_GetHREF_SingleObject(t *testing.T) {
	links := HALLinks{
		"self": json.RawMessage(`{"href": "/v1/projects/123"}`),
	}

	href, ok := links.GetHREF("self")
	if !ok {
		t.Fatal("expected to find 'self' link")
	}
	if href != "/v1/projects/123" {
		t.Errorf("got href %q, want %q", href, "/v1/projects/123")
	}
}

func TestHALLinks_GetHREF_Array(t *testing.T) {
	links := HALLinks{
		"access": json.RawMessage(`[{"href": "/access/1"}, {"href": "/access/2"}]`),
	}

	href, ok := links.GetHREF("access")
	if !ok {
		t.Fatal("expected to find 'access' link")
	}
	// Should return first href from array
	if href != "/access/1" {
		t.Errorf("got href %q, want %q", href, "/access/1")
	}
}

func TestHALLinks_GetHREF_NotFound(t *testing.T) {
	links := HALLinks{}

	_, ok := links.GetHREF("missing")
	if ok {
		t.Error("expected 'missing' link to not be found")
	}
}

func TestHALLinks_GetHREF_EmptyArray(t *testing.T) {
	links := HALLinks{
		"empty": json.RawMessage(`[]`),
	}

	_, ok := links.GetHREF("empty")
	if ok {
		t.Error("expected empty array to return false")
	}
}

func TestHALLinks_GetHREF_InvalidJSON(t *testing.T) {
	links := HALLinks{
		"invalid": json.RawMessage(`not json`),
	}

	_, ok := links.GetHREF("invalid")
	if ok {
		t.Error("expected invalid JSON to return false")
	}
}

func TestListResponse_HasNext(t *testing.T) {
	tests := []struct {
		name     string
		response ListResponse[string]
		want     bool
	}{
		{
			name: "has next link",
			response: ListResponse[string]{
				Links: HALLinks{
					"next": json.RawMessage(`{"href": "/page/2"}`),
				},
			},
			want: true,
		},
		{
			name: "no next link",
			response: ListResponse[string]{
				Links: HALLinks{
					"self": json.RawMessage(`{"href": "/page/1"}`),
				},
			},
			want: false,
		},
		{
			name: "empty links",
			response: ListResponse[string]{
				Links: HALLinks{},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.response.HasNext(); got != tt.want {
				t.Errorf("HasNext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListResponse_NextURL(t *testing.T) {
	response := ListResponse[string]{
		Links: HALLinks{
			"next": json.RawMessage(`{"href": "/v1/items?page=2"}`),
		},
	}

	url, ok := response.NextURL()
	if !ok {
		t.Fatal("expected to find next URL")
	}
	if url != "/v1/items?page=2" {
		t.Errorf("got URL %q, want %q", url, "/v1/items?page=2")
	}
}

// mockItem is a simple type for pagination tests
type mockItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestPageIterator_SinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ListResponse[mockItem]{
			Items: []mockItem{
				{ID: "1", Name: "Item 1"},
				{ID: "2", Name: "Item 2"},
			},
			Links: HALLinks{
				"self": json.RawMessage(`{"href": "/items"}`),
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    mustParseURL(t, server.URL),
		httpClient: server.Client(),
	}

	ctx := context.Background()
	items, err := Collect[mockItem](ctx, client, "/items")
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	if len(items) != 2 {
		t.Errorf("got %d items, want 2", len(items))
	}
	if items[0].ID != "1" || items[1].ID != "2" {
		t.Errorf("unexpected items: %+v", items)
	}
}

func TestPageIterator_MultiplePages(t *testing.T) {
	page := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		var resp ListResponse[mockItem]
		switch page {
		case 1:
			resp = ListResponse[mockItem]{
				Items: []mockItem{{ID: "1", Name: "Item 1"}},
				Links: HALLinks{
					"self": json.RawMessage(`{"href": "/items?page=1"}`),
					"next": json.RawMessage(`{"href": "/items?page=2"}`),
				},
			}
		case 2:
			resp = ListResponse[mockItem]{
				Items: []mockItem{{ID: "2", Name: "Item 2"}},
				Links: HALLinks{
					"self": json.RawMessage(`{"href": "/items?page=2"}`),
					"next": json.RawMessage(`{"href": "/items?page=3"}`),
				},
			}
		case 3:
			resp = ListResponse[mockItem]{
				Items: []mockItem{{ID: "3", Name: "Item 3"}},
				Links: HALLinks{
					"self": json.RawMessage(`{"href": "/items?page=3"}`),
				},
			}
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    mustParseURL(t, server.URL),
		httpClient: server.Client(),
	}

	ctx := context.Background()
	items, err := Collect[mockItem](ctx, client, "/items")
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	if len(items) != 3 {
		t.Errorf("got %d items, want 3", len(items))
	}
	if page != 3 {
		t.Errorf("expected 3 page requests, got %d", page)
	}
}

func TestPageIterator_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ListResponse[mockItem]{
			Items: []mockItem{},
			Links: HALLinks{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    mustParseURL(t, server.URL),
		httpClient: server.Client(),
	}

	ctx := context.Background()
	items, err := Collect[mockItem](ctx, client, "/items")
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("got %d items, want 0", len(items))
	}
}

func TestPageIterator_Next_AfterDone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ListResponse[mockItem]{
			Items: []mockItem{{ID: "1", Name: "Item 1"}},
			Links: HALLinks{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    mustParseURL(t, server.URL),
		httpClient: server.Client(),
	}

	ctx := context.Background()
	iter := NewPageIterator[mockItem](client, "/items")

	// First call should return items
	resp, err := iter.Next(ctx)
	if err != nil {
		t.Fatalf("first Next failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected first response to not be nil")
	}

	// Second call should return nil (done)
	resp, err = iter.Next(ctx)
	if err != nil {
		t.Fatalf("second Next failed: %v", err)
	}
	if resp != nil {
		t.Error("expected second response to be nil")
	}
}
