package models

// ImageData représente les métadonnées d'une image, y compris son chemin d'accès.
type ImageData struct {
	ID          string                `bson:"_id,omitempty"`
	UserId      string                `bson:"user_id"`
	Name        string                `bson:"name"`
	Path        string                `bson:"path"`
	ContentType string                `bson:"content_type"`
	Fulltext    string                `bson:"fulltext"`
	Recognition []IComplexRecognition `bson:"recognition"`
}

type IComplexRecognition map[string]IComplexElement

type IComplexElement struct {
	Value  string `bson:"value"`
	Left   int    `bson:"left"`
	Top    int    `bson:"top"`
	Width  int    `bson:"width"`
	Height int    `bson:"height"`
}
