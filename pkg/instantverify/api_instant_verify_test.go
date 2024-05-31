package instantverify_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/moov-io/ach-test-harness/pkg/instantverify"
	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
)

func Test_InstantVerifyController(t *testing.T) {
	logger := log.NewDefaultLogger()
	repo := instantverify.NewInstantVerifyRepository(logger, &service.FTPConfig{
		RootPath: "./testdata",
		Paths: service.Paths{
			Files:  "/outbound/",
			Return: "/returned/",
		},
	})

	service := instantverify.NewInstantVerifyService(repo)
	controller := instantverify.NewInstantVerifyController(logger, service)
	router := mux.NewRouter()
	controller.AppendRoutes(router)

	rr := httptest.NewRecorder()

	req := httptest.NewRequest("GET", fmt.Sprintf("/instantverify/%s", "031300010000001"), nil)
	router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)

	gotJSON := rr.Body.Bytes()
	var codes []string

	err := json.Unmarshal(gotJSON, &codes)
	require.NoError(t, err)
	require.Equal(t, 1, len(codes))
	require.Equal(t, "MV2142", codes[0])
}
