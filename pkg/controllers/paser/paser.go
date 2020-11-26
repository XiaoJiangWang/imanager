package paser

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/golang/glog"

	"imanager/pkg/api/dataselect"
)

const (
	sort         = "sortBy"
	filter       = "filterBy"
	page         = "page"
	itemsPerPage = "itemsPerPage"
)

func PaserDataSelectPathParameter(r *http.Request) *dataselect.DataSelectQuery {
	query := r.URL.Query()
	if len(query) == 0 {
		return nil
	}

	var paginationQuery *dataselect.PaginationQuery = nil
	var sortQuery *dataselect.SortQuery = nil
	if query[page] != nil && len(query[page]) >= 1 && query[itemsPerPage] != nil && len(query[itemsPerPage]) >= 1 {
		paginationQuery = paserPaginationQuery(query[page][0], query[itemsPerPage][0])
	}
	if query[sort] != nil && len(query[sort]) >= 1 {
		sortQuery = paserSortQuery(query[sort][0])
	}
	res := &dataselect.DataSelectQuery{
		PaginationQuery: paginationQuery,
		SortQuery:       sortQuery,
		FilterQuery:     paserFilterQuery(query[filter]),
	}

	return res
}

func paserSortQuery(sortStr string) *dataselect.SortQuery {
	sortRaw := strings.Split(sortStr, ",")
	if len(sortRaw) != 2 {
		glog.Warningf("sortStr[%v] is not format", sortStr)
		return nil
	}
	var ascending bool
	switch sortRaw[0] {
	case "a":
		ascending = true
	case "d":
		ascending = false
	default:
		glog.Warningf("sortStr[%v] can not get ascending", sortStr)
		return nil
	}
	return &dataselect.SortQuery{
		Property:  sortRaw[1],
		Ascending: ascending,
	}
}

func paserFilterQuery(filterStrs []string) *dataselect.FilterQuery {
	if filterStrs == nil || len(filterStrs) == 0 {
		return nil
	}
	filterList := make([]dataselect.FilterBy, 0, len(filterStrs))
	for _, str := range filterStrs {
		raw := strings.Split(str, ",")
		if len(raw) != 2 {
			glog.Warningf("filterStr[%v] is not format, ignore it", str)
			continue
		}
		filterList = append(filterList, dataselect.FilterBy{
			Property: raw[0],
			Value:    raw[1],
		})
	}
	return &dataselect.FilterQuery{FilterByList: filterList}
}

func paserPaginationQuery(page, itemsPerPage string) *dataselect.PaginationQuery {
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		glog.Warningf("transfer page failed, page: %v, err: %v", page, err)
		return nil
	}
	itemsPerPageInt, err := strconv.Atoi(itemsPerPage)
	if err != nil {
		glog.Warningf("transfer itemsPerPage failed, itemsPerPage: %v, err: %v", itemsPerPage, err)
		return nil
	}
	return &dataselect.PaginationQuery{
		ItemsPerPage: itemsPerPageInt,
		Page:         pageInt,
	}
}
