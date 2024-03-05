package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"saveIMG/handlers"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
)

func UserHandler(w http.ResponseWriter, r *http.Request) {
	// TODO : SUPPRIMER POUR PROD
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*") // or specify your domain
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
    
	if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }
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
	userId := strings.TrimPrefix(r.URL.Path, "/api/user/")
	if userId == "" {
		http.Error(w, "UserId is required", http.StatusBadRequest)
		return
	}

	images, err := dbHandler.FindAllImagesByIdUser(userId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Images not found", http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	
	if err := json.NewEncoder(w).Encode(images); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
