package util

import (
	"github.com/astaxie/beego/orm"

	"imanager/pkg/api/dataselect"
)

func PaserQuerySeter(origin orm.QuerySeter, userIDs []string, query *dataselect.DataSelectQuery, existKey map[string]bool) (orm.QuerySeter, int64, error) {
	if userIDs != nil && len(userIDs) != 0 {
		origin = origin.Filter("user_id__in", userIDs)
	}
	if query != nil && query.FilterQuery != nil {
		for _, filter := range query.FilterQuery.FilterByList {
			if !existKey[filter.Property] {
				continue
			}
			origin = origin.Filter(filter.Property+"__contains", filter.Value)
		}
	}
	if query != nil && query.SortQuery != nil && existKey[query.SortQuery.Property] {
		if query.SortQuery.Ascending {
			origin = origin.OrderBy(query.SortQuery.Property)
		} else {
			origin = origin.OrderBy("-" + query.SortQuery.Property)
		}
	}
	nums, err := origin.Count()
	if err != nil {
		return origin, 0, err
	}
	if query != nil && query.PaginationQuery != nil {
		start := (query.PaginationQuery.Page - 1) * query.PaginationQuery.ItemsPerPage
		origin = origin.Limit(query.PaginationQuery.ItemsPerPage, start)
	}
	return origin, nums, err
}