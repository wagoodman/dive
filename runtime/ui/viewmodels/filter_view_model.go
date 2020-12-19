package viewmodels

import (
	"regexp"

	"go.uber.org/zap"
)

type FilterViewModel struct {
	filterRegex *regexp.Regexp
}

func NewFilterViewModel(r *regexp.Regexp) *FilterViewModel {
	return &FilterViewModel{
		filterRegex: r,
	}
}

func (fm *FilterViewModel) SetFilter(r *regexp.Regexp) {
	if r != nil {
		zap.S().Info("setting filter ", r.String())
	} else {
		zap.S().Info("setting filter nil")
	}
	fm.filterRegex = r
}

func (fm *FilterViewModel) GetFilter() *regexp.Regexp {
	return fm.filterRegex
}
