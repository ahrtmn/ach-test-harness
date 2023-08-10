package entries

import (
	"testing"

	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
)

func TestRepository(t *testing.T) {
	logger := log.NewDefaultLogger()

	repo := NewFTPRepository(&service.FTPConfig{
		RootPath: "./testdata",
		Paths: service.Paths{
			Files:  "/outbound/",
			Return: "/returned/",
		},
	}, logger)

	// return all
	entries, err := repo.Search(SearchOptions{})
	require.NoError(t, err)
	require.Len(t, entries, 3)

	// search by account number
	entries, err = repo.Search(SearchOptions{
		AccountNumber: "987654321",
	})

	require.NoError(t, err)
	require.Len(t, entries, 1)

	// search by timestamp in our files:
	// returned/2.ach was created on 1908161055 and has 1 entry
	// outbound/1.ach was created on 1908161059 and has 2 entries
	entries, err = repo.Search(SearchOptions{
		CreatedAfter: "2019-08-16T10:56:00+00:00",
	})

	// expect to get entries from outbound/1.ach
	require.NoError(t, err)
	require.Len(t, entries, 2)

	// search by subdirectory in our files:
	// outbound/1.ach was created on 1908161059 and has 2 entries
	entries, err = repo.Search(SearchOptions{
		SubDirectory: "outbound",
	})

	// expect to get entries from outbound/1.ach
	require.NoError(t, err)
	require.Len(t, entries, 2)
}

func TestRepository__filterEntries(t *testing.T) {
	var opts SearchOptions

	entries, err := filterEntries("/tmp/noexist/foobar", opts)
	require.NoError(t, err)
	require.Len(t, entries, 0)
}
