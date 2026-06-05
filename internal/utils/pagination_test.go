package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParsePaginationUsesDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := contextWithQuery("")

	pagination := ParsePagination(ctx)
	if pagination.Page != 1 || pagination.PageSize != 20 || pagination.Offset != 0 {
		t.Fatalf("unexpected pagination: %#v", pagination)
	}
}

func TestParsePaginationClampsPageSize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := contextWithQuery("?page=3&page_size=200")

	pagination := ParsePagination(ctx)
	if pagination.Page != 3 || pagination.PageSize != 100 || pagination.Offset != 200 {
		t.Fatalf("unexpected pagination: %#v", pagination)
	}
}

func TestParsePaginationIgnoresInvalidValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := contextWithQuery("?page=-1&page_size=bad")

	pagination := ParsePagination(ctx)
	if pagination.Page != 1 || pagination.PageSize != 20 || pagination.Offset != 0 {
		t.Fatalf("unexpected pagination: %#v", pagination)
	}
}

func contextWithQuery(query string) *gin.Context {
	request := httptest.NewRequest(http.MethodGet, "/"+query, nil)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = request
	return ctx
}
