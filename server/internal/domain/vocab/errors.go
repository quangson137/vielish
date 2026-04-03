package domain

import "errors"

var (
	ErrTopicNotFound = errors.New("topic not found")
	ErrWordNotFound  = errors.New("word not found")
)
