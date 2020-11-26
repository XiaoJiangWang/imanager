package dataselect

type DataSelectQuery struct {
	PaginationQuery *PaginationQuery
	SortQuery       *SortQuery
	FilterQuery     *FilterQuery
}

// PaginationQuery structure represents pagination settings
type PaginationQuery struct {
	// How many items per page should be returned
	ItemsPerPage int
	// Number of page that should be returned when pagination is applied to the list
	Page int
}

// SortQuery holds the name of the property that should be sorted and whether order should be ascending or descending.
type SortQuery struct {
	Property  string
	Ascending bool
}

type FilterQuery struct {
	FilterByList []FilterBy
}

type FilterBy struct {
	Property string
	Value    string
}