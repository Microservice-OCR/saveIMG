package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"saveIMG/handlers"
	"saveIMG/models"
)

type UploadResponse struct {
	Message string `json:"message"`
	Id      string `json:"id"`
}

func UploadImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	DB_URI, ok := os.LookupEnv("DB_URI")
	if !ok {
		log.Fatal("DB uri not found")
	}

	AWS_URI, ok := os.LookupEnv("AWS_URI")
	if !ok {
		log.Fatal("AWS uri not found")
	}

	awsURL := fmt.Sprintf("%s/api/upload", AWS_URI)

	var requestBody bytes.Buffer
	multiPartWriter := multipart.NewWriter(&requestBody)

	part, err := multiPartWriter.CreateFormFile("image", header.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = io.Copy(part, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = multiPartWriter.Close()
	if err != nil {
		log.Panic("Error closing multipart writer: %v", err)
		return
	}

	req, err := http.NewRequest("POST", awsURL, &requestBody)
	if err != nil {
		log.Panic("Error creating HTTP request: %v", err)
		return
	}
	req.Header.Set("Content-Type", multiPartWriter.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Panic("Error performing HTTP request: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Panic("API returned non-OK status: %s", resp.Status)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panic("Error during request: ", err)
		return
	}
    responseAWS,_:=parseResponse(body)

	// Connect to MongoDB
	dbHandler := handlers.NewDBHandler(DB_URI, "imageDB")

	userId := r.Form.Get("userId")

	// Création de l'objet ImageData avec le chemin de l'image
	imageData := models.ImageData{
		UserId:      userId,
		Name:        header.Filename,
		Path:        responseAWS.TrueName, // Utilisez le chemin de l'image au lieu des données binaires
		ContentType: header.Header.Get("Content-Type"),
	}

	// Sauvegarder le chemin de l'image dans MongoDB
	var insertedID string
	insertedID, err = dbHandler.SaveImagePath(imageData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := UploadResponse{
		Message: "Image uploaded and path saved successfully",
		Id:      insertedID,
	}

	w.Header().Set("Content-Type", "application/json")
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func parseResponse(response []byte) (models.IAWSResponse, error) {
	var output models.IAWSResponse
	err := json.Unmarshal(response, &output)
	return output, err
}
