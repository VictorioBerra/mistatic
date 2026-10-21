package api

import (
	"archive/zip"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"mistatic/helpers"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

func RegisterDeployRoute(se *core.ServeEvent) {
	se.Router.POST("/api/mistatic/deploy/{site_id}", func(e *core.RequestEvent) error {
		// Basic auth & ownership check
		siteID := e.Request.PathValue("site_id")
		siteRecord, err := findOwnedSite(e, siteID)
		if err != nil {
			return err
		}

		// Parse multipart form (up to 128MB in memory, rest on disk)
		if err := e.Request.ParseMultipartForm(128 << 20); err != nil && err != http.ErrNotMultipart {
			return e.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to parse multipart form: " + err.Error()})
		}

		if e.Request.MultipartForm == nil {
			return e.JSON(http.StatusBadRequest, map[string]string{"error": "Missing multipart form data"})
		}

		// Collect files from "files" or "file" form keys
		var allFiles []*multipart.FileHeader
		if fList, ok := e.Request.MultipartForm.File["files"]; ok {
			allFiles = append(allFiles, fList...)
		}
		if fList, ok := e.Request.MultipartForm.File["file"]; ok {
			allFiles = append(allFiles, fList...)
		}

		if len(allFiles) == 0 {
			return e.JSON(http.StatusBadRequest, map[string]string{"error": "Missing file(s) for deployment"})
		}

		// Optional relative paths corresponding to each file
		var paths []string
		if p, ok := e.Request.MultipartForm.Value["paths"]; ok {
			paths = p
		} else if p, ok := e.Request.MultipartForm.Value["paths[]"]; ok {
			paths = p
		}

		// Create a deployment record in "pending" status
		collection, err := e.App.FindCollectionByNameOrId("deployments")
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Deployments collection not found"})
		}

		deployment := core.NewRecord(collection)
		deployment.Set("site", siteID)
		deployment.Set("status", "pending")
		if err := e.App.Save(deployment); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to initialize deployment record: " + err.Error()})
		}

		// Target site deployment directory
		cwd, _ := os.Getwd()
		siteDir := filepath.Join(cwd, "..", "sites", siteID, deployment.Id)
		if err := os.MkdirAll(siteDir, os.ModePerm); err != nil {
			deployment.Set("status", "failed")
			_ = e.App.Save(deployment)
			return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create deployment directory"})
		}

		// Check if the upload is a single zip file
		isSingleZip := len(allFiles) == 1 && strings.HasSuffix(strings.ToLower(allFiles[0].Filename), ".zip")

		var totalSize int64
		tempZipPath := filepath.Join(os.TempDir(), fmt.Sprintf("deploy_artifact_%s_%s.zip", siteID, deployment.Id))
		defer os.Remove(tempZipPath)

		if isSingleZip {
			// Save uploaded zip to temp file
			src, err := allFiles[0].Open()
			if err != nil {
				deployment.Set("status", "failed")
				_ = e.App.Save(deployment)
				_ = os.RemoveAll(siteDir)
				return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to read uploaded zip"})
			}

			dst, err := os.Create(tempZipPath)
			if err != nil {
				src.Close()
				deployment.Set("status", "failed")
				_ = e.App.Save(deployment)
				_ = os.RemoveAll(siteDir)
				return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to stage zip archive"})
			}
			_, err = io.Copy(dst, src)
			src.Close()
			dst.Close()
			if err != nil {
				deployment.Set("status", "failed")
				_ = e.App.Save(deployment)
				_ = os.RemoveAll(siteDir)
				return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to write staged zip archive"})
			}

			// Unzip into siteDir
			if err := unzip(tempZipPath, siteDir); err != nil {
				deployment.Set("status", "failed")
				_ = e.App.Save(deployment)
				_ = os.RemoveAll(siteDir)
				return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to extract zip file: " + err.Error()})
			}

			// Calculate extracted total size
			_ = filepath.Walk(siteDir, func(_ string, info os.FileInfo, _ error) error {
				if info != nil && !info.IsDir() {
					totalSize += info.Size()
				}
				return nil
			})
		} else {
			// Save each file into siteDir, preserving subdirectories
			cleanSiteDir := filepath.Clean(siteDir) + string(os.PathSeparator)

			for i, fh := range allFiles {
				relPath := ""
				if i < len(paths) && strings.TrimSpace(paths[i]) != "" {
					relPath = strings.TrimSpace(paths[i])
				} else {
					relPath = fh.Filename
				}

				relPath = filepath.ToSlash(filepath.Clean(relPath))
				relPath = strings.TrimPrefix(relPath, "/")
				if strings.HasPrefix(relPath, "../") || relPath == ".." || relPath == "." || relPath == "" {
					relPath = filepath.Base(fh.Filename)
				}

				destFile := filepath.Join(siteDir, filepath.FromSlash(relPath))
				cleanDest := filepath.Clean(destFile)
				if !strings.HasPrefix(cleanDest, cleanSiteDir) && cleanDest != filepath.Clean(siteDir) {
					deployment.Set("status", "failed")
					_ = e.App.Save(deployment)
					_ = os.RemoveAll(siteDir)
					return e.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid file path traversal: " + relPath})
				}

				if err := os.MkdirAll(filepath.Dir(destFile), os.ModePerm); err != nil {
					deployment.Set("status", "failed")
					_ = e.App.Save(deployment)
					_ = os.RemoveAll(siteDir)
					return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create directory: " + err.Error()})
				}

				src, err := fh.Open()
				if err != nil {
					deployment.Set("status", "failed")
					_ = e.App.Save(deployment)
					_ = os.RemoveAll(siteDir)
					return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to open uploaded file"})
				}

				dst, err := os.Create(destFile)
				if err != nil {
					src.Close()
					deployment.Set("status", "failed")
					_ = e.App.Save(deployment)
					_ = os.RemoveAll(siteDir)
					return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to write file"})
				}

				n, err := io.Copy(dst, src)
				src.Close()
				dst.Close()
				if err != nil {
					deployment.Set("status", "failed")
					_ = e.App.Save(deployment)
					_ = os.RemoveAll(siteDir)
					return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save file content"})
				}

				totalSize += n
			}

			// Compress siteDir into artifact zip
			if err := zipDir(siteDir, tempZipPath); err != nil {
				deployment.Set("status", "failed")
				_ = e.App.Save(deployment)
				_ = os.RemoveAll(siteDir)
				return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate artifact zip: " + err.Error()})
			}
		}

		// Attach artifact zip to deployment record
		pbFile, err := filesystem.NewFileFromPath(tempZipPath)
		if err != nil {
			deployment.Set("status", "failed")
			_ = e.App.Save(deployment)
			_ = os.RemoveAll(siteDir)
			return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to prepare deployment artifact: " + err.Error()})
		}
		if isSingleZip {
			pbFile.OriginalName = filepath.Base(allFiles[0].Filename)
		} else {
			pbFile.OriginalName = "deployment.zip"
		}

		deployment.Set("deployment_artifact", pbFile)
		deployment.Set("file_size", totalSize)
		deployment.Set("status", "success")

		if err := e.App.Save(deployment); err != nil {
			_ = os.RemoveAll(siteDir)
			return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save deployment record: " + err.Error()})
		}

		// Activate deployment on site
		siteRecord.Set("active_deployment", deployment.Id)
		if err := e.App.Save(siteRecord); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set active deployment: " + err.Error()})
		}

		return e.JSON(http.StatusOK, map[string]string{
			"message":       "Deployed successfully",
			"deployment_id": deployment.Id,
		})
	}).Bind(helpers.RequireRole("Admin"))
}

func unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	cleanDest := filepath.Clean(dest) + string(os.PathSeparator)

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		cleanPath := filepath.Clean(fpath)
		if !strings.HasPrefix(cleanPath, cleanDest) && cleanPath != filepath.Clean(dest) {
			continue // Zip Slip vulnerability defense
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return err
			}
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

func zipDir(srcDir string, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	w := zip.NewWriter(zipFile)
	defer w.Close()

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == srcDir {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
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

		writer, err := w.CreateHeader(header)
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
}
