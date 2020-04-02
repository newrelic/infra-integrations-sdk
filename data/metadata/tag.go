package metadata

//TagMap list of tags that will can be added as dimensions.
type TagMap map[string]interface{}

// Tag is key-value pair. Not used for storage, only for convenience
type Tag struct {
	Key   string
	Value interface{}
}
