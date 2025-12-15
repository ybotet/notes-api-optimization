package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ybotet/notes-api-optimization/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateNote(ctx context.Context, note *models.CreateNoteRequest) (*models.Note, error)
	GetNoteByID(ctx context.Context, id int64) (*models.Note, error)
	GetNotesBatch(ctx context.Context, ids []int64) ([]models.Note, error)
	ListNotes(ctx context.Context, params models.PaginationParams) (*models.NotesPage, error)
	SearchNotes(ctx context.Context, query string, limit int) ([]models.Note, error)
	UpdateNote(ctx context.Context, id int64, update models.UpdateNoteRequest) (*models.Note, error)
	DeleteNote(ctx context.Context, id int64) error
	GetStats(ctx context.Context) (map[string]interface{}, error)
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// CreateNote crea una nueva nota (usando prepared statement)
func (r *PostgresRepository) CreateNote(ctx context.Context, req *models.CreateNoteRequest) (*models.Note, error) {
	query := `
        INSERT INTO notes (title, content) 
        VALUES ($1, $2) 
        RETURNING id, title, content, created_at, updated_at
    `

	var note models.Note
	err := r.pool.QueryRow(ctx, query, req.Title, req.Content).
		Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("error creando nota: %w", err)
	}

	return &note, nil
}

// GetNoteByID obtiene una nota por ID (para cache individual)
func (r *PostgresRepository) GetNoteByID(ctx context.Context, id int64) (*models.Note, error) {
	query := `
        SELECT id, title, content, created_at, updated_at 
        FROM notes 
        WHERE id = $1
    `

	var note models.Note
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error obteniendo nota: %w", err)
	}

	return &note, nil
}

// GetNotesBatch obtiene múltiples notas en un solo query (evita N+1)
func (r *PostgresRepository) GetNotesBatch(ctx context.Context, ids []int64) ([]models.Note, error) {
	if len(ids) == 0 {
		return []models.Note{}, nil
	}

	query := `
        SELECT id, title, content, created_at, updated_at 
        FROM notes 
        WHERE id = ANY($1)
        ORDER BY id
    `

	rows, err := r.pool.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo notas batch: %w", err)
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		var note models.Note
		if err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error escaneando nota: %w", err)
		}
		notes = append(notes, note)
	}

	return notes, nil
}

// ListNotes lista notas con paginación por keyset (más eficiente que OFFSET)
func (r *PostgresRepository) ListNotes(ctx context.Context, params models.PaginationParams) (*models.NotesPage, error) {
	if params.Limit == 0 {
		params.Limit = 20
	}

	var notes []models.Note
	var query string
	var args []interface{}

	// Keyset pagination: usar cursor (created_at, id) en lugar de OFFSET
	if params.CursorTime.IsZero() || params.CursorID == 0 {
		// Primera página
		query = `
            SELECT id, title, content, created_at, updated_at 
            FROM notes 
            ORDER BY created_at DESC, id DESC 
            LIMIT $1
        `
		args = []interface{}{params.Limit + 1} // +1 para saber si hay más páginas
	} else {
		// Páginas siguientes
		query = `
            SELECT id, title, content, created_at, updated_at 
            FROM notes 
            WHERE (created_at, id) < ($1, $2)
            ORDER BY created_at DESC, id DESC 
            LIMIT $3
        `
		args = []interface{}{params.CursorTime, params.CursorID, params.Limit + 1}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error listando notas: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var note models.Note
		if err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error escaneando nota: %w", err)
		}
		notes = append(notes, note)
	}

	// Determinar si hay siguiente página
	hasNext := len(notes) > params.Limit
	if hasNext {
		notes = notes[:params.Limit] // Remover el elemento extra
	}

	// Preparar cursor para siguiente página
	var nextCursor string
	if hasNext && len(notes) > 0 {
		lastNote := notes[len(notes)-1]
		nextCursor = fmt.Sprintf("cursor_time=%s&cursor_id=%d",
			lastNote.CreatedAt.Format(time.RFC3339),
			lastNote.ID)
	}

	return &models.NotesPage{
		Notes:    notes,
		NextPage: hasNext,
		Cursor:   nextCursor,
	}, nil
}

