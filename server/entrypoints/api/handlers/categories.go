package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
)

// CategoryService defines the interface for category operations
type CategoryService interface {
	Create(ctx context.Context, category domain.Category) (domain.Category, error)
	GetByID(ctx context.Context, id string) (domain.Category, error)
	List(ctx context.Context, orgID *string) ([]domain.Category, error)
	Update(ctx context.Context, category domain.Category) error
	Delete(ctx context.Context, id string) error
}

// CategoriesHandler handles HTTP requests for categories
type CategoriesHandler struct {
	service CategoryService
}

// NewCategoriesHandler creates a new CategoriesHandler
func NewCategoriesHandler(service CategoryService) *CategoriesHandler {
	return &CategoriesHandler{service: service}
}

// RegisterRoutes registers category routes
func (h *CategoriesHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
}

// CreateCategoryRequest represents the request body for creating a category
type CreateCategoryRequest struct {
	Name         string  `json:"name"`
	OrgID        *string `json:"org_id,omitempty"`
	DisplayOrder int     `json:"display_order"`
}

// CategoryResponse represents the response for a category
type CategoryResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	IsSystem     bool    `json:"is_system"`
	OrgID        *string `json:"org_id,omitempty"`
	DisplayOrder int     `json:"display_order"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

func categoryToResponse(c domain.Category) CategoryResponse {
	return CategoryResponse{
		ID:           c.ID,
		Name:         c.Name,
		IsSystem:     c.IsSystem,
		OrgID:        c.OrgID,
		DisplayOrder: c.DisplayOrder,
		CreatedAt:    c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    c.UpdatedAt.Format(time.RFC3339),
	}
}

// List handles GET /categories
func (h *CategoriesHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("org_id")
	var orgIDPtr *string
	if orgID != "" {
		orgIDPtr = &orgID
	}

	categories, err := h.service.List(r.Context(), orgIDPtr)
	if err != nil {
		log.Printf("Failed to list categories: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var response []CategoryResponse
	for _, c := range categories {
		response = append(response, categoryToResponse(c))
	}

	// Return empty array instead of null
	if response == nil {
		response = []CategoryResponse{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode categories response: %v", err)
	}
}

// Get handles GET /categories/{id}
func (h *CategoriesHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	category, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "category not found", http.StatusNotFound)
			return
		}
		log.Printf("Failed to get category %s: %v", id, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(categoryToResponse(category)); err != nil {
		log.Printf("Failed to encode category response: %v", err)
	}
}

// Create handles POST /categories
func (h *CategoriesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	category := domain.Category{
		Name:         req.Name,
		OrgID:        req.OrgID,
		DisplayOrder: req.DisplayOrder,
		IsSystem:     false, // User-created categories are never system categories
	}

	if err := category.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	created, err := h.service.Create(r.Context(), category)
	if err != nil {
		log.Printf("Failed to create category: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(categoryToResponse(created)); err != nil {
		log.Printf("Failed to encode created category response: %v", err)
	}
}

// Update handles PUT /categories/{id}
func (h *CategoriesHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	existing, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "category not found", http.StatusNotFound)
			return
		}
		log.Printf("Failed to get category %s for update: %v", id, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if existing.IsSystem {
		http.Error(w, "cannot modify system categories", http.StatusForbidden)
		return
	}

	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	existing.Name = req.Name
	existing.DisplayOrder = req.DisplayOrder

	if err := existing.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.Update(r.Context(), existing); err != nil {
		log.Printf("Failed to update category %s: %v", id, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Delete handles DELETE /categories/{id}
func (h *CategoriesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	existing, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "category not found", http.StatusNotFound)
			return
		}
		log.Printf("Failed to get category %s for delete: %v", id, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if existing.IsSystem {
		http.Error(w, "cannot delete system categories", http.StatusForbidden)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		log.Printf("Failed to delete category %s: %v", id, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
