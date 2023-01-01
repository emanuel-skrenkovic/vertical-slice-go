package core

type CommandResponse struct {
	payload    interface{}
	StatusCode int
	Reason     *string
}
