package models

type PageableObject struct {
	Sort *SortObject `json:"sort,omitempty"`

	PageNumber int32 `json:"pageNumber,omitempty"`

	PageSize int32 `json:"pageSize,omitempty"`

	Offset int32 `json:"offset,omitempty"`

	Paged bool `json:"paged,omitempty"`

	Unpaged bool `json:"unpaged,omitempty"`
}

type SortObject struct {
	Sorted bool `json:"sorted,omitempty"`

	Unsorted bool `json:"unsorted,omitempty"`
}
