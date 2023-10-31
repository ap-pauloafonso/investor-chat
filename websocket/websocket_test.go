package websocket

import (
	"context"
	"encoding/json"
	"github.com/ap-pauloafonso/investor-chat/pb"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net/http/httptest"
	"nhooyr.io/websocket"
	"reflect"
	"strings"
	"testing"
	"time"
)

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

// Create a mock for the archive service
type MockArchiveService struct {
	messages []*pb.Message
	err      error
}

func (m *MockArchiveService) GetRecentMessages(_ context.Context, _ *pb.GetRecentMessagesRequest, _ ...grpc.CallOption) (*pb.GetRecentMessagesResponse, error) {
	// Return predefined messages or errors for testing
	// For example:
	return &pb.GetRecentMessagesResponse{Messages: m.messages}, nil
}

func TestUserConnected(t *testing.T) {
	archive := &MockArchiveService{
		messages: []*pb.Message{{
			Channel:   "channel1",
			User:      "randomuser",
			Text:      "text1",
			Timestamp: timestamppb.New(time.Now()),
		},
			{
				Channel:   "channel1",
				User:      "randomuser",
				Text:      "text2",
				Timestamp: timestamppb.New(time.Now().Add(1 * time.Hour)),
			}},
		err: nil,
	}

	// Create a new Handler for testing
	wH := NewWebSocketHandler(nil, archive)

	// Create an Echo instance
	e := echo.New()

	// Register the WebSocket route
	e.GET("/ws/:channel", func(c echo.Context) error {

		c.Set("username", "paulo")
		return wH.HandleRequest(c) // Call the HandleRequest method of your Handler

	})

	// Create an HTTP test server with the Echo instance
	server := httptest.NewServer(e)
	defer server.Close()

	// Construct the WebSocket URL
	reqURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/channel1"

	// Perform a regular HTTP request, not a WebSocket upgrade
	conn, _, err := websocket.Dial(context.Background(), reqURL, nil) //nolint
	if err != nil {
		t.Fatalf("Failed to connect to the WebSocket endpoint: %v", err)
	}

	_, body, err := conn.Read(context.Background())
	if err != nil {
		return
	}

	var list []payload

	err = json.Unmarshal(body, &list)
	if err != nil {
		t.Fatal(err)
	}

	var want []payload
	for _, v := range archive.messages {
		want = append(want, payload{
			Username: v.User,
			Msg:      v.Text,
			IsBot:    false,
			Time:     v.Timestamp.AsTime().Round(0),
		})
	}

	if !reflect.DeepEqual(list, want) {
		t.Fatal("got", list, "want", want)
	}

}
