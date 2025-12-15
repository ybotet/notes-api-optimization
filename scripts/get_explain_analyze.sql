-- Подключиться к PostgreSQL
docker-compose exec postgres psql -U user -d notes

-- 1. Пагинация с OFFSET (неэффективно)
EXPLAIN (ANALYZE, BUFFERS, VERBOSE)
SELECT id, title, content, created_at
FROM notes
ORDER BY created_at DESC, id DESC
OFFSET 1000 LIMIT 20;

-- 2. Пагинация с Keyset (эффективно)
EXPLAIN (ANALYZE, BUFFERS, VERBOSE)
SELECT id, title, content, created_at
FROM notes
WHERE (created_at, id) < (now() - interval '1 day', 1000)
ORDER BY created_at DESC, id DESC
LIMIT 20;

-- 3. Поиск без индекса (если удалить индекс)
DROP INDEX IF EXISTS idx_notes_title_gin;
EXPLAIN (ANALYZE, BUFFERS, VERBOSE)
SELECT id, title, content
FROM notes
WHERE title LIKE '%Заметка%'
LIMIT 10;

-- 4. Поиск с GIN индексом (восстановить индекс)
CREATE INDEX IF NOT EXISTS idx_notes_title_gin 
ON notes USING GIN (to_tsvector('simple', title));
EXPLAIN (ANALYZE, BUFFERS, VERBOSE)
SELECT id, title, content
FROM notes
WHERE to_tsvector('simple', title) @@ plainto_tsquery('simple', 'Заметка')
LIMIT 10;

-- 5. Batch запрос
EXPLAIN (ANALYZE, BUFFERS, VERBOSE)
SELECT id, title, content
FROM notes
WHERE id = ANY(ARRAY[1,2,3,4,5,6,7,8,9,10]);