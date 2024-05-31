package instantverify

import (
	"strings"
)

type InstantVerifyService interface {
	Search(ops SearchOptions) ([]string, error)
}

type instantVerifyService struct {
	repository InstantVerifyRepository
}

func NewInstantVerifyService(repository InstantVerifyRepository) InstantVerifyService {
	return &instantVerifyService{
		repository: repository,
	}
}

type SearchOptions struct {
	TraceNumber string
	Path        string
}

func (s *instantVerifyService) Search(opts SearchOptions) (codes []string, err error) {
	headers, err := s.repository.Search(opts)
	if err != nil {
		return nil, err
	}

	for _, h := range headers {
		c := strings.TrimSpace(h.CompanyNameField())
		if strings.HasPrefix(c, "MV") {
			codes = append(codes, c)
		}
	}

	return codes, nil
}
