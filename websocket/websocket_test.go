package websocket

import "testing"

func TestCheckBot(t *testing.T) {
	testCases := []struct {
		input string
		want  bool
	}{
		{"/stock=aapl.us", true},
		{"/stock=msft.us", true},
		{"/stocks=aapl.us", false},
		{"/price=123.asdasdas", false},
		{"this is a normal message", false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got, _ := checkBot(tc.input)
			if got != tc.want {
				t.Errorf("input: %s, got: %v, want: %v", tc.input, got, tc.want)
			}
		})
	}
}
