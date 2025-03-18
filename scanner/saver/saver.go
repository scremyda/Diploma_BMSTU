package saver

import "fmt"

type Saver struct {
}

func NewSaver() *Saver {
	return &Saver{}
}

func (s *Saver) Save(err error) {
	fmt.Println("[WARNING]", err)
}
