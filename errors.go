package ebnf

type JsonError struct {
	wrapped error
}

func (j *JsonError) Error() string {
	return j.wrapped.Error()
}

func (j *JsonError) Unwrap() error {
	return j.wrapped
}
