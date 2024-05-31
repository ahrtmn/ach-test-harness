package instantverify

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach-test-harness/pkg/response/match"
	"github.com/moov-io/ach-test-harness/pkg/service"
	"github.com/moov-io/base/log"
)

type InstantVerifyRepository interface {
	Search(opts SearchOptions) ([]*ach.BatchHeader, error)
}

type ftpBatchHeaderRepository struct {
	logger   log.Logger
	rootPath string
}

func NewInstantVerifyRepository(logger log.Logger, cfg *service.FTPConfig) *ftpBatchHeaderRepository {
	return &ftpBatchHeaderRepository{
		logger:   logger,
		rootPath: cfg.RootPath,
	}
}

func (r *ftpBatchHeaderRepository) Search(opts SearchOptions) ([]*ach.BatchHeader, error) {
	out := make([]*ach.BatchHeader, 0)

	var search fs.WalkDirFunc
	search = func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		r.logger.Logf("reading %s", path)
		if strings.ToLower(filepath.Ext(path)) != ".ach" {
			return nil
		}

		entries, err := filterEntries(path, opts)
		if err != nil {
			return err
		}
		out = append(out, entries...)
		return nil
	}

	var walkingPath = r.rootPath
	if opts.Path != "" {
		walkingPath = filepath.Join(r.rootPath, opts.Path)
	}

	r.logger.Logf("Walking directory %s", walkingPath)
	if err := filepath.WalkDir(walkingPath, search); err != nil {
		return nil, fmt.Errorf("failed reading directory content %s: %v", walkingPath, err)
	}

	return out, nil
}

func filterEntries(path string, opts SearchOptions) ([]*ach.BatchHeader, error) {
	file, _ := ach.ReadFile(path)
	if file == nil {
		return nil, nil
	}

	mm := service.Match{
		TraceNumber: opts.TraceNumber,
	}

	var out []*ach.BatchHeader
	for i := range file.Batches {
		entries := file.Batches[i].GetEntries()
		header := file.Batches[i].GetHeader()
		if mm.Empty() {
			out = append(out, header)
			continue
		}

		for j := range entries {
			if match.TraceNumber(mm, entries[j]) {
				out = append(out, header)
				break
			}
		}
	}
	return out, nil
}
