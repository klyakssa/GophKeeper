package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"

	"github.com/fasthttp/websocket"
	"github.com/go-resty/resty/v2"
)

// Message структура для обмена сообщениями
type Message struct {
	Type string `json:"type"`
}

// Client структура WebSocket клиента
type WsClient struct {
	conn        *websocket.Conn
	url         string
	done        chan struct{}
	messageChan chan Message
	dialer      *websocket.Dialer
	closeOnce   sync.Once
}

// NewClient создает новый экземпляр клиента
func NewWsClient(serverURL string) *WsClient {
	dialer, err := createWebSocketDialerWithCert("server.crt")
	if err != nil {
		panic(err)
	}

	return &WsClient{
		url:         serverURL,
		done:        make(chan struct{}),
		messageChan: make(chan Message, 100),
		dialer:      dialer,
	}
}

func (c *WsClient) Connect(token string) error {
	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}

	headers := make(map[string][]string)
	headers["User-Agent"] = []string{"GophKeeper-Client/1.0"}
	headers["X-Client-ID"] = []string{fmt.Sprintf("client-%d", time.Now().Unix())}
	headers["Authorization"] = []string{fmt.Sprintf("Bearer %s", token)}

	conn, resp, err := c.dialer.Dial(u.String(), headers)
	if err != nil {
		if resp != nil {
			log.Printf("HTTP response status: %s", resp.Status)
		}
		return fmt.Errorf("dial error: %v", err)
	}
	defer resp.Body.Close()

	c.conn = conn

	c.conn.SetPingHandler(func(appData string) error {
		if err := c.SendPong(); err != nil {
			log.Printf("KeepAlive ping failed: %v", err)
		}
		return c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	})
	c.conn.SetCloseHandler(c.closeHandler)

	go c.readMessages()
	go c.handleMessages()

	return nil
}

func (c *WsClient) readMessages() {
	for {
		select {
		case <-c.done:
			return
		default:
			var msg Message
			err := c.conn.ReadJSON(&msg)
			if err != nil {
				if !websocket.IsUnexpectedCloseError(err,
					websocket.CloseGoingAway,
					websocket.CloseAbnormalClosure,
					websocket.CloseNormalClosure) {
					log.Printf("Connection error: %v", err)
				}
				return
			}

			c.messageChan <- msg
		}
	}
}

func (c *WsClient) handleMessages() {
	for {
		select {
		case msg := <-c.messageChan:
			c.processMessage(msg)
		case <-c.done:
			return
		}
	}
}

func (c *WsClient) processMessage(msg Message) {
	switch msg.Type {
	case "send_message_to_update":
		printWarning("Received message to update")
	default:
		log.Printf("Received message type: %s", msg.Type)
	}
}

func (c *WsClient) SendPong() error {
	if c.conn == nil {
		return fmt.Errorf("connection is nil")
	}

	if err := c.conn.WriteMessage(websocket.PongMessage, []byte{}); err != nil {
		log.Printf("error sending ping %v", err)
		return err
	}
	return nil
}

func (c *WsClient) closeHandler(code int, text string) error {
	defer close(c.done)
	fmt.Println("WebSocket connection closed. Try to relogin.")
	return c.conn.Close()
}

func (c *WsClient) Close() error {
	var err error

	c.closeOnce.Do(func() {
		if c.done != nil {
			close(c.done)
		}

		if c.conn != nil {
			sendErr := c.conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "closing"))
			if sendErr != nil {
				log.Printf("Close message error: %v", sendErr)
			}

			err = c.conn.Close()
		}
	})

	return err
}

func createWebSocketDialerWithCert(certFile string) (*websocket.Dialer, error) {
	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read cert file: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(certPEM) {
		return nil, fmt.Errorf("failed to parse certificate")
	}

	return &websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		ReadBufferSize:   4096,
		WriteBufferSize:  4096,
		TLSClientConfig: &tls.Config{
			RootCAs: certPool,
		},
	}, nil
}

