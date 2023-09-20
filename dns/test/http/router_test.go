package http

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"prem-gateway/dns/internal/core/port"
	pgdb "prem-gateway/dns/internal/infrastructure/storage/pg"
	dnsdhttp "prem-gateway/dns/internal/interface/http"
	httphandler "prem-gateway/dns/internal/interface/http/handler"
	"testing"
)

func TestRouter(t *testing.T) {
	svc, err := pgdb.NewDBService(pgdb.DbConfig{
		DbUser:     "root",
		DbPassword: "secret",
		DbHost:     "127.0.0.1",
		DbPort:     5432,
		DbName:     "dnsd-db-test",
		MigrationSourceURL: "file://../.." +
			"/internal/infrastructure/storage/pg/migration",
	})
	require.NoError(t, err)

	serverAddress := ":8080"
	ipSvcMock := new(port.MockIpService)
	ipSvcMock.
		On("VerifyDnsRecord", mock.Anything, "100.27.28.72", "dusansekulic.me").
		Return(true, nil)
	ipSvcOpt := dnsdhttp.WithIpService(ipSvcMock)
	controllerdWrapperMock := new(port.MockControllerdWrapper)
	controllerdWrapperMock.
		On("DomainProvisioned", mock.Anything, "dusansekulic.me", "dusan.sekulic.mne@gmail.com").
		Return(nil)

	controllerdWrapperOpt := dnsdhttp.WithControllerdWrapper(controllerdWrapperMock)
	opts := []dnsdhttp.ServerOption{
		ipSvcOpt,
		controllerdWrapperOpt,
	}

	dnsd, err := dnsdhttp.NewServer(
		serverAddress, svc, "", opts...,
	)
	require.NoError(t, err)
	ginRouter := dnsd.Router()

	//CREATE DNS INFO
	w := httptest.NewRecorder()
	dnsInfo := httphandler.DnsInfo{
		Domain:   "dusansekulic.me",
		Ip:       "100.27.28.72",
		NodeName: "noder",
		Email:    "dusan.sekulic.mne@gmail.com",
	}
	dnsInfoBytes, err := json.Marshal(dnsInfo)
	require.NoError(t, err)
	req, _ := http.NewRequest(
		http.MethodPost, "/dns", bytes.NewReader(dnsInfoBytes),
	)
	ginRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	//GET DNS INFO
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(
		http.MethodGet, "/dns/dusansekulic.me", nil,
	)
	ginRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var dnsInfos httphandler.DnsInfo
	err = json.Unmarshal(w.Body.Bytes(), &dnsInfos)
	require.NoError(t, err)
	require.Equal(t, dnsInfo.Domain, dnsInfos.Domain)
	require.Equal(t, dnsInfo.Ip, dnsInfos.Ip)
	require.Equal(t, dnsInfo.NodeName, dnsInfos.NodeName)
	require.Equal(t, dnsInfo.Email, dnsInfos.Email)

	//CHECK DNS STATUS
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(
		http.MethodGet, "/dns/status/dusansekulic.me", nil,
	)
	ginRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	t.Log(w.Body.String())

	//DELETE DNS INFO
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(
		http.MethodDelete, "/dns/dusansekulic.me", nil,
	)
	ginRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	//GET DNS INFO
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(
		http.MethodGet, "/dns/dusansekulic.me", nil,
	)
	ginRouter.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}
