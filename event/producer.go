package event

import "errors"

//go:generate moq -out mock/marshaller.go -pkg mock . Marshaller

type AvroProducer struct {
	out        chan []byte
	marshaller Marshaller
}

// Marshaller marshals events into messages.
type Marshaller interface {
	Marshal(s interface{}) ([]byte, error)
}

// NewAvroProducer returns a new instance of AvroProducer.
func NewAvroProducer(outputChannel chan []byte, marshaller Marshaller) *AvroProducer {
	return &AvroProducer{
		out:        outputChannel,
		marshaller: marshaller,
	}
}

// InteractiveUploaded produces a new InteractiveUploaded event.
func (producer *AvroProducer) InteractiveUploaded(event *InteractiveUploaded) error {
	if event == nil {
		return errors.New("event required but was nil")
	}
	return producer.marshalAndSendEvent(event)
}

//marshalAndSendEvent is a generic function that marshals avro events and sends them to the output channel of the producer
func (producer *AvroProducer) marshalAndSendEvent(event interface{}) error {
	bytes, err := producer.marshaller.Marshal(event)
	if err != nil {
		return err
	}
	// highly unlikely but worth checking
	if producer.out != nil {
		producer.out <- bytes
	}
	return nil
}
