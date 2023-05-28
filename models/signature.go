package models

import "time"

type signature struct {
	text       string
	created_at time.Time
}

func NewSignature(text string, created_at time.Time) *signature {
	s := signature{}
	s.text = text
	s.created_at = created_at
	return &s
}
