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
	{"name": "current_files", "type":["null",{"type":"array","items":"string"}]}
  ]
}`

// InteractiveUploadedEvent is the Avro schema
var InteractiveUploadedEvent = &avro.Schema{
	Definition: interactiveUploadedEvent,
}
