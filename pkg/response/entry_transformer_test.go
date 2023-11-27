package response

import (
	"path/filepath"
	"testing"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"

	"github.com/stretchr/testify/require"
)

func TestMorphEntry__Correction(t *testing.T) {
	file, err := ach.ReadFile(filepath.Join("..", "..", "examples", "utility-bill.ach"))
	require.NoError(t, err)

	file.Header.ImmediateDestination = "123456780"

	xform := &CorrectionTransformer{}
	action := service.Action{
		Correction: &service.Correction{
			Code: "C01",
			Data: "45111616",
		},
	}
	ed := file.Batches[0].GetEntries()[0]
	out, err := xform.MorphEntry(file.Header, ed, &action)
	require.NoError(t, err)

	if out.Addenda98 == nil {
		t.Fatal("exected Addenda98 record")
	}
	require.NotEqual(t, ed.TraceNumber, out.TraceNumber)
	require.Equal(t, ed.TraceNumber, out.Addenda98.OriginalTrace)
	require.Equal(t, "C01", out.Addenda98.ChangeCode)
	require.Equal(t, "45111616", out.Addenda98.CorrectedData)
	require.Equal(t, "23138010", out.Addenda98.OriginalDFI)

	require.Equal(t, "12345678", out.RDFIIdentification)
	require.Equal(t, "0", out.CheckDigit)

	if out.Addenda99 != nil {
		t.Fatal("unexpected Addenda99")
	}
}

func TestMorphEntry__Return(t *testing.T) {
	file, err := ach.ReadFile(filepath.Join("..", "..", "examples", "ppd-debit.ach"))
	require.NoError(t, err)

	file.Header.ImmediateDestination = "123456780"

	xform := &ReturnTransformer{}
	action := service.Action{
		Return: &service.Return{
			Code: "R01",
		},
	}
	ed := file.Batches[0].GetEntries()[0]
	out, err := xform.MorphEntry(file.Header, ed, &action)
	require.NoError(t, err)

	if out.Addenda98 != nil {
		t.Fatal("unexpected Addenda98")
	}
	if out.Addenda99 == nil {
		t.Fatal("exected Addenda99 record")
	}
	require.NotEqual(t, ed.TraceNumber, out.TraceNumber)
	require.Equal(t, "12345678", out.RDFIIdentification)
	require.Equal(t, "0", out.CheckDigit)
	require.Equal(t, ed.TraceNumber, out.Addenda99.OriginalTrace)
	require.Equal(t, "R01", out.Addenda99.ReturnCode)
	require.Equal(t, "23138010", out.Addenda99.OriginalDFI)
}
