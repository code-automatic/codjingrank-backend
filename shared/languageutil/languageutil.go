package languageutil

import "innogen-backend/shared/models"

// DetermineFileName returns the file name for code execution.
// Uses language.default_file_name, or "solution" + extension, or "main.txt".
func DetermineFileName(lang *models.Language) string {
	if lang.DefaultFileName != nil && *lang.DefaultFileName != "" {
		return *lang.DefaultFileName
	}
	if lang.FileExtension != nil && *lang.FileExtension != "" {
		return "solution" + *lang.FileExtension
	}
	return "main.txt"
}
