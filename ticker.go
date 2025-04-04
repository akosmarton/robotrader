package main

import (
	"cmp"
	"encoding/json"
	"slices"
	"sync"
	"time"

	"github.com/markcheno/go-talib"
)

const (
	KEEP = 500
)

type Candle struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

type Ticker struct {
	mu sync.RWMutex

	symbol   string
	buyPrice float64

	timestamp []time.Time
	open      []float64
	high      []float64
	low       []float64
	close     []float64
	volume    []float64

	sma        []float64
	stochK     []float64
	stochD     []float64
	rsi        []float64
	macd       []float64
	macdSignal []float64
	macdHist   []float64
	bbh        []float64
	bbm        []float64
	bbl        []float64
	mfi        []float64
	adx        []float64

	signal Signal
}

func NewTicker(symbol string, buyPrice float64) *Ticker {
	return &Ticker{
		symbol:     symbol,
		buyPrice:   buyPrice,
		timestamp:  []time.Time{},
		open:       []float64{},
		high:       []float64{},
		low:        []float64{},
		close:      []float64{},
		volume:     []float64{},
		sma:        []float64{},
		rsi:        []float64{},
		macd:       []float64{},
		macdSignal: []float64{},
		macdHist:   []float64{},
		bbh:        []float64{},
		bbm:        []float64{},
		bbl:        []float64{},
		mfi:        []float64{},
		adx:        []float64{},
		stochK:     []float64{},
		stochD:     []float64{},
		signal:     SignalHold,
	}
}

func (t *Ticker) calc() Signal {
	if len(t.close) < 30 {
		return SignalHold
	}
	t.sma = talib.Sma(t.close, 200)
	t.rsi = talib.Rsi(t.close, 14)
	t.macd, t.macdSignal, t.macdHist = talib.Macd(t.close, 12, 26, 9)
	t.bbh, t.bbm, t.bbl = talib.BBands(t.close, 50, 2.5, 2.5, talib.SMA)
	t.stochK, t.stochD = talib.Stoch(t.high, t.low, t.close, 14, 3, talib.SMA, 3, talib.SMA)
	t.mfi = talib.Mfi(t.high, t.low, t.close, t.volume, 14)
	t.adx = talib.Adx(t.high, t.low, t.close, 14)
	lastSignal := t.signal

	i := len(t.close) - 1

	// Not enough data
	if t.bbh[i] == 0 || t.bbl[i] == 0 || t.close[i] == 0 || t.stochD[i] == 0 || t.stochK[i] == 0 {
		return SignalHold
	}

	// Algorithm
	if t.close[i]*1.01 > t.bbh[i] && t.adx[i] > 25 && t.mfi[i] < 30 {
		t.signal = SignalSell
	} else if t.close[i]*0.99 < t.bbl[i] && t.adx[i] > 25 && t.mfi[i] > 70 {
		t.signal = SignalBuy
	} else {
		t.signal = SignalHold
	}

	// Update only if signal changed
	if t.signal != SignalHold && lastSignal != t.signal {
		return t.signal
	}

	return SignalHold
}

func (t *Ticker) Insert(candle ...Candle) Signal {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, c := range candle {
		n, found := slices.BinarySearchFunc(t.timestamp, c.Timestamp, func(a, b time.Time) int {
			return cmp.Compare(a.Unix(), b.Unix())
		})

		if found {
			t.timestamp[n] = c.Timestamp
			t.open[n] = c.Open
			t.high[n] = c.High
			t.low[n] = c.Low
			t.close[n] = c.Close
			t.volume[n] = c.Volume
		} else {
			t.timestamp = slices.Insert(t.timestamp, n, c.Timestamp)
			t.open = slices.Insert(t.open, n, c.Open)
			t.high = slices.Insert(t.high, n, c.High)
			t.low = slices.Insert(t.low, n, c.Low)
			t.close = slices.Insert(t.close, n, c.Close)
			t.volume = slices.Insert(t.volume, n, c.Volume)
		}
	}
	t.keep(KEEP)
	return t.calc()
}

func (t *Ticker) keep(number int) {
	if len(t.timestamp) > number {
		t.timestamp = t.timestamp[len(t.timestamp)-number:]
		t.open = t.open[len(t.open)-number:]
		t.high = t.high[len(t.high)-number:]
		t.low = t.low[len(t.low)-number:]
		t.close = t.close[len(t.close)-number:]
		t.volume = t.volume[len(t.volume)-number:]
	}
}

type Signal string

const (
	SignalBuy  Signal = "buy"
	SignalHold Signal = ""
	SignalSell Signal = "sell"
)

func (t *Ticker) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		BuyPrice float64 `json:"buyPrice"`
	}{
		BuyPrice: t.buyPrice,
	})
}

func (t *Ticker) UnmarshalJSON(input []byte) error {
	data := struct {
		BuyPrice float64 `json:"buyPrice"`
	}{}
	if err := json.Unmarshal(input, &data); err != nil {
		return err
	}
	t.buyPrice = data.BuyPrice
	return nil
}
