package http

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"gophkeeper/internal/config"
	"gophkeeper/internal/logger"
	"gophkeeper/internal/middleware"
	"gophkeeper/pkg/jwt"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// Router
type Router struct {
	cfg        *config.WebServerConfig
	engine     *gin.Engine
	server     *http.Server
	log        *logger.Logger
	jwtManager *jwt.JWTManager
}

// NewRouter creates a new instance of Router
func NewRouter(logger *logger.Logger, cfg *config.Config, jwtManager *jwt.JWTManager) *Router {
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(
		middleware.Recovery(logger.Logger),
		middleware.LoggingMiddleware(logger.Logger),
		middleware.GzipMiddleware(),
	)
	return &Router{
		cfg:    cfg.Web,
		engine: engine,
		server: &http.Server{
			Addr:         cfg.Web.RunAddress,
			Handler:      engine,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		log:        logger,
		jwtManager: jwtManager,
	}
}

// Run starts the HTTP server
func (r *Router) Run(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		if r.cfg.EnableHTTPS {
			if err := generateTLSCertificates(r.cfg.CertFile, r.cfg.KeyFile); err != nil {
				errChan <- err
			}
			if err := r.server.ListenAndServeTLS(r.cfg.CertFile, r.cfg.KeyFile); err != nil {
				errChan <- err
			}
			return
		}
		if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return r.server.Shutdown(shutdownCtx)
	case err := <-errChan:
		return err
	}
}

// RegisterRoutes registers routes
func (r *Router) RegisterRoutes(authHandler *AuthHandler, wsHandler *WebSocketHandler, secureHandler *SecureHandler) {
	api := r.engine.Group("/api/user")
	{
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)
		auth := api.Group("/")
		auth.Use(middleware.AuthMiddleware(r.jwtManager))
		{
			auth.GET("/ws", wsHandler.WebSocketHandler)
			auth.GET("/secure", secureHandler.GetSecureHandler)
			auth.POST("/secure", secureHandler.CreateSecureHandler)
			auth.DELETE("/secure", secureHandler.DeleteSecureHandler)
			auth.PUT("/secure", secureHandler.UpdateSecureHandler)
		}
	}
}

func generateTLSCertificates(certFile, keyFile string) error {

	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			return nil
		}
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 год
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1)},
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	certFileHandle, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("failed to create cert file: %w", err)
	}
	defer func() {
		err = certFileHandle.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	if err := pem.Encode(certFileHandle, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	keyFileHandle, err := os.Create(keyFile)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer func() {
		err = keyFileHandle.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	if err := pem.Encode(keyFileHandle, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	return nil
}
