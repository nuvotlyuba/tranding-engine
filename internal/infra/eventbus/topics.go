package eventbus

const (
	// Matching Engine
	TopicOrderPlaced    = "order.placed"
	TopicOrderFilled    = "order.filled"
	TopicOrderPartial   = "order.partially_filled"
	TopicOrderCancelled = "order.cancelled"
	TopicTradeExecuted  = "trade.executed"

	// TA Engine
	TopicCandleClosed    = "candle.closed"
	TopicSignalGenerated = "signal.generated"
)
