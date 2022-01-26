# dp-interactives-api
Manages all visualisation state and metadata (in persistent store). Talks to dp-interactives-importer via kafka.

## Getting started

* Run `make debug`

### Dependencies

* No further dependencies other than those defined in `go.mod`

### Configuration
| BIND_ADDR                    | 27500                             | The host and port to bind to
| INTERACTIVES_API_URL         | http://localhost:27500            | The interactives api url
| AWS_REGION                   | eu-west-1                         | The AWS region
| UPLOAD_BUCKET_NAME           | dp-interactives-file-uploads      | Name of the S3 bucket 
| KAFKA_ADDR                   | `localhost:9092`                  | The address of Kafka brokers (comma-separated values)
| KAFKA_VERSION                | `1.0.2`                           | The version of Kafka
| KAFKA_MAX_BYTES              | 2000000                           | Maximum number of bytes in a kafka message
| KAFKA_SEC_PROTO              | _unset_                           | if set to `TLS`, kafka connections will use TLS [1]
| KAFKA_SEC_CLIENT_KEY         | _unset_                           | PEM for the client key [1]
| KAFKA_SEC_CLIENT_CERT        | _unset_                           | PEM for the client certificate [1]
| KAFKA_SEC_CA_CERTS           | _unset_                           | CA cert chain for the server cert [1]
| KAFKA_SEC_SKIP_VERIFY        | false                             | ignores server certificate issues if `true` [1]
| MONGODB_BIND_ADDR            | localhost:27017                   | The MongoDB bind address
| MONGODB_COLLECTION           | visualisations                    | The MongoDB interactives database
| MONGODB_DATABASE             | interactives                      | MongoDB collection
| MONGODB_USERNAME             | test                              | MongoDB Username
| MONGODB_PASSWORD             | test                              | MongoDB Password
| MONGODB_IS_SSL               | false                             | is SSL enabled for mongo server
| KAFKA_CONSUMER_WORKERS       | 1                                 | The maximum number of parallel kafka consumers
| INTERACTIVES_GROUP           | dp-interactives-api               | The consumer group this application uses
| ZEBEDEE_URL                  | http://localhost:8082             | The URL of zebedee

### License

Copyright © 2021, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
