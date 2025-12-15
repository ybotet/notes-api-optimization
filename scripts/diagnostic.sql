-- 1. Ver índices existentes
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE schemaname = 'public'
ORDER BY tablename, indexname;

-- 2. Ver tamaño de tablas e índices
SELECT 
    tablename,
    pg_size_pretty(pg_total_relation_size(quote_ident(tablename))) as total_size,
    pg_size_pretty(pg_relation_size(quote_ident(tablename))) as table_size,
    pg_size_pretty(pg_total_relation_size(quote_ident(tablename)) - pg_relation_size(quote_ident(tablename))) as index_size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(quote_ident(tablename)) DESC;

-- 3. Ver estadísticas de pg_stat_statements (top 10 queries más pesados)
SELECT 
    query,
    calls,
    total_exec_time,
    mean_exec_time,
    rows
FROM pg_stat_statements 
WHERE query NOT LIKE '%pg_stat_statements%'
ORDER BY total_exec_time DESC 
LIMIT 10;

-- 4. Ver conexiones activas
SELECT 
    pid,
    usename,
    application_name,
    client_addr,
    backend_start,
    state,
    query
FROM pg_stat_activity
WHERE state IS NOT NULL
ORDER BY backend_start DESC;

-- 5. EXPLAIN ANALYZE de consultas comunes

-- Consulta de paginación con OFFSET (ineficiente)
EXPLAIN (ANALYZE, BUFFERS, VERBOSE)
SELECT id, title, content, created_at
FROM notes
ORDER BY created_at DESC, id DESC
OFFSET 5000 LIMIT 20;

-- Consulta de paginación con keyset (eficiente)
EXPLAIN (ANALYZE, BUFFERS, VERBOSE)
SELECT id, title, content, created_at
FROM notes
WHERE (created_at, id) < ('2024-01-01 00:00:00', 5000)
ORDER BY created_at DESC, id DESC
LIMIT 20;

-- Consulta de búsqueda con índice GIN
EXPLAIN (ANALYZE, BUFFERS, VERBOSE)
SELECT id, title, content
FROM notes
WHERE to_tsvector('simple', title) @@ plainto_tsquery('simple', 'importante');

-- Consulta de batch
EXPLAIN (ANALYZE, BUFFERS, VERBOSE)
SELECT id, title, content
FROM notes
WHERE id = ANY(ARRAY[1,2,3,4,5,6,7,8,9,10]);

-- 6. Ver estadísticas de tablas
SELECT 
    schemaname,
    tablename,
    seq_scan,
    seq_tup_read,
    idx_scan,
    idx_tup_fetch,
    n_tup_ins,
    n_tup_upd,
    n_tup_del,
    n_live_tup,
    n_dead_tup
FROM pg_stat_user_tables
ORDER BY schemaname, tablename;