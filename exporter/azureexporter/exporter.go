package azureexporter

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/fileexporter"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

var _ exporter.Exporter = &Exporter{}

type Exporter struct {
	// Container is the name of your Azure Blob Storage Container similar to Buckets in S3.
	Container string

	// Storage Account Name and Key
	AccountName string
	AccountKey  string

	Format                  string
	Path                    string
	Filename                string
	CsvTemplate             string
	ParquetCompressionCodec string

	// ServiceURL is the URL of the storage account e.g. https://<account>.blob.core.windows.net/
	// It can be overridden by the user to use a custom URL.
	// Default: https://<account>.blob.core.windows.net/
	ServiceURL string
}

func (f *Exporter) initializeAzureClient() (*azblob.Client, error) {
	url := fmt.Sprintf("https://%s.blob.core.windows.net/", f.AccountName)
	if f.ServiceURL != "" {
		url = f.ServiceURL
	}
	if f.AccountKey == "" {
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, err
		}
		return azblob.NewClient(url, cred, nil)
	}
	cred, err := azblob.NewSharedKeyCredential(f.AccountName, f.AccountKey)
	if err != nil {
		return nil, err
	}
	return azblob.NewClientWithSharedKeyCredential(url, cred, nil)
}

func (f *Exporter) Export(
	ctx context.Context,
	logger *fflog.FFLogger,
	featureEvents []exporter.ExportableEvent,
) error {
	if f.AccountName == "" {
		return fmt.Errorf("you should specify an AccountName. %v is invalid", f.AccountName)
	}

	client, err := f.initializeAzureClient()
	if err != nil {
		return err
	}

	if f.Container == "" {
		return fmt.Errorf("you should specify a container. %v is invalid", f.Container)
	}

	outputDir, err := os.MkdirTemp("", "go_feature_flag_AzureBlobStorage_export")
	if err != nil {
		return err
	}

	fileExporter := fileexporter.Exporter{
		Format:                  f.Format,
		OutputDir:               outputDir,
		Filename:                f.Filename,
		CsvTemplate:             f.CsvTemplate,
		ParquetCompressionCodec: f.ParquetCompressionCodec,
	}
	err = fileExporter.Export(ctx, logger, featureEvents)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(outputDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		fileName := file.Name()
		of, err := os.Open(path.Clean(outputDir + "/" + fileName))
		if err != nil {
			logger.Error(
				"[Azure Exporter] impossible to open file",
				slog.String("path", outputDir+"/"+fileName),
			)
			continue
		}
		defer func() { _ = of.Close() }()

		// prepend the path
		source := fileName
		if f.Path != "" {
			source = f.Path + "/" + fileName
		}

		_, err = client.UploadFile(ctx, f.Container, source, of, nil)
		if err != nil {
			logger.Error(
				"[Azure Exporter] failed to upload file",
				slog.String("path", outputDir+"/"+fileName),
			)
			return err
		}

		logger.Info(
			"[Azure Exporter] file uploaded.",
			slog.String("location", f.Container+"/"+fileName),
		)
	}
	return nil
}

func (f *Exporter) IsBulk() bool {
	return true
}
