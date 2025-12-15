package handlers

import (
	"net/http"
	"strconv"

	"github.com/ybotet/notes-api-optimization/internal/db"
	"github.com/ybotet/notes-api-optimization/internal/models"

	"github.com/gin-gonic/gin"
)

type NoteHandler struct {
	repo db.Repository
}

func NewNoteHandler(repo db.Repository) *NoteHandler {
	return &NoteHandler{repo: repo}
}

// CreateNote crea una nueva nota
func (h *NoteHandler) CreateNote(c *gin.Context) {
	var req models.CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note, err := h.repo.CreateNote(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando nota"})
		return
	}

	c.JSON(http.StatusCreated, note)
}

// GetNote obtiene una nota por ID
func (h *NoteHandler) GetNote(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	note, err := h.repo.GetNoteByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo nota"})
		return
	}

	if note == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Nota no encontrada"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// GetNotesBatch obtiene múltiples notas en batch
func (h *NoteHandler) GetNotesBatch(c *gin.Context) {
	idsParam := c.QueryArray("ids")
	if len(idsParam) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Se requieren IDs"})
		return
	}

	var ids []int64
	for _, idStr := range idsParam {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
			return
		}
		ids = append(ids, id)
	}

	notes, err := h.repo.GetNotesBatch(c.Request.Context(), ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo notas"})
		return
	}

	c.JSON(http.StatusOK, notes)
}

// ListNotes lista notas con paginación
func (h *NoteHandler) ListNotes(c *gin.Context) {
	var params models.PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	page, err := h.repo.ListNotes(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error listando notas"})
		return
	}

	c.JSON(http.StatusOK, page)
}

// SearchNotes busca notas por título
func (h *NoteHandler) SearchNotes(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Se requiere query de búsqueda"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	notes, err := h.repo.SearchNotes(c.Request.Context(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error buscando notas"})
		return
	}

	c.JSON(http.StatusOK, notes)
}

// UpdateNote actualiza una nota
func (h *NoteHandler) UpdateNote(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var req models.UpdateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note, err := h.repo.UpdateNote(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando nota"})
		return
	}

	if note == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Nota no encontrada"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// DeleteNote elimina una nota
func (h *NoteHandler) DeleteNote(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.repo.DeleteNote(c.Request.Context(), id); err != nil {
		if err.Error() == "nota no encontrada" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Nota no encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando nota"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetStats obtiene estadísticas del sistema
func (h *NoteHandler) GetStats(c *gin.Context) {
	stats, err := h.repo.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo estadísticas"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
