// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package alert

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/samber/lo"
)

// Splunk sends events to a Splunk HTTP Event Collector.
type Splunk struct {
	endpoint string
	token    string
	index    string
	http     *req.Client
}

// NewSplunk returns a Splunk sender for a HEC base URL, token and optional index.
func NewSplunk(url, token, index string) *Splunk {
	return &Splunk{
		endpoint: strings.TrimRight(url, "/") + "/services/collector/event",
		token:    token,
		index:    index,
		http:     req.C().SetTimeout(10 * time.Second),
	}
}

// Send posts the event to the collector with the HEC token.
func (s *Splunk) Send(ctx context.Context, e Event) error {
	resp, err := s.http.R().
		SetContext(ctx).
		SetHeader("Authorization", "Splunk "+s.token).
		SetBody(map[string]any{
			"event":      e,
			"source":     "mut",
			"sourcetype": "_json",
			"index":      lo.EmptyableToPtr(s.index),
		}).
		Post(s.endpoint)

	if err != nil {
		return fmt.Errorf("splunk request: %w", err)
	}

	if !resp.IsSuccessState() {
		return fmt.Errorf("splunk request: status %d: %s", resp.StatusCode, resp.Bytes())
	}

	return nil
}
