package server

import (
	"context"
	"embed"
	"fmt"
	"github.com/ap-pauloafonso/investor-chat/channel"
	"github.com/ap-pauloafonso/investor-chat/eventbus"
	"github.com/ap-pauloafonso/investor-chat/user"
	"github.com/ap-pauloafonso/investor-chat/utils"
	"github.com/ap-pauloafonso/investor-chat/websocket"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"time"
)

var (
	jwtSecret = []byte("my-secret-key")
)

// Server represents the application instance
type Server struct {
	E                *echo.Echo
	userService      *user.Service
	channelService   *channel.Service
	q                *eventbus.Eventbus
	webSocketHandler *websocket.Handler
}

type UserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ResultMessage struct {
	Message string `json:"message"`
}

func (s *Server) RegisterUserHandler(c echo.Context) error {
	var u UserRequest

	if err := c.Bind(&u); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorMessage{ErrorMessage: fmt.Sprintf("Failed register user: %s", err.Error())})
	}

	if err := s.userService.Register(c.Request().Context(), u.Username, u.Password); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorMessage{ErrorMessage: fmt.Sprintf("Failed register user: %s", err.Error())})

	}

	// Create a token with user information
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": u.Username,
	})

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorMessage{ErrorMessage: "Internal Server Error"})
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

func (s *Server) LoginUserHandler(c echo.Context) error {
	var u UserRequest

	if err := c.Bind(&u); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorMessage{ErrorMessage: fmt.Sprintf("Failed register user: %s", err.Error())})
	}

	if err := s.userService.Login(c.Request().Context(), u.Username, u.Password); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorMessage{ErrorMessage: fmt.Sprintf("Failed register user: %s", err.Error())})

	}

	// Create a token with user information
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": u.Username,
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorMessage{ErrorMessage: "Internal Server Error"})
	}

	setCookie(c, tokenString)

	return c.JSON(http.StatusOK, ResultMessage{Message: u.Username})
}

func (s *Server) GetChannelsHandler(c echo.Context) error {
	type ChannelListResponse struct {
		Channels []string `json:"channels"`
	}

	channels, err := s.channelService.GetAllChannels(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorMessage{ErrorMessage: "Internal Server Error"})
	}

	// Create the response and return it
	response := ChannelListResponse{
		Channels: channels,
	}

	return c.JSON(http.StatusOK, response)

}

func (s *Server) CreateChannelHandler(c echo.Context) error {

	type CreateChannelRequest struct {
		Name string `json:"name"`
	}

	var req CreateChannelRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorMessage{ErrorMessage: "Bad Request"})
	}

	// Call the service to create the channel
	err := s.channelService.CreateChannel(c.Request().Context(), req.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorMessage{ErrorMessage: err.Error()})
	}

	// Return a success response
	return c.JSON(http.StatusCreated, ResultMessage{Message: "Channel created successfully"})
}

// NewApp creates a new instance of the Server
func NewApp(ctx context.Context, userService *user.Service, channelService *channel.Service, q *eventbus.Eventbus, frontendFS embed.FS, webSocketHandler *websocket.Handler) *Server {
	server := &Server{
		E:                echo.New(),
		userService:      userService,
		channelService:   channelService,
		q:                q,
		webSocketHandler: webSocketHandler,
	}

	server.E.Use(middleware.Recover())

	// Set up API routes
	server.E.POST("/api/register", server.RegisterUserHandler)
	server.E.POST("/api/login", server.LoginUserHandler)
	server.E.GET("/api/channels", server.GetChannelsHandler, jwtCheck())
	server.E.POST("/api/channels", server.CreateChannelHandler, jwtCheck())
	server.E.GET("/ws/:channel", server.webSocketHandler.HandleRequest, jwtCheck())
	server.E.GET("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	// Set up Frontend routes
	var contentHandler = echo.WrapHandler(http.FileServer(http.FS(frontendFS)))
	var contentRewrite = middleware.Rewrite(map[string]string{"/app*": "/build/$1",
		"/app": "/build/index.html"})
	server.E.GET("/app/*", contentHandler, contentRewrite)

	server.E.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusSeeOther, "/app/")
	})

	server.initConsumers(ctx)

	return server

}

func jwtCheck() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenString, err := c.Cookie("token")
			if err != nil {
				return c.JSON(http.StatusUnauthorized, utils.ErrorMessage{ErrorMessage: "Unauthorized"})
			}

			token, err := jwt.Parse(tokenString.Value, func(token *jwt.Token) (interface{}, error) {

				// Validate the signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("invalid signing method")
				}
				return jwtSecret, nil
			})

			if err != nil {
				return c.JSON(http.StatusUnauthorized, utils.ErrorMessage{ErrorMessage: "Unauthorized"})
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				// Extract and store the username in the context
				if username, ok := claims["username"].(string); ok {
					c.Set("username", username)
				} else {
					return c.JSON(http.StatusUnauthorized, utils.ErrorMessage{ErrorMessage: "Unauthorized"})
				}

				return next(c)
			}

			return c.JSON(http.StatusUnauthorized, utils.ErrorMessage{ErrorMessage: "Unauthorized"})
		}
	}
}
