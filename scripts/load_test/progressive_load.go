package loadtest

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func ProgressiveLoad() {
	url := "http://localhost:8081/api/v1/notes?limit=20"

	fmt.Println(" Прогрессивный нагрузочный тест")
	fmt.Println("URL:", url)
	fmt.Println("==============================")

	// Probar con diferentes niveles de concurrencia
	concurrencyLevels := []int{1, 5, 10, 20, 30, 40, 50}

	for _, concurrentWorkers := range concurrencyLevels {
		fmt.Printf("\n Уровень параллелизма: %d рабочие\n", concurrentWorkers)

		var successCount, errorCount int64
		var totalTime int64

		var wg sync.WaitGroup
		start := time.Now()

		requestsPerWorker := 100

		for w := 0; w < concurrentWorkers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				client := &http.Client{Timeout: 10 * time.Second}

				for i := 0; i < requestsPerWorker; i++ {
					reqStart := time.Now()

					resp, err := client.Get(url)
					reqTime := time.Since(reqStart).Milliseconds()

					atomic.AddInt64(&totalTime, reqTime)

					if err != nil || (resp != nil && resp.StatusCode >= 400) {
						atomic.AddInt64(&errorCount, 1)
					} else {
						atomic.AddInt64(&successCount, 1)
					}

					if resp != nil {
						resp.Body.Close()
					}

					// Pequeña pausa
					time.Sleep(50 * time.Millisecond)
				}
			}()
		}

		wg.Wait()
		totalDuration := time.Since(start)

		// Calcular métricas
		totalRequests := successCount + errorCount
		rps := float64(successCount) / totalDuration.Seconds()
		avgLatency := float64(totalTime) / float64(totalRequests)
		errorRate := float64(errorCount) / float64(totalRequests) * 100

		fmt.Printf("  Запросы: %d (%.0f RPS)\n", totalRequests, rps)
		fmt.Printf("  Успехи: %d, Ошибки: %d (%.1f%%)\n", successCount, errorCount, errorRate)
		fmt.Printf("  Средняя задержка: %.2f ms\n", avgLatency)
		fmt.Printf("  Общее время: %v\n", totalDuration.Round(time.Millisecond))

		// Esperar entre niveles
		time.Sleep(3 * time.Second)
	}

	fmt.Println("\n Прогрессивный нагрузочный тест завершен")
}
