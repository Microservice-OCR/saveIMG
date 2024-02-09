package handlers

import (
	"context"
	"fmt"
	"log"
	"saveIMG/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// DBHandler contient la connexion à la base de données.
type DBHandler struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewDBHandler crée une nouvelle instance de DBHandler.
func NewDBHandler(uri, dbName string) *DBHandler {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Vérifie la connexion
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	db := client.Database(dbName)
	return &DBHandler{client: client, database: db}
}

// SaveImagePath sauvegarde le chemin de l'image dans la base de données.
func (handler *DBHandler) SaveImagePath(image models.ImageData) (string, error) {
	collection := handler.database.Collection("images")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := collection.InsertOne(ctx, bson.M{
		"idUser":       image.UserId, // Remplacez par l'ID de l'utilisateur
		"name":         image.Name,
		"path":         image.Path,
		"contentType":  image.ContentType,
	})
	if err != nil {
		return "", err
	}
	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("Failed to retrieve inserted ID")
	}
	return insertedID.Hex(), nil
}

func (handler *DBHandler) FindImageByID(id string) (*models.ImageData, error) {
	var image models.ImageData
	collection := handler.database.Collection("images")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	formattedId, _ := primitive.ObjectIDFromHex(id)

	err := collection.FindOne(ctx, bson.M{"_id": formattedId}).Decode(&image)
	if err != nil {
		return nil, err
	}

	return &image, nil
}

func (handler *DBHandler) FindAllImagesByIdUser(idUser string) ([]models.ImageData, error) {
	var images []models.ImageData
	collection := handler.database.Collection("images")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Convertir l'ID utilisateur en un type approprié si nécessaire (par exemple, ObjectID pour MongoDB)
	formattedIdUser, _ := primitive.ObjectIDFromHex(idUser)

	// Utiliser bson.M pour filtrer les documents par idUser
	cursor, err := collection.Find(ctx, bson.M{"idUser": formattedIdUser})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var image models.ImageData
		err := cursor.Decode(&image)
		if err != nil {
			return nil, err
		}
		images = append(images, image)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return images, nil
}

func (handler *DBHandler) UpdateImage(id string, updateData models.ImageData) error {
	collection := handler.database.Collection("images")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"name":         updateData.Name,
			"path":         updateData.Path,
			"contentType":  updateData.ContentType,
			"fulltext": 	updateData.Fulltext,
			"recognition": 	updateData.Recognition,
		},
	}
	formattedId, _ := primitive.ObjectIDFromHex(id)

	_, err := collection.UpdateByID(ctx, formattedId, update)
	return err
}
