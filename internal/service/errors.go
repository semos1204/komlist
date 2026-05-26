package service

import "errors"

// Sentinel errors returned by TaskService.
var (
	// ErrEmptyTitle is returned when a task is created or renamed with a
	// blank title.
	ErrEmptyTitle = errors.New("title must not be empty")

	// ErrInvalidStatus is returned when an unknown status is requested.
	ErrInvalidStatus = errors.New("invalid status")

	// ErrInvalidPriority is returned when an unknown priority is requested.
	ErrInvalidPriority = errors.New("invalid priority")

	// ErrInvalidSort is returned when an unknown sort key is requested.
	ErrInvalidSort = errors.New("invalid sort")

	// ErrEmptyTag is returned when a tag rename target is blank.
	ErrEmptyTag = errors.New("tag must not be empty")
)
