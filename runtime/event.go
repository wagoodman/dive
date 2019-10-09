package runtime

type eventChannel chan event

type event struct {
	stdout      string
	stderr      string
	err         error
	errorOnExit bool
}

func (ec eventChannel) message(msg string) {
	ec <- event{
		stdout: msg,
	}
}

func (ec eventChannel) exitWithError(err error) {
	ec <- event{
		err:         err,
		errorOnExit: true,
	}
}

func (ec eventChannel) exitWithErrorMessage(msg string, err error) {
	ec <- event{
		stderr:      msg,
		err:         err,
		errorOnExit: true,
	}
}
