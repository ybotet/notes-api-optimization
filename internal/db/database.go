package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

// InitDB inicializa el connection pool con PostgreSQL
func InitDB() error {
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		connString = "postgres://user:pass@localhost:5432/notes"
	}

	// Configurar el pool
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return fmt.Errorf("ошибка парсинга конфигурации: %w", err)
	}

	// Ajustar parámetros del pool (valores iniciales)
	config.MaxConns = 20                     // Máximo de conexiones
	config.MinConns = 5                      // Mínimo de conexiones en reposo
	config.MaxConnLifetime = time.Hour       // Vida máxima de conexión
	config.MaxConnIdleTime = 5 * time.Minute // Tiempo máximo inactivo
	config.HealthCheckPeriod = time.Minute   // Frecuencia de health check

	// Configurar opciones de conexión
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement
	config.ConnConfig.StatementCacheCapacity = 100

	// Crear el pool
	Pool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("ошибка создания пула: %w", err)
	}

	// Verificar conexión
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Pool.Ping(ctx); err != nil {
		return fmt.Errorf("ошибка подключения к BD: %w", err)
	}

	log.Println("Пул соединений PostgreSQL инициализирован")
	return nil
}

// GetPool retorna el pool de conexiones
func GetPool() *pgxpool.Pool {
	return Pool
}

// CloseDB cierra el pool
func CloseDB() {
	if Pool != nil {
		Pool.Close()
		log.Println("Закрытый пул соединений PostgreSQL")
	}
}
