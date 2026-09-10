package api

import (
	"archive/zip"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

func RegisterFileRoutes(se *core.ServeEvent) {
	// Download zip
	se.Router.GET("/api/mistatic/deploy/{site_id}/{deployment_id}/download", func(e *core.RequestEvent) error {
		if e.Auth == nil {
			return e.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}
		siteID := e.Request.PathValue("site_id")
		deploymentID := e.Request.PathValue("deployment_id")
		
		siteRecord, err := e.App.FindRecordById("sites", siteID)
		if err != nil || siteRecord.GetString("user") != e.Auth.Id {
			return e.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden"})
		}

		cwd, _ := os.Getwd()
		siteDir := filepath.Join(cwd, "..", "sites", siteID, deploymentID)

		if _, err := os.Stat(siteDir); os.IsNotExist(err) {
			return e.JSON(http.StatusNotFound, map[string]string{"error": "Deployment not found"})
		}

		e.Response.Header().Set("Content-Type", "application/zip")
		e.Response.Header().Set("Content-Disposition", "attachment; filename="+deploymentID+".zip")

		zipWriter := zip.NewWriter(e.Response)
		defer zipWriter.Close()

		filepath.Walk(siteDir, func(path string, info os.FileInfo, err error) error {
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

		return nil
	})

	// List files
	se.Router.GET("/api/mistatic/deploy/{site_id}/{deployment_id}/files", func(e *core.RequestEvent) error {
		if e.Auth == nil {
			return e.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}
		siteID := e.Request.PathValue("site_id")
		deploymentID := e.Request.PathValue("deployment_id")
		
		siteRecord, err := e.App.FindRecordById("sites", siteID)
		if err != nil || siteRecord.GetString("user") != e.Auth.Id {
			return e.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden"})
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
		err = filepath.Walk(siteDir, func(path string, info os.FileInfo, err error) error {
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

		return e.JSON(http.StatusOK, files)
	})

	// Upload/Replace file
	se.Router.POST("/api/mistatic/deploy/{site_id}/{deployment_id}/files/upload", func(e *core.RequestEvent) error {
		if e.Auth == nil {
			return e.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}
		siteID := e.Request.PathValue("site_id")
		deploymentID := e.Request.PathValue("deployment_id")
		targetPath := e.Request.FormValue("path")

		if targetPath == "" {
			return e.JSON(http.StatusBadRequest, map[string]string{"error": "Path is required"})
		}
		
		siteRecord, err := e.App.FindRecordById("sites", siteID)
		if err != nil || siteRecord.GetString("user") != e.Auth.Id {
			return e.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden"})
		}

		cwd, _ := os.Getwd()
		destFile := filepath.Join(cwd, "..", "sites", siteID, deploymentID, targetPath)

		baseDir := filepath.Join(cwd, "..", "sites", siteID, deploymentID)
		if !strings.HasPrefix(filepath.Clean(destFile), filepath.Clean(baseDir)) {
			return e.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid path"})
		}

		file, _, err := e.Request.FormFile("file")
		if err != nil {
			return e.JSON(http.StatusBadRequest, map[string]string{"error": "Missing file"})
		}
		defer file.Close()

		os.MkdirAll(filepath.Dir(destFile), os.ModePerm)
		out, err := os.Create(destFile)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create file"})
		}
		defer out.Close()

		io.Copy(out, file)
		return e.JSON(http.StatusOK, map[string]string{"message": "File uploaded"})
	})

	// Delete file
	se.Router.DELETE("/api/mistatic/deploy/{site_id}/{deployment_id}/files/delete", func(e *core.RequestEvent) error {
		if e.Auth == nil {
			return e.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}
		siteID := e.Request.PathValue("site_id")
		deploymentID := e.Request.PathValue("deployment_id")
		targetPath := e.Request.URL.Query().Get("path")

		if targetPath == "" {
			return e.JSON(http.StatusBadRequest, map[string]string{"error": "Path is required"})
		}
		
		siteRecord, err := e.App.FindRecordById("sites", siteID)
		if err != nil || siteRecord.GetString("user") != e.Auth.Id {
			return e.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden"})
		}

		cwd, _ := os.Getwd()
		destFile := filepath.Join(cwd, "..", "sites", siteID, deploymentID, targetPath)

		baseDir := filepath.Join(cwd, "..", "sites", siteID, deploymentID)
		if !strings.HasPrefix(filepath.Clean(destFile), filepath.Clean(baseDir)) {
			return e.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid path"})
		}

		err = os.RemoveAll(destFile)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete file"})
		}

		return e.JSON(http.StatusOK, map[string]string{"message": "File deleted"})
	})
}
