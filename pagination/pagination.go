package pagination

import (
	"errors"
	"github.com/ONSdigital/dp-net/v2/responder"
	"net/http"
	"reflect"
	"strconv"

	"github.com/ONSdigital/log.go/v2/log"
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
	respond         *responder.Responder
	DefaultLimit    int
	DefaultOffset   int
	DefaultMaxLimit int
}

func NewPaginator(respond *responder.Responder, defaultLimit, defaultOffset, defaultMaxLimit int) *Paginator {

	return &Paginator{
		respond:         respond,
		DefaultLimit:    defaultLimit,
		DefaultOffset:   defaultOffset,
		DefaultMaxLimit: defaultMaxLimit,
	}
}

// Paginate wraps a http endpoint to return a paginated list from the list returned by the provided function
func (p *Paginator) Paginate(paginatedHandler PaginatedHandler) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		offset, limit, err := p.getPaginationParameters(r)
		if err != nil {
			p.respond.Error(r.Context(), w, http.StatusBadRequest, err)
			return
		}
		list, totalCount, err := paginatedHandler(r, limit, offset)
		if err != nil {
			p.respond.Error(r.Context(), w, http.StatusInternalServerError, err)
			return
		}

		returnPaginatedResults(p.respond, w, r, renderPage(list, offset, limit, totalCount))
	}
}

func (p *Paginator) getPaginationParameters(r *http.Request) (offset int, limit int, err error) {

	logData := log.Data{}
	offsetParameter := r.URL.Query().Get("offset")
	limitParameter := r.URL.Query().Get("limit")

	offset = p.DefaultOffset
	limit = p.DefaultLimit

	if offsetParameter != "" {
		logData["offset"] = offsetParameter
		offset, err = strconv.Atoi(offsetParameter)
		if err != nil || offset < 0 {
			return 0, 0, errors.New("invalid query parameter")
		}
	}

	if limitParameter != "" {
		logData["limit"] = limitParameter
		limit, err = strconv.Atoi(limitParameter)
		if err != nil || limit < 0 {
			return 0, 0, errors.New("invalid query parameter")
		}
	}

	if limit > p.DefaultMaxLimit {
		logData["max_limit"] = p.DefaultMaxLimit
		log.Warn(r.Context(), "defaulting to max limit", logData)
		limit = p.DefaultMaxLimit
	}
	return
}

func renderPage(list interface{}, offset int, limit int, totalCount int) page {

	return page{
		Items:      list,
		Count:      listLength(list),
		Offset:     offset,
		Limit:      limit,
		TotalCount: totalCount,
	}
}

func listLength(list interface{}) int {
	l := reflect.ValueOf(list)
	return l.Len()
}

func returnPaginatedResults(respond *responder.Responder, w http.ResponseWriter, r *http.Request, list page) {

	respond.JSON(r.Context(), w, http.StatusOK, list)
}
