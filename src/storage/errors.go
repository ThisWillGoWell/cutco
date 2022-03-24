package storage

import "errors"

var (
	KeyAlreadyExists = errors.New("pk/sk already exists in ddb")
	NoEntriesFound   = errors.New("no entries found")
)
