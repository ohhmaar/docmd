package gdrive

import (
	"fmt"
	"strings"
	"time"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

type DocInfo struct {
	ID           string
	URL          string
	Title        string
	ModifiedTime time.Time
	ModifiedBy   string
}

func CreateDoc(title string, htmlContent string, folderID string) (*DocInfo, error) {
	srv, err := GetDriveService()
	if err != nil {
		return nil, err
	}

	file := &drive.File{
		Name:     title,
		MimeType: "application/vnd.google-apps.document",
	}

	if folderID != "" {
		file.Parents = []string{folderID}
	}

	createdFile, err := srv.Files.Create(file).
		Media(strings.NewReader(htmlContent), googleapi.ContentType("text/html")).
		Fields("id, name, webViewLink, modifiedTime").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	modTime, _ := time.Parse(time.RFC3339, createdFile.ModifiedTime)

	return &DocInfo{
		ID:           createdFile.Id,
		URL:          createdFile.WebViewLink,
		Title:        createdFile.Name,
		ModifiedTime: modTime,
	}, nil
}

func UpdateDoc(docID string, htmlContent string) (*DocInfo, error) {
	srv, err := GetDriveService()
	if err != nil {
		return nil, err
	}

	updatedFile, err := srv.Files.Update(docID, nil).
		Media(strings.NewReader(htmlContent), googleapi.ContentType("text/html")).
		Fields("id, name, webViewLink, modifiedTime").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	modTime, _ := time.Parse(time.RFC3339, updatedFile.ModifiedTime)

	return &DocInfo{
		ID:           updatedFile.Id,
		URL:          updatedFile.WebViewLink,
		Title:        updatedFile.Name,
		ModifiedTime: modTime,
	}, nil
}

func GetDocInfo(docID string) (*DocInfo, error) {
	srv, err := GetDriveService()
	if err != nil {
		return nil, err
	}

	file, err := srv.Files.Get(docID).
		Fields("id, name, webViewLink, modifiedTime, lastModifyingUser").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get document info: %w", err)
	}

	modTime, _ := time.Parse(time.RFC3339, file.ModifiedTime)

	info := &DocInfo{
		ID:           file.Id,
		URL:          file.WebViewLink,
		Title:        file.Name,
		ModifiedTime: modTime,
	}

	if file.LastModifyingUser != nil {
		info.ModifiedBy = file.LastModifyingUser.EmailAddress
	}

	return info, nil
}

func DeleteDoc(docID string) error {
	srv, err := GetDriveService()
	if err != nil {
		return err
	}

	if err := srv.Files.Delete(docID).Do(); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

func DocExists(docID string) bool {
	srv, err := GetDriveService()
	if err != nil {
		return false
	}

	_, err = srv.Files.Get(docID).Fields("id").Do()
	return err == nil
}
