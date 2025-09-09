package main

import (
	"context"
	"log"
	"net/http"
)

// Define structs for data from the database.
// This struct has been moved from main.go
type Asset struct {
	ID        int
	Name      string
	AssetType string
	Location  string
}

// assetsHandler handles the asset register page and form submissions.
// This handler has been moved from main.go
func assetsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		assetType := r.FormValue("asset-type")
		location := r.FormValue("location")

		_, err := db.ExecContext(context.Background(), "INSERT INTO assets (name, asset_type, location) VALUES (?, ?, ?)", name, assetType, location)
		if err != nil {
			log.Printf("Error inserting asset: %v\n", err)
			http.Error(w, "Error saving asset", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/assets", http.StatusSeeOther)
		return
	}

	data, err := getPageData()
	if err != nil {
		log.Printf("Error fetching page data: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	renderTemplate(w, r, data)
}
