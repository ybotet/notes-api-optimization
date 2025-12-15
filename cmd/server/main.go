package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ybotet/notes-api-optimization/internal/db"
	"github.com/ybotet/notes-api-optimization/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Inicializar base de datos
	if err := db.InitDB(); err != nil {
		log.Fatal("Error inicializando base de datos:", err)
	}
	defer db.CloseDB()

	// 2. Crear repositorio
	repo := db.NewPostgresRepository(db.GetPool())

	// 3. Crear handlers
	noteHandler := handlers.NewNoteHandler(repo)
	healthHandler := handlers.NewHealthHandler(db.GetPool())

	// 4. Configurar router
	router := gin.Default()

	// Middleware de logging
	router.Use(gin.Logger())

	// Middleware de recuperación
	router.Use(gin.Recovery())

	// Rutas
	api := router.Group("/api/v1")
	{
		api.GET("/health", healthHandler.HealthCheck)
		api.GET("/stats", noteHandler.GetStats)

		// CRUD de notas
		notes := api.Group("/notes")
		{
			notes.POST("", noteHandler.CreateNote)
			notes.GET("", noteHandler.ListNotes)
			notes.GET("/batch", noteHandler.GetNotesBatch)
			notes.GET("/search", noteHandler.SearchNotes)
			notes.GET("/:id", noteHandler.GetNote)
			notes.PUT("/:id", noteHandler.UpdateNote)
			notes.DELETE("/:id", noteHandler.DeleteNote)
		}
	}

	// 5. Configurar servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	// 6. Iniciar servidor en goroutine
	go func() {
		log.Printf("Servidor iniciado en http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Error iniciando servidor:", err)
		}
	}()

	// 7. Esperar señal de terminación
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Apagando servidor...")

	// 8. Apagado ordenado
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Error apagando servidor:", err)
	}

	log.Println("Servidor apagado correctamente")
}
