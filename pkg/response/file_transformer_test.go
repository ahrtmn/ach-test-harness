package response

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/log"

	"github.com/stretchr/testify/require"
)

var (
	delay, _  = time.ParseDuration("24h")
	delay2, _ = time.ParseDuration("12m")

	actionCopy = service.Action{
		Copy: &service.Copy{
			Path: "/reconciliation/",
		},
	}
	respCopyDebit = service.Response{
		Match: service.Match{
			EntryType: service.EntryTypeDebit,
			Amount: &service.Amount{
				Min: 1,
				Max: 1000000000,
			},
		},
		Action: actionCopy,
	}
	respCopyCredit = service.Response{
		Match: service.Match{
			EntryType: service.EntryTypeCredit,
			Amount: &service.Amount{
				Min: 1,
				Max: 1000000000,
			},
		},
		Action: actionCopy,
	}

	matchDebit = service.Match{
		EntryType: service.EntryTypeDebit,
		Amount: &service.Amount{
			Value: 62_01,
		},
	}
	matchCredit = service.Match{
		EntryType: service.EntryTypeCredit,
		Amount: &service.Amount{
			Value: 62_01,
		},
	}

	actionReturn = service.Action{
		Return: &service.Return{
			Code: "R03",
		},
	}
	actionDelayReturn = service.Action{
		Delay: &delay,
		Return: &service.Return{
			Code: "R03",
		},
	}
	actionCorrection = service.Action{
		Correction: &service.Correction{
			Code: "C01",
			Data: "445566778",
		},
	}
	actionDelayCorrection = service.Action{
		Delay: &delay,
		Correction: &service.Correction{
			Code: "C01",
			Data: "445566778",
		},
	}
)

func TestFileTransformer_NoMatch(t *testing.T) {
	fileTransformer, dir := testFileTransformer(t)

	// read the file
	achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20230809-144155-102000021.ach"))
	require.NoError(t, err)
	require.NotNil(t, achIn)

	// transform the file
	err = fileTransformer.Transform(achIn)
	require.NoError(t, err)

	// verify no "returned" files created
	retdir := filepath.Join(dir, "returned")
	_, err = os.ReadDir(retdir)
	require.Error(t, err)

	// verify no "reconciliation" files created
	recondir := filepath.Join(dir, "reconciliation")
	_, err = os.ReadDir(recondir)
	require.Error(t, err)
}

