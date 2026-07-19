package secure

//go:generate mockgen -source=service.go -destination=mocks/service.go

// Service is an interface for auth service
type Service interface {
}
