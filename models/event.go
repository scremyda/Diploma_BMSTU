package models

type ErrorEvent struct {
	Target  string `json:"target"`
	Message string `json:"message"`
}
