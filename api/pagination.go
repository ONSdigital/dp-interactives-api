package api

import (
	"fmt"
	"github.com/ONSdigital/dp-net/v2/responder"
	"net/http"
	"reflect"
	"strconv"
)

// PaginatedHandler is a func type for an endpoint that returns a list of values that we want to paginate
type PaginatedHandler func(r *http.Request, limit int, offset int) (list interface{}, totalCount int, err error)

type page struct {
	Items      interface{} `json:"items"`
	Count      int         `json:"count"`
	Offset     int         `json:"offset"`
	Limit      int         `json:"limit"`
	TotalCount int         `json:"total_count"`
}

type Paginator struct {
	DefaultLimit    int
	DefaultOffset   int
	DefaultMaxLimit int
}

func NewPaginator(defaultLimit, defaultOffset, defaultMaxLimit int) *Paginator {
	return &Paginator{
		DefaultLimit:    defaultLimit,
		DefaultOffset:   defaultOffset,
		DefaultMaxLimit: defaultMaxLimit,
	}
}

// Paginate wraps a http endpoint to return a paginated list from the list returned by the provided function
func (p *Paginator) Paginate(respond *responder.Responder, paginatedHandler PaginatedHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		offset, limit, err := p.getPaginationParameters(r)
		if err != nil {
			respond.Error(ctx, w, http.StatusBadRequest, err)
			return
		}

		list, totalCount, err := paginatedHandler(r, limit, offset)
		if err != nil {
			respond.Error(ctx, w, http.StatusInternalServerError, err)
			return
		}

		respond.JSON(ctx, w, http.StatusOK, page{
			Items:      list,
			Count:      listLength(list),
			Offset:     offset,
			Limit:      limit,
			TotalCount: totalCount,
		})
	}
}

func (p *Paginator) getPaginationParameters(r *http.Request) (offset int, limit int, err error) {
	offsetParameter := r.URL.Query().Get("offset")
	limitParameter := r.URL.Query().Get("limit")

	offset = p.DefaultOffset
	limit = p.DefaultLimit

	if offsetParameter != "" {
		offset, err = strconv.Atoi(offsetParameter)
		if err != nil || offset < 0 {
			return 0, 0, fmt.Errorf("invalid query parameter offset=%s", offsetParameter)
		}
	}

	if limitParameter != "" {
		limit, err = strconv.Atoi(limitParameter)
		if err != nil || limit < 0 {
			return 0, 0, fmt.Errorf("invalid query parameter limit=%s", limitParameter)
		}
	}

	if limit > p.DefaultMaxLimit {
		limit = p.DefaultMaxLimit
	}
	return
}

func listLength(list interface{}) int {
	l := reflect.ValueOf(list)
	return l.Len()
}
