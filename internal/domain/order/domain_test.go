package order

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestFill(t *testing.T) {
	tests := []struct {
		name       string
		order      *Order
		fillQty    decimal.Decimal
		wantErr    bool
		wantFilled decimal.Decimal
		wantStatus OrderStatus
	}{
		{
			name:       "partial fill",
			order:      NewOrder("BTCUSDT", SideBuy, decimal.NewFromInt(100), decimal.NewFromInt(10), OrderTypeLimit),
			fillQty:    decimal.NewFromInt(3),
			wantErr:    false,
			wantFilled: decimal.NewFromInt(3),
			wantStatus: StatusPartiallyFilled,
		},
		{
			name:       "full fill",
			order:      NewOrder("BTCUSDT", SideBuy, decimal.NewFromInt(100), decimal.NewFromInt(10), OrderTypeLimit),
			fillQty:    decimal.NewFromInt(10),
			wantErr:    false,
			wantFilled: decimal.NewFromInt(10),
			wantStatus: StatusFilled,
		},
		{
			name:    "zero qty returns error",
			order:   NewOrder("BTCUSDT", SideBuy, decimal.NewFromInt(100), decimal.NewFromInt(10), OrderTypeLimit),
			fillQty: decimal.Zero,
			wantErr: true,
		},
		{
			name:    "exceeds remaining returns error",
			order:   NewOrder("BTCUSDT", SideBuy, decimal.NewFromInt(100), decimal.NewFromInt(10), OrderTypeLimit),
			fillQty: decimal.NewFromInt(11),
			wantErr: true,
		},
		{
			name:    "negative qty returns error",
			order:   NewOrder("BTCUSDT", SideBuy, decimal.NewFromInt(100), decimal.NewFromInt(10), OrderTypeLimit),
			fillQty: decimal.NewFromInt(-1),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.order.Fill(tt.fillQty)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fill() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !tt.order.Filled.Equal(tt.wantFilled) {
					t.Errorf("Filled = %s, want %s", tt.order.Filled, tt.wantFilled)
				}
				if tt.order.Status != tt.wantStatus {
					t.Errorf("Status = %s, want %s", tt.order.Status, tt.wantStatus)
				}
			}
		})
	}
}

func TestRemaining(t *testing.T) {
	tests := []struct {
		name             string
		order            *Order
		filled           decimal.Decimal
		wantRemainingQty decimal.Decimal
	}{
		{
			name:             "новый ордер (Filled=0)",
			order:            NewOrder("BTCUSDT", SideBuy, decimal.NewFromInt(100), decimal.NewFromInt(10), OrderTypeLimit),
			wantRemainingQty: decimal.NewFromInt(10),
			filled:           decimal.NewFromInt(0),
		},
		{
			name:             "частично исполненный",
			order:            NewOrder("BTCUSDT", SideBuy, decimal.NewFromInt(100), decimal.NewFromInt(10), OrderTypeLimit),
			wantRemainingQty: decimal.NewFromInt(5),
			filled:           decimal.NewFromInt(5),
		},
		{
			name:             "полностью исполненный",
			order:            NewOrder("BTCUSDT", SideBuy, decimal.NewFromInt(100), decimal.NewFromInt(10), OrderTypeLimit),
			wantRemainingQty: decimal.NewFromInt(0),
			filled:           decimal.NewFromInt(10),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.order.Filled = tt.filled
			remainingQty := tt.order.Remaining()
			if !remainingQty.Equal(tt.wantRemainingQty) {
				t.Errorf("Remaining() = %s, want %s", remainingQty, tt.wantRemainingQty)
			}
		})
	}
}

func TestCancel(t *testing.T) {
	tests := []struct {
		name       string
		order      *Order
		wantStatus OrderStatus
	}{
		{
			name:       "статус стал StatusCancelled",
			order:      NewOrder("BTCUSDT", SideBuy, decimal.NewFromInt(100), decimal.NewFromInt(10), OrderTypeLimit),
			wantStatus: StatusCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.order.Cancel()
			if tt.order.Status != StatusCancelled {
				t.Errorf("Cancel() = %s, want %s", tt.order.Status, tt.wantStatus)
			}
		})
	}
}

func TestIsFilled(t *testing.T) {
	tests := []struct {
		name         string
		filledQty    decimal.Decimal
		order        *Order
		wantIsFilled bool
	}{
		{
			name:         "исполнен",
			filledQty:    decimal.NewFromInt(10),
			order:        NewOrder("BTCUSDT", SideBuy, decimal.NewFromInt(100), decimal.NewFromInt(10), OrderTypeLimit),
			wantIsFilled: true,
		},
		{
			name:         "не исполнен",
			filledQty:    decimal.NewFromInt(0),
			order:        NewOrder("BTCUSDT", SideBuy, decimal.NewFromInt(100), decimal.NewFromInt(10), OrderTypeLimit),
			wantIsFilled: false,
		},
		{
			name:         "частично исполнен",
			filledQty:    decimal.NewFromInt(5),
			order:        NewOrder("BTCUSDT", SideBuy, decimal.NewFromInt(100), decimal.NewFromInt(10), OrderTypeLimit),
			wantIsFilled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.order.Filled = tt.filledQty
			isFilled := tt.order.IsFilled()

			if isFilled != tt.wantIsFilled {
				t.Errorf("IsFilled() = %v, want %v", isFilled, tt.wantIsFilled)
			}
		})
	}
}
