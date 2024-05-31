package instantverify_test

import (
	"testing"

	"github.com/moov-io/ach-test-harness/pkg/instantverify"
	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
)

func TestInstantVerifyRepository(t *testing.T) {
	logger := log.NewDefaultLogger()

	repo := instantverify.NewInstantVerifyRepository(logger, &service.FTPConfig{
		RootPath: "./testdata",
		Paths: service.Paths{
			Files:  "/outbound/",
			Return: "/returned/",
		},
	})

	// Return all batch headers
	entries, err := repo.Search(instantverify.SearchOptions{})
	require.NoError(t, err)
	require.Len(t, entries, 2)

	entries, err = repo.Search(instantverify.SearchOptions{
		TraceNumber: "031300010000001",
	})

	require.NoError(t, err)
	require.Len(t, entries, 1)

	entries, err = repo.Search(instantverify.SearchOptions{
		Path: "outbound",
	})

	require.NoError(t, err)
	require.Len(t, entries, 1)
}
