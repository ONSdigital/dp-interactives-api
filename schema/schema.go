package schema

import (
	"github.com/ONSdigital/dp-kafka/v3/avro"
)

var interactiveUploadedEvent = `{
  "type": "record",
  "name": "interactive-uploaded",
  "fields": [
    {"name": "id", "type": "string"},
    {"name": "path", "type": "string"},
    {"name": "title", "type": "string"},
    {"name": "collection_id", "type": "string"}
  ]
}`

// InteractiveUploadedEvent is the Avro schema
var InteractiveUploadedEvent = &avro.Schema{
	Definition: interactiveUploadedEvent,
}
