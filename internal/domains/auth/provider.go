package auth

type Provider string

const (
	ProviderGoogle Provider = "google"
)

type Identity struct {
	Provider    Provider
	ProviderUID string // Google 'sub'
	Email       string
	Name        string
}
