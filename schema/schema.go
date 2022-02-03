package schema

import (
	"github.com/ONSdigital/go-ns/avro"
)

var interactiveUploadedEvent = `{
  "type": "record",
  "name": "interactive-uploaded",
  "fields": [
    {"name": "id", "type": "string", "default": ""},
    {"name": "path", "type": "string", "default": ""}
  ]
}`

// InteractiveUploadedEvent is the Avro schema
var InteractiveUploadedEvent = &avro.Schema{
	Definition: interactiveUploadedEvent,
}
