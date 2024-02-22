package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"saveIMG/handlers"
	"saveIMG/models"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
)

func PatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed", http.StatusMethodNotAllowed)
		return
	}

    DB_URI, ok := os.LookupEnv("DB_URI")
    if !ok{
        log.Fatal("DB uri not found")
    }

    // Configuration de la base de données
    dbHandler := handlers.NewDBHandler(DB_URI, "imageDB")

	// Extraire le nom du fichier de l'URL
	id := strings.TrimPrefix(r.URL.Path, "/api/image/")
	if id == "" {
		http.Error(w, "Id is required", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	var input models.ImageData
	if err := json.Unmarshal(body, &input); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err := dbHandler.UpdateImage(id, input); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	image, err := dbHandler.FindImageByID(id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Image not found", http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	
	if err := json.NewEncoder(w).Encode(image); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
