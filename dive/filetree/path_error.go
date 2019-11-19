package filetree

import "fmt"

const (
	ActionAdd FileAction = iota
	ActionRemove
)

type FileAction int

func (fa FileAction) String() string {
	switch fa {
	case ActionAdd:
		return "add"
	case ActionRemove:
		return "remove"
	default:
		return "<unknown file action>"
	}
}

type PathError struct {
	Path   string
	Action FileAction
	Err    error
}

func NewPathError(path string, action FileAction, err error) PathError {
	return PathError{
		Path:   path,
		Action: action,
		Err:    err,
	}
}

func (pe PathError) String() string {
	return fmt.Sprintf("unable to %s '%s': %+v", pe.Action.String(), pe.Path, pe.Err)
}
