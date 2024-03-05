package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func ImagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	AWS_URI, ok := os.LookupEnv("AWS_URI")
	if !ok {
		log.Fatal("OCR engine URI not found")
	}

	// Extraire le nom du fichier de l'URL
	filename := strings.TrimPrefix(r.URL.Path, "/api/images/")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/image/%s", AWS_URI, filename), nil)
	if err != nil {
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	file, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "image/png") // Ajustez selon le type d'image
	w.WriteHeader(http.StatusOK)
	w.Write(file)
}
