package main

import (
	"context"
	"fmt"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata/stream"
)

type StreamData struct {
	Symbol string
	Candle Candle
}

type Fetcher struct {
	sub           chan string
	unsub         chan string
	stream        chan StreamData
	client        *marketdata.Client
	stream_client *stream.StocksClient
}

func NewFetcher(apiKey string, secretKey string) *Fetcher {
	return &Fetcher{
		sub:    make(chan string, 100),
		unsub:  make(chan string, 100),
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
		TimeFrame: marketdata.OneDay,
		Start:     start,
		End:       end,
	})
	if err != nil {
		return nil, err
	}
	candles := make([]Candle, len(bars))
	for _, bar := range bars {
		candles = append(candles, Candle{
			Timestamp: bar.Timestamp,
			Open:      bar.Open,
			High:      bar.High,
			Low:       bar.Low,
			Close:     bar.Close,
			Volume:    float64(bar.Volume),
		})
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

func (f *Fetcher) Run(ctx context.Context) error {
	if err := f.stream_client.Connect(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-f.stream_client.Terminated():
			return fmt.Errorf("Terminated: %v", err)
		case s := <-f.sub:
			if err := f.stream_client.SubscribeToDailyBars(f.handler, s); err != nil {
				return fmt.Errorf("SubscribeToDailyBars: %v", err)
			}
		case s := <-f.unsub:
			fmt.Println("Unsubscribing from", s)
			if err := f.stream_client.UnsubscribeFromDailyBars(s); err != nil {
				return fmt.Errorf("UnsubscribeFromDailyBars: %v", err)
			}
		}
	}
}

func (f *Fetcher) Sub(symbol string) error {
	if len(f.sub) == cap(f.sub) {
		return fmt.Errorf("sub buffer is full")
	}
	f.sub <- symbol
	return nil
}

func (f *Fetcher) Unsub(symbol string) error {
	if len(f.unsub) == cap(f.unsub) {
		return fmt.Errorf("unsub buffer is full")
	}
	f.unsub <- symbol
	return nil
}

func (f *Fetcher) Stream() chan StreamData {
	return f.stream
}
