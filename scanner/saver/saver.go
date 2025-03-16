package saver

import "fmt"

type Saver struct {
}

func NewSaver() *Saver {
	return &Saver{}
}

func (s *Saver) Save(warnings []string) {
	for _, w := range warnings {
		fmt.Println("[WARNING]", w)
	}
}
