package binance

type KlinesRequest struct {
	Symbol   string `json:"symbol"`
	Interval string `json:"interval"`
	Limit    int    `json:"limit"`
}
