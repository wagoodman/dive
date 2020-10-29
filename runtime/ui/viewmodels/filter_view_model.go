package viewmodels

import "regexp"

type FilterViewModel struct {
	filterRegex *regexp.Regexp
}

func NewFilterViewModel(r *regexp.Regexp) *FilterViewModel {
	return &FilterViewModel{
		filterRegex: r,
	}
}

func (fm *FilterViewModel) SetFilter(r *regexp.Regexp) {
	fm.filterRegex = r
}

func (fm *FilterViewModel) GetFilter() *regexp.Regexp {
	return fm.filterRegex
}