type SecureData struct {
	ID                  int64  `json:"id,omitempty"`
	DataType            string `json:"data_type"`
	Login               string `json:"login,omitempty"`
	PasswordEncrypted   []byte `json:"password_encrypted,omitempty"`
	TextData            string `json:"text_data,omitempty"`
	BinaryData          []byte `json:"binary_data,omitempty"`
	BinaryMimeType      string `json:"binary_mime_type,omitempty"`
	CardNumberEncrypted []byte `json:"card_number_encrypted,omitempty"`
	CardHolder          string `json:"card_holder,omitempty"`
	CardExpiryMonth     int16  `json:"card_expiry_month,omitempty"`
	CardExpiryYear      int16  `json:"card_expiry_year,omitempty"`
	CardCvvEncrypted    []byte `json:"card_cvv_encrypted,omitempty"`
	CardType            string `json:"card_type,omitempty"`
	Metadata            string `json:"metadata"`
	CreatedAt           string `json:"created_at,omitempty"`
	UpdatedAt           string `json:"updated_at,omitempty"`
}

type SecureDataCreate struct {
	DataType        string `json:"data_type" validate:"required,oneof=credentials text binary card"`
	Login           string `json:"login,omitempty"`
	Password        []byte `json:"password,omitempty"`
	TextData        string `json:"text_data,omitempty"`
	BinaryData      []byte `json:"binary_data,omitempty"`
	BinaryMimeType  string `json:"binary_mime_type,omitempty"`
	CardNumber      []byte `json:"card_number,omitempty"`
	CardHolder      string `json:"card_holder,omitempty"`
	CardExpiryMonth int16  `json:"card_expiry_month,omitempty"`
	CardExpiryYear  int16  `json:"card_expiry_year,omitempty"`
	CardCvv         []byte `json:"card_cvv,omitempty"`
	CardType        string `json:"card_type,omitempty"`
	Metadata        string `json:"metadata"`
}

type SecureDataUpdate struct {
	ID                  int64   `json:"id"`
	DataType            *string `json:"data_type,omitempty"`
	Login               *string `json:"login,omitempty"`
	PasswordEncrypted   []byte  `json:"password_encrypted,omitempty"`
	TextData            *string `json:"text_data,omitempty"`
	BinaryData          []byte  `json:"binary_data,omitempty"`
	BinaryMimeType      *string `json:"binary_mime_type,omitempty"`
	CardNumberEncrypted []byte  `json:"card_number_encrypted,omitempty"`
	CardHolder          *string `json:"card_holder,omitempty"`
	CardExpiryMonth     *int16  `json:"card_expiry_month,omitempty"`
	CardExpiryYear      *int16  `json:"card_expiry_year,omitempty"`
	CardCvvEncrypted    []byte  `json:"card_cvv_encrypted,omitempty"`
	CardType            *string `json:"card_type,omitempty"`
	Metadata            *string `json:"metadata,omitempty"`
}

type Client struct {
	BaseURL string
	Resty   *resty.Client
	Token   string
	Ws      *WsClient
}

func NewClient(baseURL string) *Client {
	httpClient, err := createHTTPClientWithCert("server.crt")
	if err != nil {
		panic(err)
	}

	client := resty.NewWithClient(httpClient).
		SetBaseURL(baseURL).
		SetTimeout(30*time.Second).
		SetHeader("Content-Type", "application/json").
		SetAuthScheme("Bearer").
		SetRetryCount(3).
		SetRetryWaitTime(2 * time.Second).
		SetRetryMaxWaitTime(10 * time.Second)

	wsURL := "wss://" + strings.TrimPrefix(baseURL, "https://") + "/api/user/ws"

	fmt.Printf("Ws url: %s\n", wsURL)

	return &Client{
		BaseURL: baseURL,
		Resty:   client,
		Ws:      NewWsClient(wsURL),
	}
}

const (
	TokenFile = "token.txt"
)

func (c *Client) SetToken(token string) {
	if c.Token != "" {
		if c.Ws.conn != nil {
			if err := c.Ws.Close(); err != nil {
				fmt.Printf("Failed to close WebSocket connection: %v\n", err)
				os.Exit(1)
			}
		}
	}
	c.Token = token
	c.SaveTokenToFile(TokenFile)
	if err := c.Ws.Connect(token); err != nil {
		fmt.Printf("Failed to connect to WebSocket serve: %v\n Please register first or login again\n", err)
	}
	c.Resty.SetAuthToken(token)
}

func (c *Client) SaveTokenToFile(filename string) error {
	if c.Token == "" {
		return fmt.Errorf("no token to save")
	}
	return os.WriteFile(filename, []byte(c.Token), 0600)
}

func (c *Client) LoadTokenFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return fmt.Errorf("empty token")
	}
	c.SetToken(token)
	printSuccess("Token loaded from file!")
	return nil
}

type AuthRequest struct {
	Login    string `json:"login" binding:"required"`    // login of user
	Password string `json:"password" binding:"required"` // password of user
}

