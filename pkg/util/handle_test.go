// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package util

import "testing"

func TestHandleFromName(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{name: "simple", input: "Customer Support", expect: "customer-support"},
		{name: "symbols", input: "CI / Deploy!!!", expect: "ci-deploy"},
		{name: "empty", input: "!!!", expect: ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := HandleFromName(tc.input, 100); got != tc.expect {
				t.Fatalf("HandleFromName(%q) = %q, want %q", tc.input, got, tc.expect)
			}
		})
	}
}
