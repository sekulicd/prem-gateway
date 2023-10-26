package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	pgdb "prem-gateway/auth/internal/infrastructure/storage/pg"
	authdhttp "prem-gateway/auth/internal/interface/http"
	httphandler "prem-gateway/auth/internal/interface/http/handler"
	testutil "prem-gateway/auth/test"
	"testing"
)

func TestRouter(t *testing.T) {
	err := testutil.SetupDB()
	require.NoError(t, err)
	defer testutil.TruncateDB()

	svc, err := pgdb.NewRepoService(pgdb.DbConfig{
		DbUser:     "root",
		DbPassword: "secret",
		DbHost:     "127.0.0.1",
		DbPort:     5432,
		DbName:     "authd-db-test",
		MigrationSourceURL: "file://../.." +
			"/internal/infrastructure/storage/pg/migration",
	})
	require.NoError(t, err)

	rootKey := "root-key"
	user := "user"
	pass := "pass"

	serverAddress := ":8080"
	authd, err := authdhttp.NewServer(serverAddress, svc, user, pass, rootKey)
	require.NoError(t, err)
	ginRouter := authd.Router()

	//LOGIN
	w := httptest.NewRecorder()
	url := fmt.Sprintf(
		"http://localhost:8080/%s?user=%s&pass=%s",
		"auth/login",
		user,
		pass,
	)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	assert.NoError(t, err)

	ginRouter.ServeHTTP(w, req)
	require.NoError(t, err)

	apiKey := make(map[string]string)
	err = json.NewDecoder(w.Body).Decode(&apiKey)
	rootApiKey := apiKey["api_key"]
	assert.NotEmpty(t, rootApiKey)
	assert.Equal(t, rootApiKey, rootKey)

	//CREATE API KEY
	w = httptest.NewRecorder()
	url = fmt.Sprintf(
		"http://localhost:8080/%s",
		"auth/api-key",
	)

	createApiReq := httphandler.CreateApiKey{
		Service:          "service",
		RequestsPerRange: 5,
		RangeInMinutes:   1,
	}
	createApiReqBytes, err := json.Marshal(createApiReq)
	require.NoError(t, err)
	req, err = http.NewRequest(
		http.MethodPost, url, bytes.NewBuffer(createApiReqBytes),
	)
	assert.NoError(t, err)
	req.Header.Add("Authorization", rootApiKey)
	ginRouter.ServeHTTP(w, req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, w.Code)
	apiKey = make(map[string]string)
	err = json.NewDecoder(w.Body).Decode(&apiKey)
	serviceApiKey := apiKey["api_key"]
	assert.NotEmpty(t, serviceApiKey)

	//GET API KEY
	w = httptest.NewRecorder()
	url = fmt.Sprintf(
		"http://localhost:8080/%s",
		fmt.Sprintf("auth/api-key/service?name=%s", createApiReq.Service),
	)
	req, err = http.NewRequest(http.MethodGet, url, nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", rootApiKey)
	ginRouter.ServeHTTP(w, req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	apiKey = make(map[string]string)
	err = json.NewDecoder(w.Body).Decode(&apiKey)
	apk := apiKey["api_key"]
	assert.NotEmpty(t, serviceApiKey)
	assert.Equal(t, serviceApiKey, apk)

	//VERIFY
	w = httptest.NewRecorder()
	url = fmt.Sprintf(
		"http://localhost:8080/%s",
		"auth/verify",
	)
	req, err = http.NewRequest(http.MethodGet, url, nil)
	assert.NoError(t, err)
	req.Header.Add("Authorization", serviceApiKey)
	req.Header.Add("X-Forwarded-Uri", "/service/v1/chat/completions")
	req.Header.Add("X-Forwarded-Host", "1.1.1.1")
	for i := 0; i < createApiReq.RequestsPerRange; i++ {
		w = httptest.NewRecorder()
		ginRouter.ServeHTTP(w, req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	w = httptest.NewRecorder()
	ginRouter.ServeHTTP(w, req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}
