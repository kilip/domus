package model

type PaginatedResult[T any] struct {
	Data       []T
	NextCursor string
}

