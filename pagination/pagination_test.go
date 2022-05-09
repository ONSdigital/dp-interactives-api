package pagination

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-net/v2/responder"
	"github.com/stretchr/testify/assert"
)

var (
	respond = responder.New()
)

func TestGetPaginationParametersReturnsErrorWhenOffsetIsNegative(t *testing.T) {

	r := httptest.NewRequest("GET", "/test?offset=-1", nil)
	paginator := &Paginator{}

	offset, limit, err := paginator.getPaginationParameters(r)

	assert.Equal(t, errors.New("invalid query parameter"), err)
	assert.Equal(t, 0, offset)
	assert.Equal(t, 0, limit)
}

func TestGetPaginationParametersReturnsErrorWhenLimitIsNegative(t *testing.T) {

	r := httptest.NewRequest("GET", "/test?limit=-1", nil)
	paginator := &Paginator{}

	offset, limit, err := paginator.getPaginationParameters(r)

	assert.Equal(t, errors.New("invalid query parameter"), err)
	assert.Equal(t, 0, offset)
	assert.Equal(t, 0, limit)
}

func TestGetPaginationParametersReturnsMAxLimitWhenLimitIsGreaterThanMaxLimit(t *testing.T) {

	r := httptest.NewRequest("GET", "/test?limit=1001", nil)
	paginator := &Paginator{DefaultMaxLimit: 1000}

	offset, limit, err := paginator.getPaginationParameters(r)

	assert.Equal(t, nil, err)
	assert.Equal(t, 0, offset)
	assert.Equal(t, paginator.DefaultMaxLimit, limit)
}

func TestGetPaginationParametersReturnsLimitAndOffsetProvidedFromQuery(t *testing.T) {

	r := httptest.NewRequest("GET", "/test?limit=10&offset=5", nil)
	paginator := &Paginator{DefaultMaxLimit: 1000}

	offset, limit, err := paginator.getPaginationParameters(r)

	assert.Equal(t, nil, err)
	assert.Equal(t, 5, offset)
	assert.Equal(t, 10, limit)
}

func TestGetPaginationParametersReturnsDefaultValuesWhenNotProvided(t *testing.T) {

	r := httptest.NewRequest("GET", "/test", nil)
	paginator := &Paginator{DefaultLimit: 20, DefaultOffset: 1, DefaultMaxLimit: 1000}

	offset, limit, err := paginator.getPaginationParameters(r)

	assert.Equal(t, nil, err)
	assert.Equal(t, 1, offset)
	assert.Equal(t, 20, limit)
}

func TestRenderPageReturnsPageStrucWithFilledValues(t *testing.T) {

	expectedPage := page{
		Items:      []int{1, 2, 3},
		Count:      3,
		Offset:     0,
		Limit:      10,
		TotalCount: 3,
	}
	list := []int{1, 2, 3}
	actualPage := renderPage(list, 0, 10, 3)

	assert.Equal(t, expectedPage, actualPage)
}

func TestRenderPageTakesListOfAnyType(t *testing.T) {

	type custom struct {
		name string
	}

	expectedPage := page{
		Items:      []custom{},
		Count:      0,
		Offset:     0,
		Limit:      20,
		TotalCount: 0,
	}
	list := []custom{}
	actualPage := renderPage(list, 0, 20, 0)

	assert.Equal(t, expectedPage, actualPage)
}

func TestNewPaginatorReturnsPaginatorStructWithFilledValues(t *testing.T) {

	expectedPaginator := &Paginator{
		respond:         respond,
		DefaultLimit:    10,
		DefaultOffset:   5,
		DefaultMaxLimit: 100,
	}
	actualPaginator := NewPaginator(respond, 10, 5, 100)

	assert.Equal(t, expectedPaginator, actualPaginator)
}

func TestReturnPaginatedResultsWritesJSONPageToHTTPResponseBody(t *testing.T) {

	r := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	inputPage := page{
		Items:      []int{1, 2, 3},
		Count:      3,
		Offset:     0,
		Limit:      20,
		TotalCount: 3,
	}
	expectedPage := page{
		Items:      []int{1, 2, 3},
		Count:      3,
		Offset:     0,
		Limit:      20,
		TotalCount: 3,
	}
	returnPaginatedResults(respond, w, r, inputPage)

	content, _ := ioutil.ReadAll(w.Body)
	expectedContent, _ := json.Marshal(expectedPage)
	assert.Equal(t, expectedContent, content)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Equal(t, 200, w.Code)
}

func TestReturnPaginatedResultsReturnsErrorIfCanNotMarshalJSON(t *testing.T) {

	r := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	inputPage := page{
		Items:      make(chan int),
		Count:      3,
		Offset:     0,
		Limit:      20,
		TotalCount: 3,
	}

	returnPaginatedResults(respond, w, r, inputPage)
	content, _ := ioutil.ReadAll(w.Body)

	assert.Contains(t, string(content), "errors")
	assert.Equal(t, 500, w.Code)
}

func TestPaginateFunctionPassesParametersDownToProvidedFunction(t *testing.T) {
	r := httptest.NewRequest("GET", "/test?limit=1&offset=2", nil)
	w := httptest.NewRecorder()

	fetchListFunc := func(r *http.Request, limit int, offset int) (interface{}, int, error) {
		return []int{limit, offset}, 10, nil
	}

	paginator := &Paginator{
		DefaultLimit:    10,
		DefaultOffset:   0,
		DefaultMaxLimit: 100,
	}

	paginatedHandler := paginator.Paginate(fetchListFunc)

	expectedPage := page{
		Items:      []int{1, 2},
		Count:      2,
		Offset:     2,
		Limit:      1,
		TotalCount: 10,
	}

	paginatedHandler(w, r)

	content, _ := ioutil.ReadAll(w.Body)
	expectedContent, _ := json.Marshal(expectedPage)

	assert.Equal(t, string(expectedContent), string(content))
	assert.Equal(t, 200, w.Code)
}

func TestPaginateFunctionReturnsBadRequestWhenInvalidQueryParametersAreGiven(t *testing.T) {
	r := httptest.NewRequest("GET", "/test?limit=-1", nil)
	w := httptest.NewRecorder()
	fetchListFunc := func(r *http.Request, limit int, offset int) (interface{}, int, error) {
		return []int{}, 0, nil
	}

	paginator := &Paginator{}
	paginatedHandler := paginator.Paginate(fetchListFunc)

	paginatedHandler(w, r)
	content, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, 400, w.Code)
	assert.Equal(t, "{\"errors\":[\"invalid query parameter\"]}", string(content))
}

func TestPaginateFunctionReturnsListFuncImplementedHttpErrorIfListFuncReturnsAnError(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	fetchListFunc := func(r *http.Request, limit int, offset int) (interface{}, int, error) {
		return 0, 0, errors.New("internal error")
	}

	paginator := &Paginator{}
	paginatedHandler := paginator.Paginate(fetchListFunc)

	paginatedHandler(w, r)
	content, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, 500, w.Code)
	assert.Equal(t, "{\"errors\":[\"internal error\"]}", string(content))
}
