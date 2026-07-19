package secure

//go:generate mockgen -source=repository.go -destination=mocks/repository.go

// Repository is an interface for secure repository
type Repository interface {
}
