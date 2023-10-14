package bot

import "testing"

func TestProcess(t *testing.T) {
	tests := []struct {
		name      string
		stockCode string
		wantErr   bool
	}{
		{
			name:      "Valid Command Code",
			stockCode: "aapl.us",
			wantErr:   false,
		},
		{
			name:      "Invalid Command Code",
			stockCode: "invalid",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetStockMessage(tt.stockCode)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetStockMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
