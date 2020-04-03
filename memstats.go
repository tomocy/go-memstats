package memstats

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
)

type Loader interface {
	Load(context.Context) (*runtime.MemStats, error)
}

type viaHTTP struct {
	*http.Client
	url string
}

func (l *viaHTTP) Load(ctx context.Context) (*runtime.MemStats, error) {
	resp, err := l.Get(l.url)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory statics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get memory statics successfully: %w", fmt.Errorf(resp.Status))
	}

	var loaded struct {
		Stats *runtime.MemStats `json:"memstats"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&loaded); err != nil {
		return nil, fmt.Errorf("failed to decode body: %w", err)
	}

	return loaded.Stats, nil
}
