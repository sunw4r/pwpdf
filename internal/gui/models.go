package gui

import "github.com/sunw4r/pwpdf/internal/application"

const (
	EventPreparedDocument = "app:prepared-document"
	EventErrorMessage     = "app:error-message"
)

type ApplicationState struct {
	Version          string                        `json:"version"`
	PendingSelection *application.PreparedDocument `json:"pendingSelection,omitempty"`
	PendingError     string                        `json:"pendingError,omitempty"`
}

type EncryptionForm struct {
	InputPath            string `json:"inputPath"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"passwordConfirmation"`
	OwnerPassword        string `json:"ownerPassword,omitempty"`
}

type OperationResult struct {
	Cancelled      bool   `json:"cancelled"`
	OutputPath     string `json:"outputPath,omitempty"`
	OutputFilename string `json:"outputFilename,omitempty"`
	Message        string `json:"message,omitempty"`
}
