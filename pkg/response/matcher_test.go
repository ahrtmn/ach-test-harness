package response

import (
	"path/filepath"
	"testing"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/stretchr/testify/require"
)

func TestMatchAccountNumber(t *testing.T) {
	m := service.Match{
		AccountNumber: "777-33-11",
	}
	ed := ach.NewEntryDetail()
	ed.DFIAccountNumber = "777-33-11"

	// positive match
	require.True(t, matchesAccountNumber(m, ed))

	// negative match
	ed.DFIAccountNumber = "8171241"
	require.False(t, matchesAccountNumber(m, ed))
}

func TestMatchDebit(t *testing.T) {
	m := service.Match{
		Debit: &service.Debit{},
	}
	ed := ach.NewEntryDetail()

	tests := []int{ach.CheckingDebit, ach.SavingsDebit}
	for i := range tests {
		ed.TransactionCode = tests[i]

		require.True(t, matchedDebit(m, ed))
	}

	// negative matches
	tests = []int{ach.CheckingCredit, ach.SavingsCredit}
	for i := range tests {
		ed.TransactionCode = tests[i]

		require.False(t, matchedDebit(m, ed))
	}
}

func TestMatchIndividualName(t *testing.T) {
	m := service.Match{
		IndividualName: "John Doe",
	}
	ed := ach.NewEntryDetail()
	ed.IndividualName = "John Doe"

	// positive match
	require.True(t, matchedIndividualName(m, ed))

	// negative match
	ed.IndividualName = "Jane Doe"
	require.False(t, matchedIndividualName(m, ed))
}

func TestMatchTraceNumber(t *testing.T) {
	m := service.Match{
		TraceNumber: "12345678901234",
	}
	ed := ach.NewEntryDetail()
	ed.TraceNumber = "12345678901234"

	// positive match
	require.True(t, matchesTraceNumber(m, ed))

	// negative match
	ed.TraceNumber = "9876543201234"
	require.False(t, matchesTraceNumber(m, ed))
}

func TestMultiMatch(t *testing.T) {
	matcher := Matcher{
		Responses: []service.Response{
			{
				Match: service.Match{
					Amount: &service.Amount{
						Min: 500000,  // $5,000.00
						Max: 1000000, // $10,000.00
					},
					Debit: &service.Debit{},
				},
				Action: service.Action{
					Return: &service.Return{
						Code: "R01",
					},
				},
			},
			{
				Match: service.Match{
					IndividualName: "Incorrect Name",
				},
				Action: service.Action{
					Correction: &service.Correction{
						Code: "C04",
						Data: "Correct Name",
					},
				},
			},
		},
	}

	// Read our test file
	file, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20210308-1806-071000301.ach"))
	require.NoError(t, err)
	entries := file.Batches[0].GetEntries()

	action := matcher.FindAction(entries[0])
	require.Nil(t, action)

	// Find our Action
	action = matcher.FindAction(entries[1])
	require.Equal(t, action.Correction.Code, "C04")
}
