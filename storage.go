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

type ChartData struct {
	Timestamp []time.Time
	Open      []float64
	High      []float64
	Low       []float64
	Close     []float64
	BBH       []float64
	BBM       []float64
	BBL       []float64
	StochK    []float64
	StochD    []float64
	MFI       []float64
	SMA       []float64
	BuyPrice  float64
}

func (s *Storage) GetChartData(symbol string) *ChartData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tickers[symbol]
	if !ok {
		return nil
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	ret := &ChartData{
		Timestamp: make([]time.Time, len(t.timestamp)),
		Open:      make([]float64, len(t.open)),
		High:      make([]float64, len(t.high)),
		Low:       make([]float64, len(t.low)),
		Close:     make([]float64, len(t.close)),
		BBH:       make([]float64, len(t.bbh)),
		BBM:       make([]float64, len(t.bbm)),
		BBL:       make([]float64, len(t.bbl)),
		StochK:    make([]float64, len(t.stochK)),
		StochD:    make([]float64, len(t.stochD)),
		MFI:       make([]float64, len(t.mfi)),
		SMA:       make([]float64, len(t.sma)),
		BuyPrice:  t.buyPrice,
	}
	copy(ret.Timestamp, t.timestamp)
	copy(ret.Open, t.open)
	copy(ret.High, t.high)
	copy(ret.Low, t.low)
	copy(ret.Close, t.close)
	copy(ret.BBH, t.bbh)
	copy(ret.BBM, t.bbm)
	copy(ret.BBL, t.bbl)
	copy(ret.StochK, t.stochK)
	copy(ret.StochD, t.stochD)
	copy(ret.MFI, t.mfi)
	copy(ret.SMA, t.sma)
	return ret
}

type TickerTable struct {
	Symbol   string
	BuyPrice float64
	Close    float64
	Change   float64
	Signal   Signal
}

func (s *Storage) GetTickerTable() []TickerTable {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ret := make([]TickerTable, 0, len(s.tickers))

	for s, t := range s.tickers {
		t.mu.RLock()
		defer t.mu.RUnlock()
		change := 0.0
		if t.buyPrice > 0 {
			change = t.close[len(t.close)-1]/t.buyPrice*100 - 100
		}
		if len(t.close) == 0 {
			continue
		}
		ret = append(ret, TickerTable{
			Symbol:   s,
			BuyPrice: t.buyPrice,
			Close:    t.close[len(t.close)-1],
			Change:   change,
			Signal:   t.signal,
		})
	}
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

func (s *Storage) GetStoch(symbol string) (float64, float64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tickers[symbol]
	if !ok {
		return math.NaN(), math.NaN()
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	if len(t.stochK) == 0 || len(t.stochD) == 0 {
		return math.NaN(), math.NaN()
	}
	return t.stochK[len(t.stochK)-1], t.stochD[len(t.stochD)-1]
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

func (s *Storage) GetMFI(symbol string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tickers[symbol]
	if !ok {
		return math.NaN()
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	if len(t.mfi) == 0 {
		return math.NaN()
	}
	return t.mfi[len(t.mfi)-1]
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
