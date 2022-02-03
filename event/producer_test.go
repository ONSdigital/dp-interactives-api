package event_test

import (
	"testing"

	"github.com/ONSdigital/dp-interactives-api/event"
	"github.com/ONSdigital/dp-interactives-api/event/mock"
	"github.com/ONSdigital/dp-interactives-api/schema"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var errMarshal = errors.New("Marshal error")

func TestAvroProducer(t *testing.T) {

	Convey("Given a successful message producer mock", t, func() {

		// channel to capture messages sent.
		outputChannel := make(chan []byte, 1)

		// bytes to send
		avroBytes := []byte("hello world")

		// mock that represents a marshaller
		marshallerMock := &mock.MarshallerMock{
			MarshalFunc: func(s interface{}) ([]byte, error) {
				return avroBytes, nil
			},
		}

		// eventProducer under test
		eventProducer := event.NewAvroProducer(outputChannel, marshallerMock)

		Convey("when InteractiveUploaded is called with a nil event", func() {
			err := eventProducer.InteractiveUploaded(nil)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldEqual, "event required but was nil")
			})

			Convey("and marshaller is never called", func() {
				So(marshallerMock.MarshalCalls(), ShouldHaveLength, 0)
			})
		})

		Convey("When InteractiveUploaded is called on the event producer", func() {
			event := &event.InteractiveUploaded{
				ID:       "myID",
				FilePath: "myPath",
			}
			err := eventProducer.InteractiveUploaded(event)

			Convey("The expected event is available on the output channel", func() {
				So(err, ShouldBeNil)

				messageBytes := <-outputChannel
				close(outputChannel)
				So(messageBytes, ShouldResemble, avroBytes)
			})
		})

	})

	Convey("Given a message producer mock that fails to marshall", t, func() {

		// mock that represents a marshaller
		marshallerMock := &mock.MarshallerMock{
			MarshalFunc: func(s interface{}) ([]byte, error) {
				return nil, errMarshal
			},
		}

		// eventProducer under test, without out channel because nothing is expected to be sent
		eventProducer := event.NewAvroProducer(nil, marshallerMock)

		Convey("When InteractiveUploaded is called on the event producer", func() {
			event := &event.InteractiveUploaded{
				ID:       "myID",
				FilePath: "myPath",
			}
			err := eventProducer.InteractiveUploaded(event)

			Convey("The expected error is returned", func() {
				So(err, ShouldResemble, errMarshal)
			})
		})
	})
}

// Unmarshal converts observation events to []byte.
func unmarshal(bytes []byte) *event.InteractiveUploaded {
	event := &event.InteractiveUploaded{}
	err := schema.InteractiveUploadedEvent.Unmarshal(bytes, event)
	So(err, ShouldBeNil)
	return event
}