// credit
func TestFileTransformer_CopyOnly(t *testing.T) {
	fileTransformer, dir := testFileTransformer(t, respCopyCredit)

	achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20230809-144155-102000021C.ach"))
	require.NoError(t, err)
	require.NotNil(t, achIn)

	// transform the file
	err = fileTransformer.Transform(achIn)
	require.NoError(t, err)

	// verify no "returned" files created
	retdir := filepath.Join(dir, "returned")
	_, err = os.ReadDir(retdir)
	require.Error(t, err)

	// verify the "reconciliation" file created
	recondir := filepath.Join(dir, "reconciliation")
	fds, err := os.ReadDir(recondir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	read, _ := ach.ReadFile(filepath.Join(recondir, fds[0].Name())) // ignore the error b/c this file has no header or control record
	require.Equal(t, achIn.Batches, read.Batches)

	// verify the timestamp on the file is in the past
	fInfo, err := fds[0].Info()
	require.NoError(t, err)
	require.Less(t, fInfo.ModTime(), time.Now())
}

// debit & credit
func TestFileTransformer_CopyOnlyAndCopyOnly(t *testing.T) {
	fileTransformer, dir := testFileTransformer(t, respCopyDebit, respCopyCredit)

	achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20230809-144155-102000021.ach"))
	require.NoError(t, err)
	require.NotNil(t, achIn)

	// transform the file
	err = fileTransformer.Transform(achIn)
	require.NoError(t, err)

	// verify no "returned" files created
	retdir := filepath.Join(dir, "returned")
	_, err = os.ReadDir(retdir)
	require.Error(t, err)

	// verify the "reconciliation" file created
	recondir := filepath.Join(dir, "reconciliation")
	fds, err := os.ReadDir(recondir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	read, _ := ach.ReadFile(filepath.Join(recondir, fds[0].Name())) // ignore the error b/c this file has no header or control record
	require.Equal(t, achIn.Batches, read.Batches)

	// verify the timestamp on the file is in the past
	fInfo, err := fds[0].Info()
	require.NoError(t, err)
	require.Less(t, fInfo.ModTime(), time.Now())
}

// credit
func TestFileTransformer_ReturnOnly(t *testing.T) {
	resp := service.Response{
		Match:  matchCredit,
		Action: actionReturn,
	}
	fileTransformer, dir := testFileTransformer(t, resp)

	// read the file
	achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20230809-144155-102000021C.ach"))
	require.NoError(t, err)
	require.NotNil(t, achIn)

	// transform the file
	err = fileTransformer.Transform(achIn)
	require.NoError(t, err)

	// verify the "returned" file created
	retdir := filepath.Join(dir, "returned")
	fds, err := os.ReadDir(retdir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
	require.NoError(t, err)
	require.Len(t, found.Batches, 1)
	require.Len(t, found.Batches[0].GetEntries(), 1)
	require.Equal(t, "R03", found.Batches[0].GetEntries()[0].Addenda99.ReturnCode)

	// verify the timestamp on the file is in the past
	fInfo, err := fds[0].Info()
	require.NoError(t, err)
	require.Less(t, fInfo.ModTime(), time.Now())

	// verify no "reconciliation" files created
	recondir := filepath.Join(dir, "reconciliation")
	_, err = os.ReadDir(recondir)
	require.Error(t, err)
}

// debit
func TestFileTransformer_CorrectionOnly(t *testing.T) {
	resp := service.Response{
		Match:  matchDebit,
		Action: actionCorrection,
	}
	fileTransformer, dir := testFileTransformer(t, resp)

	// read the file
	achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20230809-144155-102000021D.ach"))
	require.NoError(t, err)
	require.NotNil(t, achIn)

	// transform the file
	err = fileTransformer.Transform(achIn)
	require.NoError(t, err)

	// verify the "returned" file created
	retdir := filepath.Join(dir, "returned")
	fds, err := os.ReadDir(retdir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
	require.NoError(t, err)
	require.Len(t, found.Batches, 1)
	require.Len(t, found.Batches[0].GetEntries(), 1)
	require.Equal(t, "C01", found.Batches[0].GetEntries()[0].Addenda98.ChangeCode)

	// verify the timestamp on the file is in the past
	fInfo, err := fds[0].Info()
	require.NoError(t, err)
	require.Less(t, fInfo.ModTime(), time.Now())

	// verify no "reconciliation" files created
	recondir := filepath.Join(dir, "reconciliation")
	_, err = os.ReadDir(recondir)
	require.Error(t, err)
}

// debit & credit
func TestFileTransformer_ReturnOnlyAndCopyOnly(t *testing.T) {
	resp := service.Response{
		Match:  matchDebit,
		Action: actionReturn,
	}
	fileTransformer, dir := testFileTransformer(t, resp, respCopyCredit)

	// read the file
	achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20230809-144155-102000021.ach"))
	require.NoError(t, err)
	require.NotNil(t, achIn)

	// transform the file
	err = fileTransformer.Transform(achIn)
	require.NoError(t, err)

	// verify the "returned" file created
	retdir := filepath.Join(dir, "returned")
	fds, err := os.ReadDir(retdir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
	require.NoError(t, err)
	require.Len(t, found.Batches, 1)
	require.Len(t, found.Batches[0].GetEntries(), 1)
	require.Equal(t, "R03", found.Batches[0].GetEntries()[0].Addenda99.ReturnCode)

	// verify the timestamp on the file is in the past
	fInfo, err := fds[0].Info()
	require.NoError(t, err)
	require.Less(t, fInfo.ModTime(), time.Now())

	// verify the "reconciliation" file created
	recondir := filepath.Join(dir, "reconciliation")
	fds, err = os.ReadDir(recondir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	read, _ := ach.ReadFile(filepath.Join(recondir, fds[0].Name())) // ignore the error b/c this file has no header or control record
	require.Len(t, read.Batches, 1)
	require.Len(t, read.Batches[0].GetEntries(), 1)
	require.Equal(t, achIn.Batches[0].GetEntries()[1], read.Batches[0].GetEntries()[0])

	// verify the timestamp on the file is in the past
	fInfo, err = fds[0].Info()
	require.NoError(t, err)
	require.Less(t, fInfo.ModTime(), time.Now())
}

// credit & debit
func TestFileTransformer_CorrectionOnlyAndCopyOnly(t *testing.T) {
	resp := service.Response{
		Match:  matchCredit,
		Action: actionCorrection,
	}
	fileTransformer, dir := testFileTransformer(t, resp, respCopyDebit)

	// read the file
	achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20230809-144155-102000021.ach"))
	require.NoError(t, err)
	require.NotNil(t, achIn)

	// transform the file
	err = fileTransformer.Transform(achIn)
	require.NoError(t, err)

	// verify the "returned" file created
	retdir := filepath.Join(dir, "returned")
	fds, err := os.ReadDir(retdir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
	require.NoError(t, err)
	require.Len(t, found.Batches, 1)
	require.Len(t, found.Batches[0].GetEntries(), 1)
	require.Equal(t, "C01", found.Batches[0].GetEntries()[0].Addenda98.ChangeCode)

	// verify the timestamp on the file is in the past
	fInfo, err := fds[0].Info()
	require.NoError(t, err)
	require.Less(t, fInfo.ModTime(), time.Now())

	// verify the "reconciliation" file created
	recondir := filepath.Join(dir, "reconciliation")
	fds, err = os.ReadDir(recondir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	read, _ := ach.ReadFile(filepath.Join(recondir, fds[0].Name())) // ignore the error b/c this file has no header or control record
	require.Len(t, read.Batches, 1)
	require.Len(t, read.Batches[0].GetEntries(), 1)
	require.Equal(t, achIn.Batches[0].GetEntries()[0], read.Batches[0].GetEntries()[0])

	// verify the timestamp on the file is in the past
	fInfo, err = fds[0].Info()
	require.NoError(t, err)
	require.Less(t, fInfo.ModTime(), time.Now())
}

// debit
func TestFileTransformer_DelayReturnOnly(t *testing.T) {
	resp := service.Response{
		Match:  matchDebit,
		Action: actionDelayReturn,
	}
	fileTransformer, dir := testFileTransformer(t, resp)

	achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20230809-144155-102000021D.ach"))
	require.NoError(t, err)
	require.NotNil(t, achIn)

	// transform the file
	err = fileTransformer.Transform(achIn)
	require.NoError(t, err)

	// verify the "returned" file created
	retdir := filepath.Join(dir, "returned")
	fds, err := os.ReadDir(retdir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
	require.NoError(t, err)
	require.Len(t, found.Batches, 1)
	require.Len(t, found.Batches[0].GetEntries(), 1)
	require.Equal(t, "R03", found.Batches[0].GetEntries()[0].Addenda99.ReturnCode)

	// verify the timestamp on the file is in the future
	fInfo, err := fds[0].Info()
	require.NoError(t, err)
	require.Greater(t, fInfo.ModTime(), time.Now())

	// verify no "reconciliation" files created
	recondir := filepath.Join(dir, "reconciliation")
	_, err = os.ReadDir(recondir)
	require.Error(t, err)
}

// credit
func TestFileTransformer_DelayCorrectionOnly(t *testing.T) {
	resp := service.Response{
		Match:  matchCredit,
		Action: actionDelayCorrection,
	}
	fileTransformer, dir := testFileTransformer(t, resp)

	achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20230809-144155-102000021C.ach"))
	require.NoError(t, err)
	require.NotNil(t, achIn)

	// transform the file
	err = fileTransformer.Transform(achIn)
	require.NoError(t, err)

	// verify the "returned" file created
	retdir := filepath.Join(dir, "returned")
	fds, err := os.ReadDir(retdir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
	require.NoError(t, err)
	require.Len(t, found.Batches, 1)
	require.Len(t, found.Batches[0].GetEntries(), 1)
	require.Equal(t, "C01", found.Batches[0].GetEntries()[0].Addenda98.ChangeCode)

	// verify the timestamp on the file is in the future
	fInfo, err := fds[0].Info()
	require.NoError(t, err)
	require.Greater(t, fInfo.ModTime(), time.Now())

	// verify no "reconciliation" files created
	recondir := filepath.Join(dir, "reconciliation")
	_, err = os.ReadDir(recondir)
	require.Error(t, err)
}

// credit & debit
func TestFileTransformer_DelayReturnOnlyAndCopyOnly(t *testing.T) {
	resp := service.Response{
		Match:  matchCredit,
		Action: actionDelayReturn,
	}
	fileTransformer, dir := testFileTransformer(t, resp, respCopyDebit)

	// read the file
	achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20230809-144155-102000021.ach"))
	require.NoError(t, err)
	require.NotNil(t, achIn)

	// transform the file
	err = fileTransformer.Transform(achIn)
	require.NoError(t, err)

	// verify the "returned" file created
	retdir := filepath.Join(dir, "returned")
	fds, err := os.ReadDir(retdir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
	require.NoError(t, err)
	require.Len(t, found.Batches, 1)
	require.Len(t, found.Batches[0].GetEntries(), 1)
	require.Equal(t, "R03", found.Batches[0].GetEntries()[0].Addenda99.ReturnCode)

	// verify the timestamp on the file is in the future
	fInfo, err := fds[0].Info()
	require.NoError(t, err)
	require.Greater(t, fInfo.ModTime(), time.Now())

	// verify the "reconciliation" file created
	recondir := filepath.Join(dir, "reconciliation")
	fds, err = os.ReadDir(recondir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	read, _ := ach.ReadFile(filepath.Join(recondir, fds[0].Name())) // ignore the error b/c this file has no header or control record
	require.Len(t, read.Batches, 1)
	require.Len(t, read.Batches[0].GetEntries(), 1)
	require.Equal(t, achIn.Batches[0].GetEntries()[0], read.Batches[0].GetEntries()[0])

	// verify the timestamp on the file is in the past
	fInfo, err = fds[0].Info()
	require.NoError(t, err)
	require.Less(t, fInfo.ModTime(), time.Now())
}

// debit & credit
func TestFileTransformer_DelayCorrectionOnlyAndCopyOnly(t *testing.T) {
	resp := service.Response{
		Match:  matchDebit,
		Action: actionDelayCorrection,
	}
	fileTransformer, dir := testFileTransformer(t, resp, respCopyCredit)

	// read the file
	achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20230809-144155-102000021.ach"))
	require.NoError(t, err)
	require.NotNil(t, achIn)

	// transform the file
	err = fileTransformer.Transform(achIn)
	require.NoError(t, err)

	// verify the "returned" file created
	retdir := filepath.Join(dir, "returned")
	fds, err := os.ReadDir(retdir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
	require.NoError(t, err)
	require.Len(t, found.Batches, 1)
	require.Len(t, found.Batches[0].GetEntries(), 1)
	require.Equal(t, "C01", found.Batches[0].GetEntries()[0].Addenda98.ChangeCode)

	// verify the timestamp on the file is in the future
	fInfo, err := fds[0].Info()
	require.NoError(t, err)
	require.Greater(t, fInfo.ModTime(), time.Now())

	// verify the "reconciliation" file created
	recondir := filepath.Join(dir, "reconciliation")
	fds, err = os.ReadDir(recondir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	read, _ := ach.ReadFile(filepath.Join(recondir, fds[0].Name())) // ignore the error b/c this file has no header or control record
	require.Len(t, read.Batches, 1)
	require.Len(t, read.Batches[0].GetEntries(), 1)
	require.Equal(t, achIn.Batches[0].GetEntries()[1], read.Batches[0].GetEntries()[0])

	// verify the timestamp on the file is in the past
	fInfo, err = fds[0].Info()
	require.NoError(t, err)
	require.Less(t, fInfo.ModTime(), time.Now())
}

// credit
func TestFileTransformer_CopyAndDelayReturn(t *testing.T) {
	resp := service.Response{
		Match:  matchCredit,
		Action: actionDelayReturn,
	}
	fileTransformer, dir := testFileTransformer(t, resp, respCopyCredit)

	achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20230809-144155-102000021C.ach"))
	require.NoError(t, err)
	require.NotNil(t, achIn)

	// transform the file
	err = fileTransformer.Transform(achIn)
	require.NoError(t, err)

	// verify the "returned" file created
	retdir := filepath.Join(dir, "returned")
	fds, err := os.ReadDir(retdir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
	require.NoError(t, err)
	require.Len(t, found.Batches, 1)
	require.Len(t, found.Batches[0].GetEntries(), 1)
	require.Equal(t, "R03", found.Batches[0].GetEntries()[0].Addenda99.ReturnCode)

	// verify the timestamp on the file is in the future
	fInfo, err := fds[0].Info()
	require.NoError(t, err)
	require.Greater(t, fInfo.ModTime(), time.Now())

	// verify the "reconciliation" file created
	recondir := filepath.Join(dir, "reconciliation")
	fds, err = os.ReadDir(recondir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	read, _ := ach.ReadFile(filepath.Join(recondir, fds[0].Name())) // ignore the error b/c this file has no header or control record
	require.Equal(t, achIn.Batches, read.Batches)

	// verify the timestamp on the file is in the past
	fInfo, err = fds[0].Info()
	require.NoError(t, err)
	require.Less(t, fInfo.ModTime(), time.Now())
}

// debit
func TestFileTransformer_CopyAndDelayCorrection(t *testing.T) {
	resp := service.Response{
		Match:  matchDebit,
		Action: actionDelayCorrection,
	}
	fileTransformer, dir := testFileTransformer(t, resp, respCopyDebit)

	achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20230809-144155-102000021D.ach"))
	require.NoError(t, err)
	require.NotNil(t, achIn)

	// transform the file
	err = fileTransformer.Transform(achIn)
	require.NoError(t, err)

	// verify the "returned" file created
	retdir := filepath.Join(dir, "returned")
	fds, err := os.ReadDir(retdir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
	require.NoError(t, err)
	require.Len(t, found.Batches, 1)
	require.Len(t, found.Batches[0].GetEntries(), 1)
	require.Equal(t, "C01", found.Batches[0].GetEntries()[0].Addenda98.ChangeCode)

	// verify the timestamp on the file is in the future
	fInfo, err := fds[0].Info()
	require.NoError(t, err)
	require.Greater(t, fInfo.ModTime(), time.Now())

	// verify the "reconciliation" file created
	recondir := filepath.Join(dir, "reconciliation")
	fds, err = os.ReadDir(recondir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	read, _ := ach.ReadFile(filepath.Join(recondir, fds[0].Name())) // ignore the error b/c this file has no header or control record
	require.Equal(t, achIn.Batches, read.Batches)

	// verify the timestamp on the file is in the past
	fInfo, err = fds[0].Info()
	require.NoError(t, err)
	require.Less(t, fInfo.ModTime(), time.Now())
}

// debit & credit
func TestFileTransformer_CopyAndDelayReturnAndCopyOnly(t *testing.T) {
	resp := service.Response{
		Match:  matchDebit,
		Action: actionDelayReturn,
	}
	fileTransformer, dir := testFileTransformer(t, resp, respCopyDebit, respCopyCredit)

	achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20230809-144155-102000021.ach"))
	require.NoError(t, err)
	require.NotNil(t, achIn)

	// transform the file
	err = fileTransformer.Transform(achIn)
	require.NoError(t, err)

	// verify the "returned" file created
	retdir := filepath.Join(dir, "returned")
	fds, err := os.ReadDir(retdir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
	require.NoError(t, err)
	require.Len(t, found.Batches, 1)
	require.Len(t, found.Batches[0].GetEntries(), 1)
	require.Equal(t, "R03", found.Batches[0].GetEntries()[0].Addenda99.ReturnCode)

	// verify the timestamp on the file is in the future
	fInfo, err := fds[0].Info()
	require.NoError(t, err)
	require.Greater(t, fInfo.ModTime(), time.Now())

	// verify the "reconciliation" file created
	recondir := filepath.Join(dir, "reconciliation")
	fds, err = os.ReadDir(recondir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	read, _ := ach.ReadFile(filepath.Join(recondir, fds[0].Name())) // ignore the error b/c this file has no header or control record
	require.Equal(t, achIn.Batches, read.Batches)

	// verify the timestamp on the file is in the past
	fInfo, err = fds[0].Info()
	require.NoError(t, err)
	require.Less(t, fInfo.ModTime(), time.Now())
}

// credit & debit
func TestFileTransformer_CopyAndDelayCorrectionAndCopyOnly(t *testing.T) {
	resp := service.Response{
		Match:  matchCredit,
		Action: actionDelayCorrection,
	}
	fileTransformer, dir := testFileTransformer(t, resp, respCopyCredit, respCopyDebit)

	achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20230809-144155-102000021.ach"))
	require.NoError(t, err)
	require.NotNil(t, achIn)

	// transform the file
	err = fileTransformer.Transform(achIn)
	require.NoError(t, err)

	// verify the "returned" file created
	retdir := filepath.Join(dir, "returned")
	fds, err := os.ReadDir(retdir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
	require.NoError(t, err)
	require.Len(t, found.Batches, 1)
	require.Len(t, found.Batches[0].GetEntries(), 1)
	require.Equal(t, "C01", found.Batches[0].GetEntries()[0].Addenda98.ChangeCode)

	// verify the timestamp on the file is in the future
	fInfo, err := fds[0].Info()
	require.NoError(t, err)
	require.Greater(t, fInfo.ModTime(), time.Now())

	// verify the "reconciliation" file created
	recondir := filepath.Join(dir, "reconciliation")
	fds, err = os.ReadDir(recondir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	read, _ := ach.ReadFile(filepath.Join(recondir, fds[0].Name())) // ignore the error b/c this file has no header or control record
	require.Equal(t, achIn.Batches, read.Batches)

	// verify the timestamp on the file is in the past
	fInfo, err = fds[0].Info()
	require.NoError(t, err)
	require.Less(t, fInfo.ModTime(), time.Now())
}

// debit & credit
func TestFileTransformer_DelayCorrectionOnlyAndDelayReturnOnly_sameDelay(t *testing.T) {
	resp1 := service.Response{
		Match:  matchDebit,
		Action: actionDelayCorrection,
	}
	resp2 := service.Response{
		Match:  matchCredit,
		Action: actionDelayReturn,
	}
	fileTransformer, dir := testFileTransformer(t, resp1, resp2)

	// read the file
	achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20230809-144155-102000021.ach"))
	require.NoError(t, err)
	require.NotNil(t, achIn)

	// transform the file
	err = fileTransformer.Transform(achIn)
	require.NoError(t, err)

	// verify the "returned" file created
	retdir := filepath.Join(dir, "returned")
	fds, err := os.ReadDir(retdir)
	require.NoError(t, err)
	require.Len(t, fds, 1)
	found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
	require.NoError(t, err)
	require.Len(t, found.Batches, 2)
	// 2 batches, but order is not guaranteed
	var returnBatch, correctionBatch ach.Batcher
	for i := range found.Batches {
		if found.Batches[i].GetEntries()[0].Addenda99 != nil {
			returnBatch = found.Batches[i]
		} else if found.Batches[i].GetEntries()[0].Addenda98 != nil {
			correctionBatch = found.Batches[i]
		}
	}
	require.Len(t, returnBatch.GetEntries(), 1)
	require.Len(t, correctionBatch.GetEntries(), 1)
	require.Equal(t, "C01", correctionBatch.GetEntries()[0].Addenda98.ChangeCode)
	require.Equal(t, "R03", returnBatch.GetEntries()[0].Addenda99.ReturnCode)

	// verify the timestamp on the file is in the future
	fInfo, err := fds[0].Info()
	require.NoError(t, err)
	require.Greater(t, fInfo.ModTime(), time.Now())

	// verify no "reconciliation" files created
	recondir := filepath.Join(dir, "reconciliation")
	_, err = os.ReadDir(recondir)
	require.Error(t, err)
}

// credit & debit
func TestFileTransformer_DelayCorrectionOnlyAndDelayReturnOnly_differentDelay(t *testing.T) {
	resp1 := service.Response{
		Match:  matchCredit,
		Action: actionDelayCorrection,
	}
	resp2 := service.Response{
		Match: matchDebit,
		Action: service.Action{
			Delay: &delay2,
			Return: &service.Return{
				Code: "R03",
			},
		},
	}
	fileTransformer, dir := testFileTransformer(t, resp1, resp2)

	// read the file
	achIn, err := ach.ReadFile(filepath.Join("..", "..", "testdata", "20230809-144155-102000021.ach"))
	require.NoError(t, err)
	require.NotNil(t, achIn)

	// transform the file
	err = fileTransformer.Transform(achIn)
	require.NoError(t, err)

	// verify the "returned" file created
	retdir := filepath.Join(dir, "returned")
	fds, err := os.ReadDir(retdir)
	require.NoError(t, err)
	require.Len(t, fds, 2)
	found, err := ach.ReadFile(filepath.Join(retdir, fds[0].Name()))
	require.NoError(t, err)
	require.Len(t, found.Batches, 1)
	require.Len(t, found.Batches[0].GetEntries(), 1)
	require.Equal(t, "C01", found.Batches[0].GetEntries()[0].Addenda98.ChangeCode)
	found, err = ach.ReadFile(filepath.Join(retdir, fds[1].Name()))
	require.NoError(t, err)
	require.Len(t, found.Batches, 1)
	require.Len(t, found.Batches[0].GetEntries(), 1)
	require.Equal(t, "R03", found.Batches[0].GetEntries()[0].Addenda99.ReturnCode)

	// verify the timestamp on the file is in the future
	fInfo, err := fds[0].Info()
	require.NoError(t, err)
	require.Greater(t, fInfo.ModTime(), time.Now())
	fInfo, err = fds[1].Info()
	require.NoError(t, err)
	require.Greater(t, fInfo.ModTime(), time.Now())

	// verify no "reconciliation" files created
	recondir := filepath.Join(dir, "reconciliation")
	_, err = os.ReadDir(recondir)
	require.Error(t, err)
}

func testFileTransformer(t *testing.T, resp ...service.Response) (*FileTransfomer, string) {
	t.Helper()

	dir, ftpServer := fileBackedFtpServer(t)

	cfg := &service.Config{
		Matching: service.Matching{
			Debug: true,
		},
		Servers: service.ServerConfig{
			FTP: &service.FTPConfig{
				RootPath: dir,
				Paths: service.Paths{
					Return: "./returned/",
				},
			},
		},
	}
	responses := resp

	logger := log.NewTestLogger()
	w := NewFileWriter(logger, cfg.Servers, ftpServer)

	return NewFileTransformer(logger, cfg, responses, w), dir
}
