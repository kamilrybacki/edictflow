package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
)

type mockCategoryService struct {
	categories map[string]domain.Category
}

func newMockCategoryService() *mockCategoryService {
	return &mockCategoryService{
		categories: make(map[string]domain.Category),
	}
}

func (m *mockCategoryService) Create(ctx context.Context, cat domain.Category) (domain.Category, error) {
	cat.ID = "cat-new"
	m.categories[cat.ID] = cat
	return cat, nil
}

func (m *mockCategoryService) GetByID(ctx context.Context, id string) (domain.Category, error) {
	cat, ok := m.categories[id]
	if !ok {
		return domain.Category{}, ErrNotFound
	}
	return cat, nil
}

func (m *mockCategoryService) List(ctx context.Context, orgID *string) ([]domain.Category, error) {
	var result []domain.Category
	for _, c := range m.categories {
		result = append(result, c)
	}
	return result, nil
}

func (m *mockCategoryService) Update(ctx context.Context, cat domain.Category) error {
	if _, ok := m.categories[cat.ID]; !ok {
		return ErrNotFound
	}
	m.categories[cat.ID] = cat
	return nil
}

func (m *mockCategoryService) Delete(ctx context.Context, id string) error {
	if _, ok := m.categories[id]; !ok {
		return ErrNotFound
	}
	delete(m.categories, id)
	return nil
}

func TestCategoriesHandler_List(t *testing.T) {
	mock := newMockCategoryService()
	mock.categories["cat-1"] = domain.Category{ID: "cat-1", Name: "Security", IsSystem: true}

	h := NewCategoriesHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var response []CategoryResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response) != 1 {
		t.Errorf("expected 1 category, got %d", len(response))
	}
}

func TestCategoriesHandler_Create(t *testing.T) {
	mock := newMockCategoryService()
	h := NewCategoriesHandler(mock)

	body := `{"name": "Custom Category", "display_order": 5}`
	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCategoriesHandler_Delete(t *testing.T) {
	mock := newMockCategoryService()
	mock.categories["cat-1"] = domain.Category{ID: "cat-1", Name: "Custom", IsSystem: false}

	h := NewCategoriesHandler(mock)

	r := chi.NewRouter()
	r.Delete("/categories/{id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/categories/cat-1", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rec.Code)
	}
}

func TestCategoriesHandler_DeleteSystemCategory(t *testing.T) {
	mock := newMockCategoryService()
	mock.categories["cat-1"] = domain.Category{ID: "cat-1", Name: "Security", IsSystem: true}

	h := NewCategoriesHandler(mock)

	r := chi.NewRouter()
	r.Delete("/categories/{id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/categories/cat-1", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}
