package event

type InteractiveUploaded struct {
	FilePath     string   `avro:"path"`
	ID           string   `avro:"id"`
	CollectionID string   `avro:"collection_id"`
	Title        string   `avro:"title"`
	CurrentFiles []string `avro:"current_files"`
}
