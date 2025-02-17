package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata/stream"
)

type StreamData struct {
	Symbol string
	Candle Candle
}

type Fetcher struct {
	stream        chan StreamData
	client        *marketdata.Client
	stream_client *stream.StocksClient
	mu            sync.Mutex
}

func NewFetcher(apiKey string, secretKey string) *Fetcher {
	return &Fetcher{
		stream: make(chan StreamData, 100),
		client: marketdata.NewClient(marketdata.ClientOpts{
			APIKey:    apiKey,
			APISecret: secretKey,
			Feed:      marketdata.IEX,
		}),
		stream_client: stream.NewStocksClient(marketdata.IEX, stream.WithCredentials(apiKey, secretKey), stream.WithLogger(stream.ErrorOnlyLogger())),
	}
}

func (f *Fetcher) Fetch(symbol string, start time.Time, end time.Time) ([]Candle, error) {
	bars, err := f.client.GetBars(symbol, marketdata.GetBarsRequest{
		TimeFrame:  marketdata.OneDay,
		Start:      start,
		End:        end,
		Adjustment: marketdata.All,
	})
	if err != nil {
		return nil, err
	}
	candles := make([]Candle, len(bars))
	for k, v := range bars {
		candles[k].Timestamp = v.Timestamp
		candles[k].Open = v.Open
		candles[k].High = v.High
		candles[k].Low = v.Low
		candles[k].Close = v.Close
		candles[k].Volume = float64(v.Volume)
	}
	return candles, nil
}

func (f *Fetcher) handler(bar stream.Bar) {
	f.stream <- StreamData{
		Symbol: bar.Symbol,
		Candle: Candle{
			Timestamp: bar.Timestamp,
			Open:      bar.Open,
			High:      bar.High,
			Low:       bar.Low,
			Close:     bar.Close,
			Volume:    float64(bar.Volume),
		},
	}
}

func (f *Fetcher) Connect(ctx context.Context) error {
	return f.stream_client.Connect(ctx)
}

func (f *Fetcher) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-f.stream_client.Terminated():
			return fmt.Errorf("Terminated: %v", err)
		}
	}
}

func (f *Fetcher) Sub(symbol string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.stream_client.SubscribeToDailyBars(f.handler, symbol)
}

func (f *Fetcher) Unsub(symbol string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.stream_client.UnsubscribeFromDailyBars(symbol)
}

func (f *Fetcher) Stream() <-chan StreamData {
	return f.stream
}
