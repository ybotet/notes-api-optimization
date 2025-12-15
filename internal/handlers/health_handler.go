package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	pool *pgxpool.Pool
}

func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{pool: pool}
}

// HealthCheck verifica el estado de la API y la base de datos
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	status := gin.H{
		"status":  "healthy",
		"time":    time.Now().UTC(),
		"service": "notes-api",
	}

	// Verificar conexión a la base de datos
	ctx := c.Request.Context()
	if err := h.pool.Ping(ctx); err != nil {
		status["status"] = "unhealthy"
		status["database"] = "disconnected"
		c.JSON(http.StatusServiceUnavailable, status)
		return
	}

	// Estadísticas del pool
	stats := h.pool.Stat()
	status["database"] = "connected"
	status["pool_stats"] = gin.H{
		"total_connections":    stats.TotalConns(),
		"idle_connections":     stats.IdleConns(),
		"max_connections":      stats.MaxConns(),
		"acquired_connections": stats.AcquiredConns(),
	}

	c.JSON(http.StatusOK, status)
}
