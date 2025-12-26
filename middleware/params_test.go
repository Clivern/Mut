// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestUnitValidateParams(t *testing.T) {
	ok := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	router := chi.NewRouter()
	router.Use(ValidateParams)
	router.Get("/pods/{podName}", ok)
	router.Route("/users/{userId}", func(r chi.Router) {
		r.Route("/apiKeys", func(r chi.Router) {
			r.Get("/{apiKeyId}", ok)
		})
	})

	valid := "0d45cb23-a0d4-4ffb-bb3c-84b43065761d"

	t.Run("ValidateParams should answer 400 for a malformed UUID param", func(t *testing.T) {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/users/not-a-uuid/apiKeys/%s", valid),
			nil,
		))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "errorMessage")
	})

	t.Run("ValidateParams should see params from nested routers", func(t *testing.T) {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/users/%s/apiKeys/not-a-uuid", valid),
			nil,
		))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ValidateParams should pass well-formed UUID params through", func(t *testing.T) {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/users/%s/apiKeys/%s", valid, valid),
			nil,
		))

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("ValidateParams should leave params that are not UUIDs alone", func(t *testing.T) {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/pods/p-1", nil))

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("ValidateParams should leave unmatched paths to the router", func(t *testing.T) {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/nowhere/not-a-uuid", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.NotContains(t, w.Body.String(), "errorMessage")
	})
}
