package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
)

const (
	DAYS = 180
)

func main() {
	alpacaApiKey := os.Getenv("ALPACA_API_KEY")
	alpacaApiSecret := os.Getenv("ALPACA_API_SECRET")
	if alpacaApiKey == "" || alpacaApiSecret == "" {
		panic("ALPACA_API_KEY or ALPACA_API_SECRET is not set")
	}

	matrixHomeserver := os.Getenv("MATRIX_HOMESERVER")
	matrixUserId := os.Getenv("MATRIX_USER_ID")
	matrixAccessToken := os.Getenv("MATRIX_ACCESS_TOKEN")
	matrixRoomId := os.Getenv("MATRIX_ROOM_ID")
	if matrixHomeserver == "" || matrixUserId == "" || matrixAccessToken == "" || matrixRoomId == "" {
		panic("MATRIX_HOMESERVER, MATRIX_USER_ID, MATRIX_ACCESS_TOKEN or MATRIX_ROOM_ID is not set")
	}

	storageDir := os.Getenv("STORAGE_DIR")
	if storageDir == "" {
		storageDir = "."
	} else {
		storageDir = strings.TrimRight(storageDir, "/")
	}

	storage := NewStorage()
	if err := storage.Open(storageDir + "/tickers.json"); err != nil {
		panic(err)
	}
	defer storage.Close()

	fetcher := NewFetcher(alpacaApiKey, alpacaApiSecret)
	bot := NewBot(matrixHomeserver, matrixUserId, matrixAccessToken, matrixRoomId)
	if bot == nil {
		panic("Failed to create bot")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := fetcher.Run(ctx); err != nil && err != context.Canceled {
			panic(err)
		}
	}()

	go func() {
		if err := bot.Run(ctx); err != nil && err != context.Canceled {
			panic(err)
		}
	}()

	// Worker pool for fetching history candles
	symbols := storage.GetSymbols()
	jobs := make(chan string, len(symbols))
	results := make(chan string, len(symbols))
	for w := 0; w < 10; w++ {
		go func(symbols <-chan string, result chan<- string) {
			for symbol := range symbols {
				candles, err := fetcher.Fetch(symbol, time.Now().AddDate(0, 0, -DAYS), time.Now())
				if err == nil {
					storage.InsertCandles(symbol, candles...)
					fetcher.Sub(symbol)
				}
				result <- symbol
			}
		}(jobs, results)
	}
	// Send jobs
	for _, symbol := range symbols {
		jobs <- symbol
	}
	// Wait for jobs
	for range symbols {
		<-results
	}

	fmt.Println("Started")
	bot.SendText("Started")
	defer bot.SendText("Stopped")
	defer fmt.Println("Stopped")
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
			case "stop":
				cancel()
			case "add":
				if len(s) < 2 {
					continue
				}
				symbol := strings.ToUpper(s[1])
				buyPrice := 0.0
				if len(s) > 2 {
					buyPrice, _ = strconv.ParseFloat(s[2], 64)
				}
				storage.AddTicker(symbol, buyPrice)
				candles, _ := fetcher.Fetch(symbol, time.Now().AddDate(0, 0, -DAYS), time.Now())
				storage.InsertCandles(symbol, candles...)
				fetcher.Sub(symbol)
			case "list":
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
			case "mem":
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				msg := fmt.Sprintf("Alloc = %v MiB\nTotalAlloc = %v MiB\nSys = %v MiB\nNumGC = %v", m.Alloc/1024/1024, m.TotalAlloc/1024/1024, m.Sys/1024/1024, m.NumGC)
				bot.SendCode(msg)
			}
		case d := <-fetcher.Stream():
			signal := storage.InsertCandles(d.Symbol, d.Candle)
			if signal == SignalSell && storage.GetBuyPrice(d.Symbol) > 0 {
				msg := fmt.Sprintf("%s %s %+.02f", signal, d.Symbol, storage.GetChange(d.Symbol))
				bot.SendText(msg)
			} else if signal == SignalBuy {
				msg := fmt.Sprintf("%s %s", signal, d.Symbol)
				if storage.GetBuyPrice(d.Symbol) > 0 {
					msg += fmt.Sprintf(" %+.02f", storage.GetChange(d.Symbol))
				}
				bot.SendText(msg)
			}
		}
	}
}
