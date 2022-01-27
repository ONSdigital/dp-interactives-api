package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dp-interactives-api
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	ApiURL                     string        `envconfig:"INTERACTIVES_API_URL"`
	AwsRegion                  string        `envconfig:"AWS_REGION"`
	UploadBucketName           string        `envconfig:"UPLOAD_BUCKET_NAME"`
	Brokers                    []string      `envconfig:"KAFKA_ADDR"`
	KafkaMaxBytes              int           `envconfig:"KAFKA_MAX_BYTES"`
	KafkaVersion               string        `envconfig:"KAFKA_VERSION"`
	KafkaSecProtocol           string        `envconfig:"KAFKA_SEC_PROTO"`
	KafkaSecCACerts            string        `envconfig:"KAFKA_SEC_CA_CERTS"`
	KafkaSecClientCert         string        `envconfig:"KAFKA_SEC_CLIENT_CERT"`
	KafkaSecClientKey          string        `envconfig:"KAFKA_SEC_CLIENT_KEY"             json:"-"`
	KafkaSecSkipVerify         bool          `envconfig:"KAFKA_SEC_SKIP_VERIFY"`
	InteractivesWriteTopic     string        `envconfig:"INTERACTIVES_WRITE_TOPIC"`
	InteractivesReadTopic      string        `envconfig:"INTERACTIVES_READ_TOPIC"`
	InteractivesGroup          string        `envconfig:"INTERACTIVES_GROUP"`
	KafkaConsumerWorkers       int           `envconfig:"KAFKA_CONSUMER_WORKERS"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	ZebedeeURL                 string        `envconfig:"ZEBEDEE_URL"`
	MongoConfig                MongoConfig
}

// MongoConfig contains the config required to connect to MongoDB.
type MongoConfig struct {
	BindAddr   string `envconfig:"MONGODB_BIND_ADDR"   json:"-"`
	Collection string `envconfig:"MONGODB_COLLECTION"`
	Database   string `envconfig:"MONGODB_DATABASE"`
	Username   string `envconfig:"MONGODB_USERNAME"    json:"-"`
	Password   string `envconfig:"MONGODB_PASSWORD"    json:"-"`
	IsSSL      bool   `envconfig:"MONGODB_IS_SSL"`
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg := &Config{
		BindAddr:                   "localhost:27500",
		ApiURL:                     "http://localhost:27500",
		AwsRegion:                  "eu-west-1",
		UploadBucketName:           "dp-interactives-file-uploads",
		Brokers:                    []string{"localhost:9092", "localhost:9093", "localhost:9094"},
		KafkaVersion:               "1.0.2",
		KafkaMaxBytes:              2000000,
		InteractivesWriteTopic:     "interactives-visualisations-api",
		InteractivesReadTopic:      "interactives-visualisations-importer",
		KafkaConsumerWorkers:       1,
		InteractivesGroup:          "dp-interactives-api",
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		ZebedeeURL:                 "http://localhost:8082",
		MongoConfig: MongoConfig{
			BindAddr:   "localhost:27017",
			Collection: "visualisations",
			Database:   "interactives",
			Username:   "",
			Password:   "",
			IsSSL:      false,
		},
	}

	return cfg, envconfig.Process("", cfg)
}