type AuthResponse struct {
	Token string `json:"token"`
}

// API методы
func (c *Client) Register(login, password string) error {
	req := AuthRequest{
		Login:    login,
		Password: password,
	}

	var resp AuthResponse
	response, err := c.Resty.R().
		SetBody(req).
		SetResult(&resp).
		Post("/api/user/register")

	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	if response.StatusCode() >= 400 {
		return fmt.Errorf("registration failed with status %d: %s", response.StatusCode(), response.String())
	}

	c.SetToken(resp.Token)
	printSuccess("Registration successful! Token saved.")
	return nil
}

func (c *Client) Login(login, password string) error {
	req := AuthRequest{
		Login:    login,
		Password: password,
	}

	var resp AuthResponse
	response, err := c.Resty.R().
		SetBody(req).
		SetResult(&resp).
		Post("/api/user/login")

	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	if response.StatusCode() >= 400 {
		return fmt.Errorf("login failed with status %d: %s", response.StatusCode(), response.String())
	}

	c.SetToken(resp.Token)
	printSuccess("Login successful! Token saved.")
	return nil
}

func (c *Client) GetSecureData() error {
	var data []SecureData
	response, err := c.Resty.R().
		SetResult(&data).
		Get("/api/user/secure")

	if err != nil {
		return fmt.Errorf("failed to get secure data: %w", err)
	}

	if response.StatusCode() >= 400 {
		return fmt.Errorf("failed with status %d: %s", response.StatusCode(), response.String())
	}

	if len(data) == 0 {
		printInfo("No secure data found")
		return nil
	}

	printSuccess(fmt.Sprintf("Found %d records:", len(data)))
	fmt.Println()
	for i, item := range data {
		fmt.Printf("📋 Record #%d:\n", i+1)
		fmt.Printf("  ID: %d\n", item.ID)
		fmt.Printf("  Type: %s\n", item.DataType)
		if item.Login != "" {
			fmt.Printf("  Login: %s\n", item.Login)
		}
		if item.PasswordEncrypted != nil {
			fmt.Printf("  Password: %s\n", string(item.PasswordEncrypted))
		}
		if item.TextData != "" {
			fmt.Printf("  Text: %s\n", item.TextData)
		}
		if item.BinaryData != nil {
			fmt.Printf("  Binary Data: %s\n", item.BinaryData)
		}
		if item.BinaryMimeType != "" {
			fmt.Printf("  Binary Mime Type: %s\n", item.BinaryMimeType)
		}
		if item.CardType != "" {
			fmt.Printf("  Card Type: %s\n", item.CardType)
		}
		if item.CardNumberEncrypted != nil {
			fmt.Printf("  Card Number: %s\n", string(item.CardNumberEncrypted))
		}
		if item.CardHolder != "" {
			fmt.Printf("  Card Holder: %s\n", item.CardHolder)
		}
		if item.CardExpiryMonth != 0 {
			fmt.Printf("  Card Expiry Month: %d\n", item.CardExpiryMonth)
		}
		if item.CardExpiryYear != 0 {
			fmt.Printf("  Card Expiry Year: %d\n", item.CardExpiryYear)
		}
		if item.CardCvvEncrypted != nil {
			fmt.Printf("  Card CVV: %s\n", string(item.CardCvvEncrypted))
		}
		if item.Metadata != "" {
			fmt.Printf("  Metadata: %s\n", item.Metadata)
		}
		fmt.Printf("  Created: %s\n", item.CreatedAt)
		fmt.Printf("  Updated: %s\n", item.UpdatedAt)
		fmt.Println()
	}
	return nil
}

type APIResponse struct {
	Message string `json:"message"`
}

func (c *Client) CreateSecureData(data *SecureDataCreate) error {
	var resp APIResponse
	response, err := c.Resty.R().
		SetBody(data).
		SetResult(&resp).
		Post("/api/user/secure")

	if err != nil {
		return fmt.Errorf("failed to create secure data: %w", err)
	}

	if response.StatusCode() >= 400 {
		return fmt.Errorf("failed with status %d: %s", response.StatusCode(), response.String())
	}

	printSuccess("Secure data created successfully!")
	return nil
}

func (c *Client) UpdateSecureData(data *SecureDataUpdate) error {
	var resp APIResponse
	response, err := c.Resty.R().
		SetBody(data).
		SetResult(&resp).
		Put("/api/user/secure")

	if err != nil {
		return fmt.Errorf("failed to update secure data: %w", err)
	}

	if response.StatusCode() >= 400 {
		return fmt.Errorf("failed with status %d: %s", response.StatusCode(), response.String())
	}

	printSuccess("Secure data updated successfully!")
	return nil
}

