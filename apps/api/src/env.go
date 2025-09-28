package main

type Env struct {
	NatsURL     string `env:"NATS_URL" envDefault:"nats://localhost:4222"`
	PostgresURL string `env:"POSTGRES_URL" envDefault:"postgresql://user:password@localhost:5432/content_guardian"`
}
