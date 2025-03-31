package models

type Domain string

type AlerterEvent struct {
	Target  string `json:"target"`
	Message string `json:"message"`
}

type CerterEvent struct {
	Target string `json:"target"`
}
