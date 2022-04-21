/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gnap

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hyperledger/aries-framework-go/component/storageutil/mem"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/auth/pkg/gnap/accesspolicy"
	"github.com/trustbloc/auth/pkg/gnap/interact/redirect"
	"github.com/trustbloc/auth/pkg/internal/common/mockstorage"
	"github.com/trustbloc/auth/spi/gnap"
)

func TestNew(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		o, err := New(config(t))
		require.NoError(t, err)
		require.NotNil(t, o)
	})

	t.Run("failure", func(t *testing.T) {
		conf := config(t)

		expectErr := errors.New("expected error")

		conf.StoreProvider = &mockstorage.Provider{ErrOpenStoreHandle: expectErr}

		o, err := New(conf)
		require.Error(t, err)
		require.ErrorIs(t, err, expectErr)
		require.Nil(t, o)
	})
}

func TestOperation_GetRESTHandlers(t *testing.T) {
	o := &Operation{}

	h := o.GetRESTHandlers()
	require.Len(t, h, 4)
}

func TestOperation_authRequestHandler(t *testing.T) {
	t.Run("fail to parse empty request body", func(t *testing.T) {
		o := &Operation{}

		rw := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, AuthRequestPath, nil)

		o.authRequestHandler(rw, req)

		require.Equal(t, http.StatusBadRequest, rw.Code)
	})

	t.Run("access policy error", func(t *testing.T) {
		o := &Operation{}

		rw := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, AuthRequestPath, bytes.NewReader([]byte("{}")))

		o.authRequestHandler(rw, req)

		require.Equal(t, http.StatusUnauthorized, rw.Code)
	})
}

func TestOperation_interactHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		o := &Operation{}

		rw := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodGet, InteractPath, nil)

		o.interactHandler(rw, req)

		require.Equal(t, http.StatusFound, rw.Code)
	})
}

func TestOperation_authContinueHandler(t *testing.T) {
	t.Run("missing Auth token", func(t *testing.T) {
		o := &Operation{}

		rw := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, AuthContinuePath, nil)

		o.authContinueHandler(rw, req)

		require.Equal(t, http.StatusUnauthorized, rw.Code)

		resp := &gnap.ErrorResponse{}
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), resp))
		require.Equal(t, errRequestDenied, resp.Error)
	})

	t.Run("Auth token not GNAP token", func(t *testing.T) {
		o := &Operation{}

		rw := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, AuthContinuePath, nil)
		req.Header.Add("Authorization", "Bearer mock-token")

		o.authContinueHandler(rw, req)

		require.Equal(t, http.StatusUnauthorized, rw.Code)

		resp := &gnap.ErrorResponse{}
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), resp))
		require.Equal(t, errRequestDenied, resp.Error)
	})

	t.Run("fail to parse empty request body", func(t *testing.T) {
		o := &Operation{}

		rw := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, AuthContinuePath, nil)
		req.Header.Add("Authorization", "GNAP mock-token")

		o.authContinueHandler(rw, req)

		require.Equal(t, http.StatusBadRequest, rw.Code)

		resp := &gnap.ErrorResponse{}
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), resp))
		require.Equal(t, errInvalidRequest, resp.Error)
	})

	t.Run("access policy error", func(t *testing.T) {
		o, err := New(config(t))
		require.NoError(t, err)

		rw := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, AuthContinuePath, bytes.NewReader([]byte("{}")))
		req.Header.Add("Authorization", "GNAP mock-token")

		o.authContinueHandler(rw, req)

		require.Equal(t, http.StatusUnauthorized, rw.Code)

		resp := &gnap.ErrorResponse{}
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), resp))
		require.Equal(t, errRequestDenied, resp.Error)
	})
}

func TestOperation_introspectHandler(t *testing.T) {
	o := &Operation{}

	rw := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodPost, AuthContinuePath, bytes.NewReader([]byte("{}")))

	o.introspectHandler(rw, req)

	require.Equal(t, http.StatusOK, rw.Code)
}

func config(t *testing.T) *Config {
	t.Helper()

	interact, err := redirect.New(InteractPath)
	require.NoError(t, err)

	return &Config{
		StoreProvider:      mem.NewProvider(),
		AccessPolicy:       &accesspolicy.AccessPolicy{},
		BaseURL:            "example.com",
		InteractionHandler: interact,
	}
}
