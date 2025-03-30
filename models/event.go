package models

type Domain string

type ErrorEvent struct {
	Target  string `json:"target"`
	Message string `json:"message"`
}
