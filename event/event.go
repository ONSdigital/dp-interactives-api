package event

type InteractiveUploaded struct {
	FilePath     string   `avro:"path"`
	ID           string   `avro:"id"`
	Title        string   `avro:"title"`
	CurrentFiles []string `avro:"current_files"`
}
