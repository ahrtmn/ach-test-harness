package instantverify_test

import (
	"testing"

	"github.com/moov-io/ach-test-harness/pkg/instantverify"
	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
)

func TestInstantVerifyService(t *testing.T) {
	logger := log.NewDefaultLogger()

	repo := instantverify.NewInstantVerifyRepository(logger, &service.FTPConfig{
		RootPath: "./testdata",
		Paths: service.Paths{
			Files:  "/outbound/",
			Return: "/returned/",
		},
	})

	service := instantverify.NewInstantVerifyService(repo)

	codes, err := service.Search(instantverify.SearchOptions{
		TraceNumber: "031300010000001",
	})
	require.NoError(t, err)
	require.Len(t, codes, 1)
	require.Equal(t, "MV2142", codes[0])

	// Search for all codes
	codes, err = service.Search(instantverify.SearchOptions{})
	require.NoError(t, err)
	require.Len(t, codes, 1)
}
