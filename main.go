package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"saveIMG/handlers"
	"saveIMG/models"
	"go.mongodb.org/mongo-driver/mongo"
    "net/http"

	"github.com/gin-gonic/gin"
)

func main() {
    router := gin.Default()
    router.Use(gin.Logger())

    REACT_APP_DB_URI, ok := os.LookupEnv("REACT_APP_DB_URI")
    if !ok{
        log.Fatal("DB uri not found")
    }
    
    save_port, ok := os.LookupEnv("REACT_APP_SAVE_IMG_PORT")
    if !ok{
        log.Fatal("saving port not found")
    }

    // Configuration de la base de données
    dbHandler := handlers.NewDBHandler(REACT_APP_DB_URI, "imageDB")

    // Définir les routes
    router.POST("/upload", func(c *gin.Context) {
        // Récupération du fichier image de la requête
        file, header, err := c.Request.FormFile("image")
        if err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        defer file.Close()

        // Définir le chemin où sauvegarder l'image
        var trueName string = header.Filename
        imagePath := fmt.Sprintf("./images/%s", header.Filename)
        if err := os.MkdirAll(filepath.Dir(imagePath), os.ModePerm); err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        // Sauvegarder l'image dans le dossier img
        if err := c.SaveUploadedFile(header, imagePath); err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        userId := c.Request.FormValue("userId")


        // Création de l'objet ImageData avec le chemin de l'image
        imageData := models.ImageData{
            UserId: userId,
            Name: header.Filename,
            Path: trueName, // Utilisez le chemin de l'image au lieu des données binaires
            ContentType: header.Header.Get("Content-Type"),
        }

        // Sauvegarder le chemin de l'image dans MongoDB
        var insertedID string
        insertedID, err = dbHandler.SaveImagePath(imageData)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        c.JSON(200, gin.H{"message": "Image uploaded and path saved successfully","ID":insertedID})
    })

    router.GET("/images/:filename", func(c *gin.Context) {
        filename := c.Param("filename")
        imagePath := fmt.Sprintf("./images/%s", filename)
        c.File(imagePath)
    })

    router.GET("/image/:id", func(c *gin.Context) {
        id := c.Param("id")
        image, err := dbHandler.FindImageByID(id)
        if err != nil {
            if err == mongo.ErrNoDocuments {
                c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            }
            return
        }
    
        c.JSON(http.StatusOK, image)
    })

    router.GET("/images/user/:userId", func(c *gin.Context) {
        userId := c.Param("userId") 
        images, err := dbHandler.FindAllImagesByIdUser(userId)
        if err != nil {
            if err == mongo.ErrNoDocuments {
                c.JSON(http.StatusNotFound, gin.H{"error": "No images found for the given user ID"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            }
            return
        }

        c.JSON(http.StatusOK, images)
    })
    
    router.PATCH("/image/:id", func(c *gin.Context) {
        id := c.Param("id")
        
        var updateData models.ImageData
        if err := c.ShouldBindJSON(&updateData); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        if err := dbHandler.UpdateImage(id, updateData); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "Image updated successfully"})
    })

    router.Run(fmt.Sprintf(":%s",save_port))
}