// SearchNotes busca notas por título usando índice GIN
func (r *PostgresRepository) SearchNotes(ctx context.Context, query string, limit int) ([]models.Note, error) {
	if limit == 0 {
		limit = 10
	}

	sqlQuery := `
        SELECT id, title, content, created_at, updated_at 
        FROM notes 
        WHERE to_tsvector('simple', title) @@ plainto_tsquery('simple', $1)
        ORDER BY created_at DESC 
        LIMIT $2
    `

	rows, err := r.pool.Query(ctx, sqlQuery, query, limit)
	if err != nil {
		return nil, fmt.Errorf("error buscando notas: %w", err)
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		var note models.Note
		if err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error escaneando nota: %w", err)
		}
		notes = append(notes, note)
	}

	return notes, nil
}

// UpdateNote actualiza una nota en transacción
func (r *PostgresRepository) UpdateNote(ctx context.Context, id int64, update models.UpdateNoteRequest) (*models.Note, error) {
	// Construir query dinámica basada en campos proporcionados
	var setClauses []string
	var args []interface{}
	argIndex := 1

	if update.Title != "" {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, update.Title)
		argIndex++
	}

	if update.Content != "" {
		setClauses = append(setClauses, fmt.Sprintf("content = $%d", argIndex))
		args = append(args, update.Content)
		argIndex++
	}

	if len(setClauses) == 0 {
		return r.GetNoteByID(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = NOW()")

	args = append(args, id)
	query := fmt.Sprintf(`
        UPDATE notes 
        SET %s 
        WHERE id = $%d 
        RETURNING id, title, content, created_at, updated_at
    `, strings.Join(setClauses, ", "), argIndex)

	var note models.Note
	err := r.pool.QueryRow(ctx, query, args...).
		Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error actualizando nota: %w", err)
	}

	return &note, nil
}

// DeleteNote elimina una nota
func (r *PostgresRepository) DeleteNote(ctx context.Context, id int64) error {
	query := `DELETE FROM notes WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error eliminando nota: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("nota no encontrada")
	}

	return nil
}

// GetStats obtiene estadísticas de la base de datos
func (r *PostgresRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Estadísticas del pool
	poolStats := r.pool.Stat()
	stats["pool"] = map[string]interface{}{
		"total_conns":        poolStats.TotalConns(),
		"idle_conns":         poolStats.IdleConns(),
		"max_conns":          poolStats.MaxConns(),
		"acquired_conns":     poolStats.AcquiredConns(),
		"constructing_conns": poolStats.ConstructingConns(),
		"empty_acquire":      poolStats.EmptyAcquireCount(),
		"canceled_acquire":   poolStats.CanceledAcquireCount(),
	}

	// Estadísticas de pg_stat_statements (si está habilitado)
	query := `
        SELECT 
            query,
            calls,
            total_exec_time,
            mean_exec_time,
            rows
        FROM pg_stat_statements 
        WHERE query NOT LIKE '%pg_stat_statements%'
        ORDER BY total_exec_time DESC 
        LIMIT 10
    `

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		// Si falla, pg_stat_statements no está habilitado
		stats["query_stats"] = "pg_stat_statements no habilitado"
	} else {
		defer rows.Close()

		var queryStats []map[string]interface{}
		for rows.Next() {
			var queryText string
			var calls int64
			var totalTime, meanTime float64
			var rowsReturned int64

			if err := rows.Scan(&queryText, &calls, &totalTime, &meanTime, &rowsReturned); err != nil {
				continue
			}

			// Acortar query para mejor visualización
			if len(queryText) > 100 {
				queryText = queryText[:100] + "..."
			}

			queryStats = append(queryStats, map[string]interface{}{
				"query":      queryText,
				"calls":      calls,
				"total_time": totalTime,
				"mean_time":  meanTime,
				"rows":       rowsReturned,
			})
		}
		stats["top_queries"] = queryStats
	}

	return stats, nil
}
