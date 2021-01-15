package viewmodels

import (
	"fmt"
	"regexp"

	"github.com/sirupsen/logrus"
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
		logrus.Info(fmt.Sprintf("setting filter: %s", r.String()))
	} else {
		logrus.Info("setting filter: nil")
	}
	fm.filterRegex = r
}

func (fm *FilterViewModel) GetFilter() *regexp.Regexp {
	return fm.filterRegex
}
