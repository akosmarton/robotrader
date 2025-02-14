package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
)

const (
	DAYS = 180
)

func main() {
	log.SetFlags(0)
	alpacaApiKey := os.Getenv("ALPACA_API_KEY")
	alpacaApiSecret := os.Getenv("ALPACA_API_SECRET")
	if alpacaApiKey == "" || alpacaApiSecret == "" {
		log.Fatal("ALPACA_API_KEY or ALPACA_API_SECRET is not set")
	}

	matrixHomeserver := os.Getenv("MATRIX_HOMESERVER")
	matrixUserId := os.Getenv("MATRIX_USER_ID")
	matrixAccessToken := os.Getenv("MATRIX_ACCESS_TOKEN")
	matrixRoomId := os.Getenv("MATRIX_ROOM_ID")
	if matrixHomeserver == "" || matrixUserId == "" || matrixAccessToken == "" || matrixRoomId == "" {
		log.Fatal("MATRIX_HOMESERVER, MATRIX_USER_ID, MATRIX_ACCESS_TOKEN or MATRIX_ROOM_ID is not set")
	}

	storageDir := os.Getenv("STORAGE_DIR")
	if storageDir == "" {
		storageDir = "."
	} else {
		storageDir = strings.TrimRight(storageDir, "/")
	}

	storage := NewStorage()
	if err := storage.Open(storageDir + "/tickers.json"); err != nil {
		log.Fatalf("Failed to open storage: %v", err)
	}
	defer storage.Close()

	fetcher := NewFetcher(alpacaApiKey, alpacaApiSecret)
	bot := NewBot(matrixHomeserver, matrixUserId, matrixAccessToken, matrixRoomId)
	if bot == nil {
		log.Fatal("Failed to create bot")
	}
	log.SetOutput(io.MultiWriter(os.Stdout, bot))

	log.Print("Starting...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := fetcher.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to fetcher: %v", err)
	}

	// Worker pool for fetching history candles
	jobs := make(chan string, 10)
	wg := sync.WaitGroup{}
	for w := 0; w < 10; w++ {
		wg.Add(1)
		go func(symbols <-chan string) {
			defer wg.Done()
			for symbol := range symbols {
				candles, err := fetcher.Fetch(symbol, time.Now().AddDate(0, 0, -DAYS), time.Now())
				if err != nil {
					log.Printf("Failed to fetch candles for %s: %v", symbol, err)
				} else if len(candles) == 0 {
					log.Printf("No candles fetched for %s", symbol)
				} else {
					storage.InsertCandles(symbol, candles...)
					if err := fetcher.Sub(symbol); err != nil {
						log.Printf("Failed to subscribe to %s: %v", symbol, err)
					}
				}
			}
		}(jobs)
	}
	log.Print("Fetching history candles...")
	// Send jobs
	symbols := storage.GetSymbols()
	for _, symbol := range symbols {
		jobs <- symbol
	}
	close(jobs)
	// Wait for workers
	wg.Wait()

	go func() {
		if err := fetcher.Run(ctx); err != nil && err != context.Canceled {
			log.Fatalf("Failed to run fetcher: %v", err)
		}
	}()

	go func() {
		if err := bot.Run(ctx); err != nil && err != context.Canceled {
			log.Fatalf("Failed to run bot: %v", err)
		}
	}()

	log.Println("Ready")
	defer log.Println("Stopped")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	for {
		select {
		case <-shutdown:
			cancel()
		case <-ctx.Done():
			return
		case msg := <-bot.Message():
			s := strings.Split(msg, " ")
			switch s[0] {
			case "help": // Help
				bot.SendText("Commands: add <symbol> [buy price], rm <symbol>, ls, mem, stop")
			case "stop": // Stop the bot
				cancel()
			case "add": // Add ticker
				if len(s) < 2 {
					continue
				}
				symbol := strings.ToUpper(s[1])
				buyPrice := 0.0
				if len(s) > 2 {
					buyPrice, _ = strconv.ParseFloat(s[2], 64)
				}
				candles, _ := fetcher.Fetch(symbol, time.Now().AddDate(0, 0, -DAYS), time.Now())
				if len(candles) == 0 {
					log.Printf("Failed to fetch candles for %s", symbol)
					continue
				}
				storage.AddTicker(symbol, buyPrice)
				storage.InsertCandles(symbol, candles...)
				fetcher.Sub(symbol)
			case "rm": // Remove ticker
				if len(s) < 2 {
					continue
				}
				symbol := strings.ToUpper(s[1])
				fetcher.Unsub(symbol)
				storage.DelTicker(symbol)
			case "ls": // List tickers
				w := table.NewWriter()
				w.Style().Options.DrawBorder = false
				w.AppendHeader(table.Row{"Symbol", "Buy Price", "Close", "Change", "Signal"})
				for _, symbol := range storage.GetSymbols() {
					var buyPriceStr, closeStr, changeStr, signalStr string
					if buyPrice := storage.GetBuyPrice(symbol); buyPrice > 0 {
						buyPriceStr = fmt.Sprintf("$%.02f", buyPrice)
						changeStr = fmt.Sprintf("%+.02f%%", storage.GetChange(symbol))
					}
					closeStr = fmt.Sprintf("$%.2f", storage.GetClose(symbol))
					signalStr = string(storage.GetSignal(symbol))
					w.AppendRow([]interface{}{symbol, buyPriceStr, closeStr, changeStr, signalStr})
				}
				bot.SendCode(w.Render())
			case "mem": // Print memory stats
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				msg := fmt.Sprintf("Alloc = %v MiB\nTotalAlloc = %v MiB\nSys = %v MiB\nNumGC = %v", m.Alloc/1024/1024, m.TotalAlloc/1024/1024, m.Sys/1024/1024, m.NumGC)
				bot.SendCode(msg)
			default: // Unknown command
				bot.SendText("Unknown command")
			}
		case d := <-fetcher.Stream():
			signal := storage.InsertCandles(d.Symbol, d.Candle)
			if signal == SignalSell && storage.GetBuyPrice(d.Symbol) > 0 {
				msg := fmt.Sprintf("%s %s %+.02f", signal, d.Symbol, storage.GetChange(d.Symbol))
				bot.SendText(msg)
			} else if signal == SignalBuy {
				msg := fmt.Sprintf("%s %s", signal, d.Symbol)
				if storage.GetBuyPrice(d.Symbol) > 0 {
					msg += fmt.Sprintf(" %+.02f%%", storage.GetChange(d.Symbol))
				}
				bot.SendText(msg)
			}
		}
	}
}
