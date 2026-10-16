package api

import (
	"archive/zip"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"mistatic/helpers"

	"github.com/pocketbase/pocketbase/core"
)

func RegisterFileRoutes(se *core.ServeEvent) {
	// Download zip
	se.Router.GET("/api/mistatic/deploy/{site_id}/{deployment_id}/download", func(e *core.RequestEvent) error {
		siteID := e.Request.PathValue("site_id")
		deploymentID := e.Request.PathValue("deployment_id")

		if _, err := findOwnedSite(e, siteID); err != nil {
			return err
		}

		// Try streaming from PB storage if deployment_artifact exists
		deployment, err := e.App.FindRecordById("deployments", deploymentID)
		if err == nil && deployment.GetString("site") == siteID {
			artifact := deployment.GetString("deployment_artifact")
			if artifact != "" {
				fsys, err := e.App.NewFilesystem()
				if err == nil {
					defer fsys.Close()
					fileKey := deployment.BaseFilesPath() + "/" + artifact
					r, err := fsys.GetReader(fileKey)
					if err == nil {
						defer r.Close()
						e.Response.Header().Set("Content-Type", "application/zip")
						e.Response.Header().Set("Content-Disposition", "attachment; filename="+deploymentID+".zip")
						_, copyErr := io.Copy(e.Response, r)
						return copyErr
					}
				}
			}
		}

		// Fallback: build zip on the fly from local disk (for legacy deployments)
		cwd, _ := os.Getwd()
		siteDir := filepath.Join(cwd, "..", "sites", siteID, deploymentID)

		if _, err := os.Stat(siteDir); os.IsNotExist(err) {
			return e.JSON(http.StatusNotFound, map[string]string{"error": "Deployment not found"})
		}

		e.Response.Header().Set("Content-Type", "application/zip")
		e.Response.Header().Set("Content-Disposition", "attachment; filename="+deploymentID+".zip")

		zipWriter := zip.NewWriter(e.Response)
		defer zipWriter.Close()

		return filepath.Walk(siteDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if path == siteDir {
				return nil
			}

			relPath, err := filepath.Rel(siteDir, path)
			if err != nil {
				return err
			}

			if info.IsDir() {
				relPath += "/"
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			header.Name = filepath.ToSlash(relPath)
			header.Method = zip.Deflate

			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			return err
		})
	}).Bind(helpers.RequireRole("Admin"))

	// List files (Read-only manifest inspection)
	se.Router.GET("/api/mistatic/deploy/{site_id}/{deployment_id}/files", func(e *core.RequestEvent) error {
		siteID := e.Request.PathValue("site_id")
		deploymentID := e.Request.PathValue("deployment_id")

		if _, err := findOwnedSite(e, siteID); err != nil {
			return err
		}

		cwd, _ := os.Getwd()
		siteDir := filepath.Join(cwd, "..", "sites", siteID, deploymentID)

		type FileNode struct {
			Name  string `json:"name"`
			Path  string `json:"path"`
			IsDir bool   `json:"is_dir"`
			Size  int64  `json:"size"`
		}

		var files []FileNode
		err := filepath.Walk(siteDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || path == siteDir {
				return nil
			}
			relPath, _ := filepath.Rel(siteDir, path)
			files = append(files, FileNode{
				Name:  info.Name(),
				Path:  filepath.ToSlash(relPath),
				IsDir: info.IsDir(),
				Size:  info.Size(),
			})
			return nil
		})
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to list files"})
		}

		return e.JSON(http.StatusOK, files)
	}).Bind(helpers.RequireRole("Admin"))
}
