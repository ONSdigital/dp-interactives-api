package event

type InteractiveUploaded struct {
	FilePath string `avro:"path"`
	ID       string `avro:"id"`
}
