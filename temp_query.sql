-- 1. Пагинация с OFFSET
EXPLAIN (ANALYZE, BUFFERS, VERBOSE)
SELECT id, title, content, created_at
FROM notes
ORDER BY created_at DESC, id DESC
OFFSET 100 LIMIT 20;

-- 2. Пагинация с Keyset
EXPLAIN (ANALYZE, BUFFERS, VERBOSE)
SELECT id, title, content, created_at
FROM notes
WHERE (created_at, id) < (now() - interval '1 day', 100)
ORDER BY created_at DESC, id DESC
LIMIT 20;

-- 3. Batch запрос
EXPLAIN (ANALYZE, BUFFERS, VERBOSE)
SELECT id, title, content
FROM notes
WHERE id = ANY(ARRAY[1,2,3,4,5]);
