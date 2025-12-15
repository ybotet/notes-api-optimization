-- Tabla de notas
CREATE TABLE IF NOT EXISTS notes (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Índice para búsqueda por título (GIN para búsqueda de texto)
CREATE INDEX IF NOT EXISTS idx_notes_title_gin 
ON notes USING GIN (to_tsvector('simple', title));

-- Índice compuesto para keyset pagination
CREATE INDEX IF NOT EXISTS idx_notes_created_id 
ON notes (created_at DESC, id DESC);

-- Índice para búsquedas frecuentes por created_at
CREATE INDEX IF NOT EXISTS idx_notes_created_at 
ON notes (created_at DESC);

-- Habilitar pg_stat_statements para monitoreo
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Insertar datos de prueba (opcional)
INSERT INTO notes (title, content) 
SELECT 
    'Примечание ' || i,
    'Содержание примечания номер ' || i
FROM generate_series(1, 1000) AS i
ON CONFLICT DO NOTHING;