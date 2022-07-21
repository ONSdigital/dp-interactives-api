package config

import (
	"time"

	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/kelseyhightower/envconfig"
)

const (
	MetadataCollection = "MetadataCollection"
)

// Config represents service configuration for dp-interactives-api
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	PublishingEnabled          bool          `envconfig:"PUBLISHING_ENABLED"`
	PreviewRootURL             string        `envconfig:"PREVIEW_ROOT_URL"`
	AwsEndpoint                string        `envconfig:"AWS_ENDPOINT"`
	AwsRegion                  string        `envconfig:"AWS_REGION"`
	UploadBucketName           string        `envconfig:"UPLOAD_BUCKET_NAME"`
	Brokers                    []string      `envconfig:"KAFKA_ADDR"`
	MinBrokers                 int           `envconfig:"KAFKA_MIN_BROKERS"`
	KafkaMaxBytes              int           `envconfig:"KAFKA_MAX_BYTES"`
	KafkaVersion               string        `envconfig:"KAFKA_VERSION"`
	KafkaSecProtocol           string        `envconfig:"KAFKA_SEC_PROTO"`
	KafkaSecCACerts            string        `envconfig:"KAFKA_SEC_CA_CERTS"`
	KafkaSecClientCert         string        `envconfig:"KAFKA_SEC_CLIENT_CERT"`
	KafkaSecClientKey          string        `envconfig:"KAFKA_SEC_CLIENT_KEY"             json:"-"`
	KafkaSecSkipVerify         bool          `envconfig:"KAFKA_SEC_SKIP_VERIFY"`
	InteractivesWriteTopic     string        `envconfig:"INTERACTIVES_WRITE_TOPIC"`
	KafkaConsumerWorkers       int           `envconfig:"KAFKA_CONSUMER_WORKERS"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	ZebedeeURL                 string        `envconfig:"ZEBEDEE_URL"`
	FilesAPIURL                string        `envconfig:"FILES_API_URL"`
	DefaultMaxLimit            int           `envconfig:"DEFAULT_MAXIMUM_LIMIT"`
	DefaultLimit               int           `envconfig:"DEFAULT_LIMIT"`
	DefaultOffset              int           `envconfig:"DEFAULT_OFFSET"`
	ServiceAuthToken           string        `envconfig:"SERVICE_AUTH_TOKEN" 			   json:"-"`
	MongoConfig                MongoConfig
	AuthorisationConfig        *authorisation.Config
}

type MongoConfig struct {
	mongodriver.MongoDriverConfig
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	auth := authorisation.NewDefaultConfig()
	auth.Enabled = false

	cfg = &Config{
		BindAddr:                   ":27500",
		PublishingEnabled:          true,
		PreviewRootURL:             "http://localhost:27300",
		AwsRegion:                  "eu-west-1",
		UploadBucketName:           "dp-interactives-file-uploads",
		Brokers:                    []string{"localhost:9093"},
		MinBrokers:                 1,
		KafkaVersion:               "1.0.2",
		KafkaMaxBytes:              2000000,
		InteractivesWriteTopic:     "interactives-import",
		KafkaConsumerWorkers:       1,
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		ZebedeeURL:                 "http://localhost:8082",
		FilesAPIURL:                "http://localhost:26900",
		DefaultMaxLimit:            100,
		DefaultLimit:               20,
		DefaultOffset:              0,
		MongoConfig: MongoConfig{
			MongoDriverConfig: mongodriver.MongoDriverConfig{
				ClusterEndpoint:               "localhost:27017",
				Username:                      "",
				Password:                      "",
				Database:                      "interactives",
				Collections:                   map[string]string{MetadataCollection: "metadata"},
				ReplicaSet:                    "",
				IsStrongReadConcernEnabled:    false,
				IsWriteConcernMajorityEnabled: true,
				ConnectTimeout:                5 * time.Second,
				QueryTimeout:                  15 * time.Second,
				TLSConnectionConfig: mongodriver.TLSConnectionConfig{
					IsSSL: false,
				},
			},
		},
		AuthorisationConfig: auth,
	}

	return cfg, envconfig.Process("", cfg)
}
