package application

type PreparedDocument struct {
	InputPath             string `json:"inputPath"`
	FileName              string `json:"fileName"`
	DefaultOutputPath     string `json:"defaultOutputPath"`
	DefaultOutputFilename string `json:"defaultOutputFilename"`
	Directory             string `json:"directory"`
}

type EncryptRequest struct {
	InputPath                    string
	OutputPath                   string
	UserPassword                 string
	OwnerPassword                string
	AllowExistingOutputOverwrite bool
	AllowInputOverwrite          bool
}

type EncryptResult struct {
	OutputPath     string `json:"outputPath"`
	OutputFilename string `json:"outputFilename"`
	Message        string `json:"message"`
}