type DeleteRequest struct {
	ID int64 `json:"id"`
}

func (c *Client) DeleteSecureData(id int64) error {
	req := DeleteRequest{ID: id}
	var resp APIResponse
	response, err := c.Resty.R().
		SetBody(req).
		SetResult(&resp).
		Delete("/api/user/secure")

	if err != nil {
		return fmt.Errorf("failed to delete secure data: %w", err)
	}

	if response.StatusCode() >= 400 {
		return fmt.Errorf("failed with status %d: %s", response.StatusCode(), response.String())
	}

	printSuccess(fmt.Sprintf("Secure data with ID %d deleted successfully!", id))
	return nil
}

func readInput(prompt string, secure bool) string {
	fmt.Print(prompt)

	if secure {
		password, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка при чтении пароля: %v\n", err)
			return ""
		}
		return strings.TrimSpace(string(password))
	}

	// Обычный ввод (с отображением)
	var input string
	fmt.Scanln(&input)
	return strings.TrimSpace(input)
}

func readIntInput(prompt string, secure bool) int64 {
	for {
		input := readInput(prompt, secure)
		var id int64
		_, err := fmt.Sscanf(input, "%d", &id)
		if err == nil {
			return id
		}
		printError("Invalid input. Please enter a number.")
	}
}

func readStringPtr(prompt string, secure bool) *string {
	input := readInput(prompt, secure)
	if input == "" {
		return nil
	}
	return &input
}

func readInt16Ptr(prompt string, secure bool) *int16 {
	input := readInput(prompt, secure)
	if input == "" {
		return nil
	}
	var val int16
	_, err := fmt.Sscanf(input, "%d", &val)
	if err != nil {
		return nil
	}
	return &val
}

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
)

func printSuccess(msg string) {
	fmt.Printf("%s✅ %s%s\n", colorGreen, msg, colorReset)
}

func printError(msg string) {
	fmt.Printf("%s❌ %s%s\n", colorRed, msg, colorReset)
}

func printInfo(msg string) {
	fmt.Printf("%sℹ️  %s%s\n", colorBlue, msg, colorReset)
}

func printWarning(msg string) {
	fmt.Printf("%s⚠️  %s%s\n", colorYellow, msg, colorReset)
}

func showMenu() {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("🔐 SECURE DATA MANAGER")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("1. Register")
	fmt.Println("2. Login")
	fmt.Println("3. Get all secure data")
	fmt.Println("4. Create secure data")
	fmt.Println("5. Update secure data")
	fmt.Println("6. Delete secure data")
	fmt.Println("7. Exit")
	fmt.Println(strings.Repeat("=", 50))
}

type TokenResp struct {
	Token string `json:"token"`
}

