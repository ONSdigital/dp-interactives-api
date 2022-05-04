package api

import (
	"context"
	"fmt"
	"github.com/ONSdigital/dp-interactives-api/models"
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
func (p *Paginator) Paginate(paginatedHandler PaginatedHandler) func(ctx context.Context, w http.ResponseWriter, r *http.Request) (*models.SuccessResponse, *models.ErrorResponse) {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) (*models.SuccessResponse, *models.ErrorResponse) {
		offset, limit, err := p.getPaginationParameters(r)
		if err != nil {
			responseErr := models.NewError(ctx, err, RequestErrorCode, err.Error())
			return nil, models.NewErrorResponse(http.StatusBadRequest, nil, responseErr)
		}

		list, totalCount, err := paginatedHandler(r, limit, offset)
		if err != nil {
			responseErr := models.NewError(ctx, err, DbErrorCode, err.Error())
			return nil, models.NewErrorResponse(http.StatusInternalServerError, nil, responseErr)
		}

		jsonB, err := JSONify(page{
			Items:      list,
			Count:      listLength(list),
			Offset:     offset,
			Limit:      limit,
			TotalCount: totalCount,
		})
		if err != nil {
			responseErr := models.NewError(ctx, err, MarshallingErrorCode, err.Error())
			return nil, models.NewErrorResponse(http.StatusInternalServerError, nil, responseErr)
		}

		return models.NewSuccessResponse(jsonB, http.StatusAccepted), nil
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
