package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileInfo struct {
	Path       string `json:"path"`
	SHA256     string `json:"sha256"`
	FileID     string `json:"file_id,omitempty"`
	ManifestID string `json:"manifest_id,omitempty"`
}

type Manifest struct {
	ManifestID  string     `json:"manifest_id"`
	Files       []FileInfo `json:"files"`
	LoggingInfo LogInfo    `json:"log_info"`
}

type LogInfo struct {
	GeneratedAt   string `json:"generated_at"`
	OpenAIAPIKey  string `json:"openai_api_key"`
	ScanFolder    string `json:"scan_folder"`
	VectorStoreID string `json:"vector_store_id"`
	Cleanup       bool   `json:"cleanup"`
	DryRun        bool   `json:"dry_run"`
	OutputFile    string `json:"output_file,omitempty"`
}

var (
	apiKey        string
	cleanup       bool
	dryRun        bool
	output        string
	vectorStoreID string
	folder        string
)

func init() {
	apiKey = os.Getenv("OPENAI_API_KEY")
	flag.BoolVar(&cleanup, "cleanup", false, "enable cleanup of deleted files in OpenAI")
	flag.BoolVar(&dryRun, "dry-run", false, "disable uploading to OpenAI")
	flag.StringVar(&output, "output", "", "output file for the manifest; if not specified, print to console")
	flag.StringVar(&vectorStoreID, "vector-store-id", "", "ID of the OpenAI Vector Store")
	flag.StringVar(&folder, "folder", "./your-folder", "folder to scan for files")
}

func main() {
	flag.Parse()

	var manifest Manifest

	// Read existing manifest if available
	if _, err := os.Stat(output); err == nil {
		data, err := ioutil.ReadFile(output)
		if err == nil {
			json.Unmarshal(data, &manifest)
		}
	}

	// Generate a new manifest ID if it doesn't exist
	if manifest.ManifestID == "" {
		manifest.ManifestID = generateManifestID(folder)
	}

	// Scan the folder and update the manifest
	updatedManifest := scanFolder(folder, manifest)

	// Log configuration information
	updatedManifest.LoggingInfo = LogInfo{
		GeneratedAt:   time.Now().Format(time.RFC3339),
		OpenAIAPIKey:  hideAPIKey(apiKey),
		ScanFolder:    folder,
		VectorStoreID: vectorStoreID,
		Cleanup:       cleanup,
		DryRun:        dryRun,
		OutputFile:    output,
	}

	// Upload changed files to OpenAI if not in dry-run mode
	if !dryRun {
		for i, fileInfo := range updatedManifest.Files {
			if fileInfo.FileID == "" {
				fileID := uploadFile(fileInfo.Path, updatedManifest.ManifestID)
				updatedManifest.Files[i].FileID = fileID
				fmt.Printf("Uploaded %s, got FileID: %s\n", fileInfo.Path, fileID)

				// Add/Update file in vector store
				createVectorStoreFile(fileID)
			}
		}
	}

	// Perform cleanup if enabled and not in dry-run mode
	if cleanup && !dryRun {
		performCleanup(updatedManifest, manifest)
	}

	// Save or print the updated manifest
	saveOrPrintManifest(updatedManifest, output)
}

func generateManifestID(folder string) string {
	hash := sha256.New()
	filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			hash.Write([]byte(path))
		}
		return nil
	})
	return hex.EncodeToString(hash.Sum(nil))
}

func scanFolder(folder string, manifest Manifest) Manifest {
	manifestMap := make(map[string]FileInfo)
	for _, fileInfo := range manifest.Files {
		manifestMap[fileInfo.Path] = fileInfo
	}

	filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			hash := hashFile(path)
			if fileInfo, exists := manifestMap[path]; !exists || fileInfo.SHA256 != hash {
				manifestMap[path] = FileInfo{Path: path, SHA256: hash, ManifestID: manifest.ManifestID}
			}
		}
		return nil
	})

	var files []FileInfo
	for _, fileInfo := range manifestMap {
		files = append(files, fileInfo)
	}

	return Manifest{ManifestID: manifest.ManifestID, Files: files, LoggingInfo: manifest.LoggingInfo}
}

func hashFile(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		panic(err)
	}

	return hex.EncodeToString(hash.Sum(nil))
}

func uploadFile(filePath string, manifestID string) string {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	stat, _ := file.Stat()
	data := make([]byte, stat.Size())
	file.Read(data)

	uploadURL := "https://api.openai.com/v1/files"
	values := map[string]string{"purpose": "manifest"}
	valuesJSON, _ := json.Marshal(values)

	body := bytes.NewReader(data)
	req, _ := http.NewRequest("POST", uploadURL, body)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Manifest-ID", manifestID)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Error uploading file: %s\n", string(respBody))
		panic(fmt.Sprintf("Non-OK HTTP status: %s", resp.Status))
	}

	respBody, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	return result["id"].(string)
}

func deleteFile(fileID string) {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", "https://api.openai.com/v1/files/"+fileID, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Error deleting file: %s\n", string(respBody))
		panic(fmt.Sprintf("Non-OK HTTP status: %s", resp.Status))
	}
}

func createVectorStoreFile(fileID string) {
	url := fmt.Sprintf("https://api.openai.com/v1/vector_stores/%s/files", vectorStoreID)
	values := map[string]string{"file_id": fileID}
	valuesJSON, _ := json.Marshal(values)

	req, _ := http.NewRequest("POST", url, bytes.NewReader(valuesJSON))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Error creating vector store file: %s\n", string(respBody))
		panic(fmt.Sprintf("Non-OK HTTP status: %s", resp.Status))
	}
}

func removeFromVectorStore(fileID string) {
	url := fmt.Sprintf("https://api.openai.com/v1/vector_stores/%s/files/%s", vectorStoreID, fileID)

	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Error removing vector store file: %s\n", string(respBody))
		panic(fmt.Sprintf("Non-OK HTTP status: %s", resp.Status))
	}
}

func performCleanup(updatedManifest, oldManifest Manifest) {
	fileMap := make(map[string]FileInfo)
	for _, fileInfo := range updatedManifest.Files {
		fileMap[fileInfo.FileID] = fileInfo
	}

	for _, fileInfo := range oldManifest.Files {
		if _, exists := fileMap[fileInfo.FileID]; !exists {
			// File no longer exists, so delete it
			deleteFile(fileInfo.FileID)
			fmt.Printf("Deleted FileID: %s\n", fileInfo.FileID)

			// Remove file from vector store
			removeFromVectorStore(fileInfo.FileID)
		}
	}
}

func saveOrPrintManifest(manifest Manifest, outputPath string) {
	data, _ := json.MarshalIndent(manifest, "", "  ")

	if outputPath == "" {
		fmt.Println(string(data))
	} else {
		ioutil.WriteFile(outputPath, data, 0644)
	}
}

func hideAPIKey(apiKey string) string {
	if len(apiKey) > 6 {
		return apiKey[:3] + strings.Repeat("*", len(apiKey)-6) + apiKey[len(apiKey)-3:]
	}
	return strings.Repeat("*", len(apiKey))
}
