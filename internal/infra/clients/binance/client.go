package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nuvotlyuba/trading-engine/internal/config"
	"github.com/nuvotlyuba/trading-engine/internal/domain/candle"
	"github.com/shopspring/decimal"
)

type BinanceClient interface {
	GetKlines(ctx context.Context, symbol string, period candle.Period, limit int) ([]candle.Candle, error)
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *slog.Logger
}

func NewClient(cfg config.Config, logger *slog.Logger) *Client {
	return &Client{
		baseURL: cfg.BinanceClient.Addr,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		logger: logger,
	}
}

func (c *Client) GetKlines(ctx context.Context, symbol string, period candle.Period, limit int) ([]candle.Candle, error) {
	resp, err := c.sendRequest(
		ctx,
		http.MethodGet,
		"/api/v3/klines",
		KlinesRequest{
			Symbol:   symbol,
			Interval: convertToBinanceInterval(period),
			Limit:    limit,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	var raw [][]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	candles := make([]candle.Candle, 0, len(raw))
	for _, kline := range raw {
		k, err := parseKline(symbol, period, kline)
		if err != nil {
			c.logger.Warn("skip malformed kline", "error", err)
			continue
		}
		candles = append(candles, k)
	}

	c.logger.Debug("fetched candles from binance",
		"symbol", symbol,
		"period", period,
		"count", len(candles),
	)

	return candles, nil
}

func (c *Client) sendRequest(ctx context.Context, httpMethod, endpoint string, data any) (*http.Response, error) {
	requestURL, err := url.JoinPath(c.baseURL, endpoint)
	if err != nil {
		return nil, fmt.Errorf("join path: %w", err)
	}

	req, err := c.buildRequest(
		ctx, httpMethod,
		requestURL,
		data,
	)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"binance returned status=%d",
			resp.StatusCode,
		)
	}
	return resp, nil
}

func (c *Client) buildRequest(
	ctx context.Context,
	httpMethod string,
	requestURL string,
	data any,
) (*http.Request, error) {
	if httpMethod == http.MethodGet ||
		httpMethod == http.MethodDelete {

		query, err := buildQuery(data)
		if err != nil {
			return nil, err
		}

		if query != "" {
			requestURL += "?" + query
		}

		req, err := http.NewRequestWithContext(
			ctx,
			httpMethod,
			requestURL,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("new request: %w", err)
		}

		return req, nil
	}

	body, err := buildBody(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		httpMethod,
		requestURL,
		strings.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func buildQuery(data any) (string, error) {
	if data == nil {
		return "", nil
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("marshal query: %w", err)
	}

	var query map[string]any

	if err := json.Unmarshal(raw, &query); err != nil {
		return "", fmt.Errorf("unmarshal query: %w", err)
	}

	values := url.Values{}

	for key, value := range query {
		values.Set(key, fmt.Sprintf("%v", value))
	}

	return values.Encode(), nil
}

func buildBody(data any) (string, error) {
	if data == nil {
		return "", nil
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("marshal body: %w", err)
	}

	return string(raw), nil
}

func parseKline(symbol string, period candle.Period, kline []json.RawMessage) (candle.Candle, error) {
	if len(kline) < 6 {
		return candle.Candle{}, fmt.Errorf("kline too short: %d fields", len(kline))
	}

	var openTimeMs int64
	if err := json.Unmarshal(kline[0], &openTimeMs); err != nil {
		return candle.Candle{}, fmt.Errorf("parse open_time: %w", err)
	}

	parse := func(raw json.RawMessage, name string) (decimal.Decimal, error) {
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return decimal.Zero, fmt.Errorf("parse %s: %w", name, err)
		}
		return decimal.NewFromString(s)
	}

	open, err := parse(kline[1], "open")
	if err != nil {
		return candle.Candle{}, err
	}
	high, err := parse(kline[2], "high")
	if err != nil {
		return candle.Candle{}, err
	}
	low, err := parse(kline[3], "low")
	if err != nil {
		return candle.Candle{}, err
	}
	close, err := parse(kline[4], "close")
	if err != nil {
		return candle.Candle{}, err
	}
	volume, err := parse(kline[5], "volume")
	if err != nil {
		return candle.Candle{}, err
	}

	openTime := time.UnixMilli(openTimeMs).UTC()

	return candle.Candle{
		Symbol:    symbol,
		Period:    period,
		OpenTime:  openTime,
		CloseTime: openTime.Add(converterPeriodToDuration(period)),
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    volume,
	}, nil
}