func main() {
	baseURL := readInput("Enter server URL (default: https://localhost:8100): ", false)
	if baseURL == "" {
		baseURL = "https://localhost:8100"
	}

	client := NewClient(baseURL)

	if err := client.LoadTokenFromFile(TokenFile); err != nil {
		printError(err.Error())
	}

	defer func() {
		_, ok := <-client.Ws.done
		if !ok {
			return
		}
		client.Ws.Close()
	}()

	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		for {
			select {
			case <-interrupt:
				log.Println("Interrupt signal received, closing connection...")
				client.Ws.Close()
				log.Println("Client stopped")
				os.Exit(0)
				return
			case <-client.Ws.done:
				log.Println("Client stopped by done signal")
				os.Exit(0)
				return
			}
		}
	}()

	for {
		showMenu()
		choice := readInput("Choose an option: ", false)

		switch choice {
		case "1":
			email := readInput("Login: ", false)
			password := readInput("Password: ", true)
			if err := client.Register(email, password); err != nil {
				printError(err.Error())
			}

		case "2":
			email := readInput("Login: ", false)
			password := readInput("Password: ", true)
			if err := client.Login(email, password); err != nil {
				printError(err.Error())
			}

		case "3":
			if client.Token == "" {
				printError("Please login first!")
				continue
			}
			if err := client.GetSecureData(); err != nil {
				printError(err.Error())
			}

		case "4":
			if client.Token == "" {
				printError("Please login first!")
				continue
			}

			data := &SecureDataCreate{}

			printInfo("Select data type:")
			fmt.Println("  1. Credentials (login/password)")
			fmt.Println("  2. Text data")
			fmt.Println("  3. Binary data")
			fmt.Println("  4. Card data")

			typeChoice := readInput("Your choice (1-4): ", false)

			switch typeChoice {
			case "1": // Credentials
				data.DataType = "credentials"
				data.Login = readInput("Login: ", false)
				password := readInput("Password: ", true)
				data.Password = []byte(password)
				data.Metadata = readInput("Metadata (optional): ", false)

			case "2": // Text data
				data.DataType = "text"
				data.TextData = readInput("Text data: ", false)
				data.Metadata = readInput("Metadata (optional): ", false)

			case "3": // Binary data
				data.DataType = "binary"
				filePath := readInput("Path to binary file: ", false)
				binaryData, err := os.ReadFile(filePath)
				if err != nil {
					printError(fmt.Sprintf("Failed to read file: %v", err))
					continue
				}
				data.BinaryData = binaryData
				data.BinaryMimeType = readInput("MIME type (e.g., image/jpeg, application/pdf): ", false)
				data.Metadata = readInput("Metadata (optional): ", false)

			case "4": // Card data
				data.DataType = "card"
				cardNumber := readInput("Card number: ", false)
				data.CardNumber = []byte(cardNumber)
				data.CardHolder = readInput("Card holder name: ", false)

				expMonth := readInput("Expiry month (MM): ", false)
				fmt.Sscanf(expMonth, "%d", &data.CardExpiryMonth)

				expYear := readInput("Expiry year (YY): ", false)
				fmt.Sscanf(expYear, "%d", &data.CardExpiryYear)

				cvv := readInput("CVV: ", true)
				data.CardCvv = []byte(cvv)
				data.CardType = readInput("Card type (Visa/Mastercard/etc): ", false)
				data.Metadata = readInput("Metadata (optional): ", false)

			default:
				printError("Invalid choice!")
				continue
			}

			if err := client.CreateSecureData(data); err != nil {
				printError(err.Error())
			}

		case "5":
			if client.Token == "" {
				printError("Please login first!")
				continue
			}

			update := &SecureDataUpdate{}
			update.ID = readIntInput("ID to update: ", false)

			printInfo("Enter new values (press Enter to skip):")
			if val := readStringPtr("Data type (credentials/text/binary/card): ", false); val != nil {
				update.DataType = val
			}
			if val := readStringPtr("Login: ", false); val != nil {
				update.Login = val
			}
			if val := readStringPtr("Password: ", true); val != nil {
				update.PasswordEncrypted = []byte(*val)
			}
			if val := readStringPtr("Binary data: ", false); val != nil {
				update.BinaryData = []byte(*val)
			}
			if val := readStringPtr("Binary MIME type: ", false); val != nil {
				update.BinaryMimeType = val
			}
			if val := readStringPtr("Text data: ", false); val != nil {
				update.TextData = val
			}
			if val := readStringPtr("Metadata: ", false); val != nil {
				update.Metadata = val
			}
			if val := readStringPtr("Card holder: ", false); val != nil {
				update.CardHolder = val
			}
			if val := readStringPtr("Card type: ", false); val != nil {
				update.CardType = val
			}
			if val := readStringPtr("Card number: ", false); val != nil {
				update.CardNumberEncrypted = []byte(*val)
			}
			if val := readStringPtr("Card CVV: ", true); val != nil {
				update.CardCvvEncrypted = []byte(*val)
			}
			if val := readInt16Ptr("Card expiry month: ", false); val != nil {
				update.CardExpiryMonth = val
			}
			if val := readInt16Ptr("Card expiry year: ", false); val != nil {
				update.CardExpiryYear = val
			}

			if err := client.UpdateSecureData(update); err != nil {
				printError(err.Error())
			}

		case "6":
			if client.Token == "" {
				printError("Please login first!")
				continue
			}
			id := readIntInput("ID to delete: ", false)
			if err := client.DeleteSecureData(id); err != nil {
				printError(err.Error())
			}

		case "7":
			printInfo("Goodbye! 👋")
			client.Ws.Close()
			return

		default:
			printError("Invalid option. Please try again.")
		}
	}
}

func createHTTPClientWithCert(certFile string) (*http.Client, error) {
	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read cert file: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(certPEM) {
		return nil, fmt.Errorf("failed to parse certificate")
	}

	tlsConfig := &tls.Config{
		RootCAs: certPool,
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: 30 * time.Second,
	}, nil
}
