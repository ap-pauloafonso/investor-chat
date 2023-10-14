package app

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"investorchat/chat"
	"investorchat/queue"
	"investorchat/user"
	"net/http"
	"time"
)

// App represents the application instance
type App struct {
	E           *echo.Echo
	userService *user.Service
	chatService *chat.Service
	q           *queue.Queue
}

type UserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ResultMessage struct {
	Message string `json:"message"`
}

type ErrorMessage struct {
	ErrorMessage string `json:"errorMessage"`
}

type MessageObj struct {
	Username string
	Channel  string
	Message  string
	Time     time.Time
}

type BotCommandRequest struct {
	Command string
	Channel string
	Time    time.Time
}

type BotCommandResponse struct {
	GeneratedMessage string
	Channel          string
	Time             time.Time
}

func (a *App) RegisterUserHandler(c echo.Context) error {
	var u UserRequest

	if err := c.Bind(&u); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorMessage{ErrorMessage: fmt.Sprintf("Failed register user: %s", err.Error())})
	}

	if err := a.userService.Register(u.Username, u.Password); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorMessage{ErrorMessage: fmt.Sprintf("Failed register user: %s", err.Error())})

	}

	// Create a token with user information
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": u.Username,
	})

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorMessage{ErrorMessage: "Internal Server Error"})
	}

	setCookie(c, tokenString)

	return c.JSON(http.StatusOK, ResultMessage{Message: u.Username})
}

func setCookie(c echo.Context, tokenString string) {
	cookie := new(http.Cookie)
	cookie.Name = "token"
	cookie.Value = tokenString
	cookie.Path = "/"
	cookie.Expires = time.Now().Add(time.Hour * 24) // a day
	c.SetCookie(cookie)
}

func (a *App) LoginUserHandler(c echo.Context) error {
	var u UserRequest

	if err := c.Bind(&u); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorMessage{ErrorMessage: fmt.Sprintf("Failed register user: %s", err.Error())})
	}

	if err := a.userService.Login(u.Username, u.Password); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorMessage{ErrorMessage: fmt.Sprintf("Failed register user: %s", err.Error())})

	}

	// Create a token with user information
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": u.Username,
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorMessage{ErrorMessage: "Internal Server Error"})
	}

	setCookie(c, tokenString)

	return c.JSON(http.StatusOK, ResultMessage{Message: u.Username})
}

func (a *App) GetChannelsHandler(c echo.Context) error {
	type ChannelListResponse struct {
		Channels []string `json:"channels"`
	}

	channels, err := a.chatService.GetAllChannels()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorMessage{ErrorMessage: "Internal Server Error"})
	}

	// Create the response and return it
	response := ChannelListResponse{
		Channels: channels,
	}

	return c.JSON(http.StatusOK, response)

}

func (a *App) CreateChannelHandler(c echo.Context) error {

	type CreateChannelRequest struct {
		Name string `json:"name"`
	}

	var req CreateChannelRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorMessage{ErrorMessage: "Bad Request"})
	}

	// Call the service to create the channel
	err := a.chatService.CreateChannel(req.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorMessage{ErrorMessage: err.Error()})
	}

	channelConnections.mu.Lock()
	channelConnections.channels[req.Name] = listOfUserConnection{
		list: map[string]userConnection{},
	}
	channelConnections.mu.Unlock()

	err = a.q.PublishUpdateChannelsCommand()
	if err != nil {
		return err
	}
	// Return a success response
	return c.JSON(http.StatusCreated, ResultMessage{Message: "Channel created successfully"})
}

var jwtSecret = []byte("your-secret-key") // Replace with your secret key

func (a *App) loadExistingChannels() {
	channels, err := a.chatService.GetAllChannels()
	if err != nil {
		return
	}

	for _, c := range channels {
		channelConnections.channels[c] = listOfUserConnection{
			list: map[string]userConnection{},
		}
	}
}

// NewApp creates a new instance of the App
func NewApp(userService *user.Service, chatService *chat.Service, q *queue.Queue, frontendFS embed.FS) *App {
	app := &App{
		E:           echo.New(),
		userService: userService,
		chatService: chatService,
		q:           q,
	}

	app.loadExistingChannels()

	// Configure middleware
	app.E.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "[${time_rfc3339}] ${status} ${method} ${path} (${remote_ip}) ${latency_human}\n",
		Output: app.E.Logger.Output(),
	}))
	app.E.Use(middleware.Recover())

	// Set up API routes
	app.E.POST("/api/register", app.RegisterUserHandler)
	app.E.POST("/api/login", app.LoginUserHandler)
	app.E.GET("/api/channels", app.GetChannelsHandler, jwtCheck())
	app.E.POST("/api/channels", app.CreateChannelHandler, jwtCheck())
	app.E.GET("/ws/:channel", app.handleWebSocket, jwtCheck())

	// Set up Frontend routes
	var contentHandler = echo.WrapHandler(http.FileServer(http.FS(frontendFS)))
	var contentRewrite = middleware.Rewrite(map[string]string{"/app*": "/build/$1",
		"/app": "/build/index.html"})
	app.E.GET("/app/*", contentHandler, contentRewrite)

	app.E.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusSeeOther, "/app/")
	})

	err := app.q.ConsumeUserMessageCommandForWSBroadcast(func(payload []byte) error {

		var obj MessageObj
		err := json.Unmarshal(payload, &obj)
		if err != nil {
			return err
		}

		return broadcastMessage(obj.Username, obj.Channel, obj.Message, false)

	})
	if err != nil {
		log.Fatal(err)
	}

	err = app.q.ConsumeUpdateChannelsCommand(func(payload []byte) error {
		return app.handleChannelsUpdate()
	})
	if err != nil {
		log.Fatal(err)
	}

	err = app.q.ConsumeBotCommandResponse(func(payload []byte) error {

		var obj BotCommandResponse
		err := json.Unmarshal(payload, &obj)
		if err != nil {
			return err
		}

		return broadcastMessage("BOT", obj.Channel, obj.GeneratedMessage, true)

	})
	if err != nil {
		log.Fatal(err)
	}

	err = app.q.ConsumeUserMessageCommandForStorage(func(payload []byte) error {
		var obj MessageObj
		err := json.Unmarshal(payload, &obj)
		if err != nil {
			return err
		}

		err = app.chatService.SaveMessage(obj.Channel, obj.Username, obj.Message, obj.Time)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {

			for k, v := range channelConnections.channels {
				fmt.Printf("Channel: %s, users: %v \n", k, v.list)
			}
			fmt.Println()
			time.Sleep(10 * time.Second)
		}
	}()
	return app

}

func jwtCheck() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenString, err := c.Cookie("token")
			if err != nil {
				return c.JSON(http.StatusUnauthorized, ErrorMessage{ErrorMessage: "Unauthorized"})
			}

			token, err := jwt.Parse(tokenString.Value, func(token *jwt.Token) (interface{}, error) {
				// Validate the signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Invalid signing method")
				}
				return jwtSecret, nil
			})

			if err != nil {
				return c.JSON(http.StatusUnauthorized, ErrorMessage{ErrorMessage: "Unauthorized"})
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				// Extract and store the username in the context
				if username, ok := claims["username"].(string); ok {
					c.Set("username", username)
				} else {
					return c.JSON(http.StatusUnauthorized, ErrorMessage{ErrorMessage: "Unauthorized"})
				}

				return next(c)
			}

			return c.JSON(http.StatusUnauthorized, ErrorMessage{ErrorMessage: "Unauthorized"})
		}
	}
}
