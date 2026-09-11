package api

import (
	"archive/zip"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"mistatic/helpers"

	"github.com/pocketbase/pocketbase/core"
)

func RegisterDeployRoute(se *core.ServeEvent) {
	se.Router.POST("/api/mistatic/deploy/{site_id}", func(e *core.RequestEvent) error {
		// Basic auth check
		siteID := e.Request.PathValue("site_id")
		siteRecord, err := findOwnedSite(e, siteID)
		if err != nil {
			return err
		}

		// Parse multipart form
		file, header, err := e.Request.FormFile("file")
		if err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{"error": "Missing 'file'"})
		}
		defer file.Close()

		// Create a deployment record
		collection, _ := e.App.FindCollectionByNameOrId("deployments")
		deployment := core.NewRecord(collection)
		deployment.Set("site", siteID)
		deployment.Set("status", "pending")
		deployment.Set("file_size", header.Size)
		e.App.Save(deployment)

		// Create the directory
		cwd, _ := os.Getwd()
		siteDir := filepath.Join(cwd, "..", "sites", siteID, deployment.Id)
		os.MkdirAll(siteDir, os.ModePerm)

		// Save the file
		isZip := strings.HasSuffix(strings.ToLower(header.Filename), ".zip")

		var destFile string
		if isZip {
			destFile = filepath.Join(siteDir, "upload.zip")
		} else {
			// Save directly with original filename
			destFile = filepath.Join(siteDir, filepath.Base(header.Filename))
		}

		out, err := os.Create(destFile)
		if err != nil {
			deployment.Set("status", "failed")
			e.App.Save(deployment)
			return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save file"})
		}
		io.Copy(out, file)
		out.Close()

		if isZip {
			// Unzip
			err = unzip(destFile, siteDir)
			os.Remove(destFile) // cleanup zip
			if err != nil {
				deployment.Set("status", "failed")
				e.App.Save(deployment)
				return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to unzip file"})
			}
		}

		// Success
		deployment.Set("status", "success")
		e.App.Save(deployment)

		// Set as active deployment
		siteRecord.Set("active_deployment", deployment.Id)
		e.App.Save(siteRecord)

		return e.JSON(http.StatusOK, map[string]string{"message": "Deployed successfully", "deployment_id": deployment.Id})
	}).Bind(helpers.RequireRole("Admin"))
}

func unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			continue // Zip Slip vulnerability defense
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
