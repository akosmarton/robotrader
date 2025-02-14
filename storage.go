package main

import (
	"encoding/json"
	"math"
	"os"
	"sort"
	"sync"
	"time"
)

type Storage struct {
	tickers map[string]*Ticker
	mu      sync.RWMutex
	f       *os.File
}

func NewStorage() *Storage {
	return &Storage{
		tickers: map[string]*Ticker{},
	}
}

func (s *Storage) Open(filename string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	s.f, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	return s.load()
}

func (s *Storage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.f != nil {
		err := s.f.Close()
		s.f = nil
		return err
	}
	return nil
}

func (s *Storage) AddTicker(symbol string, buyPrice float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tickers[symbol] = NewTicker(symbol, buyPrice)
	return s.save()
}

func (s *Storage) DelTicker(symbol string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tickers, symbol)
	return s.save()
}

func (s *Storage) InsertCandles(symbol string, candles ...Candle) Signal {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tickers[symbol].Insert(candles...)
}

func (s *Storage) GetAllTimestamp(symbol string) []time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tickers[symbol]
	if !ok {
		return nil
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	ret := make([]time.Time, len(t.timestamp))
	copy(ret, t.timestamp)
	return ret
}

func (s *Storage) GetAllClose(symbol string) []float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tickers[symbol]
	if !ok {
		return nil
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	ret := make([]float64, len(t.close))
	copy(ret, t.close)
	return ret
}

func (s *Storage) GetClose(symbol string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tickers[symbol]
	if !ok {
		return math.NaN()
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	if len(t.close) == 0 {
		return math.NaN()
	}
	return t.close[len(t.close)-1]
}

func (s *Storage) GetSignal(symbol string) Signal {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tickers[symbol]
	if !ok {
		return SignalHold
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.signal
}

func (s *Storage) GetChange(symbol string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tickers[symbol]
	if !ok {
		return math.NaN()
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	if len(t.close) == 0 || t.buyPrice == math.NaN() {
		return math.NaN()
	}
	return t.close[len(t.close)-1]/t.buyPrice*100 - 100
}

func (s *Storage) GetBuyPrice(symbol string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tickers[symbol]
	if !ok {
		return math.NaN()
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.buyPrice
}

func (s *Storage) GetBB(symbol string) (float64, float64, float64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tickers[symbol]
	if !ok {
		return math.NaN(), math.NaN(), math.NaN()
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	if len(t.bbh) == 0 || len(t.bbm) == 0 || len(t.bbl) == 0 {
		return math.NaN(), math.NaN(), math.NaN()
	}
	return t.bbh[len(t.bbh)-1], t.bbm[len(t.bbm)-1], t.bbl[len(t.bbl)-1]
}

func (s *Storage) GetSMA(symbol string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tickers[symbol]
	if !ok {
		return math.NaN()
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	if len(t.sma) == 0 {
		return math.NaN()
	}
	return t.sma[len(t.sma)-1]
}

func (s *Storage) GetRSI(symbol string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tickers[symbol]
	if !ok {
		return math.NaN()
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	if len(t.rsi) == 0 {
		return math.NaN()
	}
	return t.rsi[len(t.rsi)-1]
}

func (s *Storage) GetSymbols() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	symbols := make([]string, 0, len(s.tickers))
	for k := range s.tickers {
		symbols = append(symbols, k)
	}
	sort.Strings(symbols)
	return symbols
}

func (s *Storage) save() error {
	if s.f == nil {
		return nil
	}
	d := map[string]float64{}
	for k, v := range s.tickers {
		d[k] = v.buyPrice
	}
	if _, err := s.f.Seek(0, 0); err != nil {
		return err
	}
	if err := s.f.Truncate(0); err != nil {
		return err
	}
	return json.NewEncoder(s.f).Encode(s.tickers)
}

func (s *Storage) load() error {
	if s.f == nil {
		return nil
	}
	s.tickers = map[string]*Ticker{}
	json.NewDecoder(s.f).Decode(&s.tickers)
	return nil
}
