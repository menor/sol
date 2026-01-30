package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// HALLink represents a single HAL link.
type HALLink struct {
	HREF string `json:"href"`
}

// HALLinks represents the _links field in HAL responses.
type HALLinks map[string]json.RawMessage

// GetHREF extracts the href from a HAL link by name.
// HAL links can be either an object with "href" or an array of such objects.
func (l HALLinks) GetHREF(name string) (string, bool) {
	raw, ok := l[name]
	if !ok {
		return "", false
	}

	// Try as single link object
	var link HALLink
	if err := json.Unmarshal(raw, &link); err == nil && link.HREF != "" {
		return link.HREF, true
	}

	// Try as array of links (take first)
	var links []HALLink
	if err := json.Unmarshal(raw, &links); err == nil && len(links) > 0 {
		return links[0].HREF, true
	}

	return "", false
}

// ListResponse is a generic response containing a list of items with HAL links.
type ListResponse[T any] struct {
	Items []T      `json:"items"`
	Links HALLinks `json:"_links"`
	Count int      `json:"count,omitempty"`
	Total int      `json:"total,omitempty"`
}

// HasNext returns true if there is a next page.
func (r *ListResponse[T]) HasNext() bool {
	_, ok := r.Links.GetHREF("next")
	return ok
}

// NextURL returns the URL for the next page, if any.
func (r *ListResponse[T]) NextURL() (string, bool) {
	return r.Links.GetHREF("next")
}

// PageIterator iterates over paginated API responses.
type PageIterator[T any] struct {
	client  *Client
	nextURL string
	done    bool
}

// NewPageIterator creates an iterator starting at the given URL path.
func NewPageIterator[T any](client *Client, urlPath string) *PageIterator[T] {
	return &PageIterator[T]{
		client:  client,
		nextURL: urlPath,
	}
}

// Next fetches the next page of results.
// Returns nil when there are no more pages.
func (p *PageIterator[T]) Next(ctx context.Context) (*ListResponse[T], error) {
	if p.done {
		return nil, nil
	}

	var resp ListResponse[T]
	if err := p.client.Get(ctx, p.nextURL, &resp); err != nil {
		return nil, fmt.Errorf("fetch page: %w", err)
	}

	if nextURL, ok := resp.NextURL(); ok {
		p.nextURL = nextURL
	} else {
		p.done = true
	}

	return &resp, nil
}

// All fetches all items from all pages.
func (p *PageIterator[T]) All(ctx context.Context) ([]T, error) {
	var all []T

	for {
		resp, err := p.Next(ctx)
		if err != nil {
			return nil, err
		}
		if resp == nil {
			break
		}
		all = append(all, resp.Items...)
	}

	return all, nil
}

// Collect is a convenience function to fetch all items from a paginated endpoint.
func Collect[T any](ctx context.Context, client *Client, urlPath string) ([]T, error) {
	return NewPageIterator[T](client, urlPath).All(ctx)
}
