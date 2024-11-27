package testlib

type ContainersNotStarted struct {
	Name string
}

func (e *ContainersNotStarted) Error() string {
	return "No containers with logs"
}
