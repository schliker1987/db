package main

import (
	"image"
	"os"
	"strings"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/gabriel-vasile/mimetype"
)

// ExtractedColors reprensents the object in a Work's metadata.colors
type ExtractedColors struct {
	Primary   string
	Secondary string
	Tertiary  string
}

// StepExtractColors executes the step "extract colors" and returns a metadata object with the `colors` entry modified accordingly.
func StepExtractColors(metadata map[string]interface{}, project ProjectTreeElement, databaseDirectory string, config Configuration) map[string]interface{} {
	// Do not overwrite manually-set `colors` metadata entry
	if _, ok := metadata["colors"]; !ok {
		// Get only image filepaths
		imageFilepaths := FilterSlice(project.MediaAbsoluteFilepaths(databaseDirectory), func(item string) bool {
			contentType, err := mimetype.DetectFile(item)
			return err == nil && strings.HasPrefix(contentType.String(), "image/")
		})
		// Extract colors from them
		extractedColors, err := extractColorsFromFiles(imageFilepaths, config)
		if err == nil {
			metadata["colors"] = extractedColors
		}
	}
	return metadata
}

func extractColorsFromFiles(files []string, config Configuration) (ExtractedColors, error) {
	if len(files) == 0 {
		return ExtractedColors{}, nil
	}
	if len(files) == 1 {
		return extractColors(files[0])
	}
	for _, filename := range files {
		if StringInSlice(config.ExtractColors.DefaultFileName, filename) {
			return extractColors(filename)
		}
	}
	return extractColors(files[0])
}

func extractColors(filename string) (ExtractedColors, error) {
	file, err := os.Open(filename)
	if err != nil {
		return ExtractedColors{}, err
	}
	img, _, err := image.Decode(file)
	if err != nil {
		return ExtractedColors{}, err
	}
	return kmeans(img)
}

// kmeans extracts colors from img
func kmeans(img image.Image) (ExtractedColors, error) {
	centroids, err := prominentcolor.Kmeans(img)
	if err != nil {
		return ExtractedColors{}, err
	}
	colors := make([]string, 3, 3)
	for _, centroid := range centroids {
		colors = append(colors, centroid.AsString())
	}
	return ExtractedColors{
		Primary:   colors[0],
		Secondary: colors[1],
		Tertiary:  colors[2],
	}, nil
}
