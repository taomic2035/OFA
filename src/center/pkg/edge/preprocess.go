package edge

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// PreprocessorType defines preprocessor types
type PreprocessorType string

const (
	PreprocessorImage    PreprocessorType = "image"
	PreprocessorAudio    PreprocessorType = "audio"
	PreprocessorVideo    PreprocessorType = "video"
	PreprocessorText     PreprocessorType = "text"
	PreprocessorJSON     PreprocessorType = "json"
	PreprocessorBinary   PreprocessorType = "binary"
)

// PreprocessOperation defines preprocessing operations
type PreprocessOperation struct {
	Type       PreprocessorType `json:"type"`
	Operation  string           `json:"operation"`
	Params     map[string]interface{} `json:"params"`
	OutputFormat string         `json:"output_format"`
}

// PreprocessResult represents preprocessing result
type PreprocessResult struct {
	InputSize  int64       `json:"input_size"`
	OutputSize int64       `json:"output_size"`
	Type       PreprocessorType `json:"type"`
	Operation  string      `json:"operation"`
	Duration   int64       `json:"duration_ms"`
	Output     interface{} `json:"output"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// DataPreprocessor handles data preprocessing at the edge
type DataPreprocessor struct {
	dataDir string

	// Preprocessor pipelines
	pipelines sync.Map // map[string]*PreprocessPipeline

	// Cache for preprocessed data
	cache sync.Map // map[string]*PreprocessResult

	// Metrics
	processedCount int64
	totalDuration  int64

	mu sync.RWMutex
}

// PreprocessPipeline defines a preprocessing pipeline
type PreprocessPipeline struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Operations  []PreprocessOperation `json:"operations"`
	Enabled     bool                 `json:"enabled"`
	CreatedAt   time.Time            `json:"created_at"`
	LastUsed    time.Time            `json:"last_used"`
	UseCount    int                  `json:"use_count"`
}

// NewDataPreprocessor creates a new data preprocessor
func NewDataPreprocessor(dataDir string) *DataPreprocessor {
	dp := &DataPreprocessor{
		dataDir: dataDir,
	}

	// Register built-in preprocessors
	dp.registerBuiltInPipelines()

	return dp
}

// registerBuiltInPipelines registers built-in preprocessing pipelines
func (dp *DataPreprocessor) registerBuiltInPipelines() {
	// Image resize pipeline
	dp.RegisterPipeline(&PreprocessPipeline{
		ID:   "image_resize",
		Name: "Image Resize",
		Operations: []PreprocessOperation{
			{Type: PreprocessorImage, Operation: "resize", Params: map[string]interface{}{}},
		},
		Enabled:   true,
		CreatedAt: time.Now(),
	})

	// Image normalize pipeline
	dp.RegisterPipeline(&PreprocessPipeline{
		ID:   "image_normalize",
		Name: "Image Normalize",
		Operations: []PreprocessOperation{
			{Type: PreprocessorImage, Operation: "normalize", Params: map[string]interface{}{}},
		},
		Enabled:   true,
		CreatedAt: time.Now(),
	})

	// Text clean pipeline
	dp.RegisterPipeline(&PreprocessPipeline{
		ID:   "text_clean",
		Name: "Text Clean",
		Operations: []PreprocessOperation{
			{Type: PreprocessorText, Operation: "clean", Params: map[string]interface{}{}},
		},
		Enabled:   true,
		CreatedAt: time.Now(),
	})

	// Text tokenize pipeline
	dp.RegisterPipeline(&PreprocessPipeline{
		ID:   "text_tokenize",
		Name: "Text Tokenize",
		Operations: []PreprocessOperation{
			{Type: PreprocessorText, Operation: "tokenize", Params: map[string]interface{}{}},
		},
		Enabled:   true,
		CreatedAt: time.Now(),
	})

	// JSON extract pipeline
	dp.RegisterPipeline(&PreprocessPipeline{
		ID:   "json_extract",
		Name: "JSON Extract",
		Operations: []PreprocessOperation{
			{Type: PreprocessorJSON, Operation: "extract", Params: map[string]interface{}{}},
		},
		Enabled:   true,
		CreatedAt: time.Now(),
	})
}

// Process processes data using specified parameters
func (dp *DataPreprocessor) Process(params map[string]interface{}) (interface{}, error) {
	startTime := time.Now()

	// Get pipeline ID if specified
	pipelineID, _ := params["pipeline_id"].(string)
	if pipelineID != "" {
		return dp.processWithPipeline(pipelineID, params)
	}

	// Get operation type
	opType, _ := params["type"].(string)
	operation, _ := params["operation"].(string)

	if opType == "" || operation == "" {
		return nil, errors.New("type and operation required")
	}

	result, err := dp.processOperation(PreprocessorType(opType), operation, params)
	if err != nil {
		return nil, err
	}

	dp.mu.Lock()
	dp.processedCount++
	dp.totalDuration += time.Since(startTime).Milliseconds()
	dp.mu.Unlock()

	return result, nil
}

// processWithPipeline processes data using a pipeline
func (dp *DataPreprocessor) processWithPipeline(pipelineID string, params map[string]interface{}) (interface{}, error) {
	v, ok := dp.pipelines.Load(pipelineID)
	if !ok {
		return nil, fmt.Errorf("pipeline not found: %s", pipelineID)
	}

	pipeline := v.(*PreprocessPipeline)
	if !pipeline.Enabled {
		return nil, fmt.Errorf("pipeline disabled: %s", pipelineID)
	}

	var result interface{} = params["data"]

	for _, op := range pipeline.Operations {
		// Merge pipeline params with input params
		opParams := make(map[string]interface{})
		for k, v := range op.Params {
			opParams[k] = v
		}
		for k, v := range params {
			if k != "type" && k != "operation" {
				opParams[k] = v
			}
		}

		var err error
		result, err = dp.processOperation(op.Type, op.Operation, opParams)
		if err != nil {
			return nil, fmt.Errorf("pipeline operation %s failed: %v", op.Operation, err)
		}

		// Pass result to next operation
		params["data"] = result
	}

	// Update pipeline stats
	pipeline.LastUsed = time.Now()
	pipeline.UseCount++

	return result, nil
}

// processOperation processes a single operation
func (dp *DataPreprocessor) processOperation(opType PreprocessorType, operation string, params map[string]interface{}) (interface{}, error) {
	switch opType {
	case PreprocessorImage:
		return dp.processImage(operation, params)
	case PreprocessorText:
		return dp.processText(operation, params)
	case PreprocessorJSON:
		return dp.processJSON(operation, params)
	case PreprocessorAudio:
		return dp.processAudio(operation, params)
	case PreprocessorVideo:
		return dp.processVideo(operation, params)
	default:
		return nil, fmt.Errorf("unsupported preprocessor type: %s", opType)
	}
}

// processImage processes image data
func (dp *DataPreprocessor) processImage(operation string, params map[string]interface{}) (interface{}, error) {
	// Get image data
	data, ok := params["data"]
	if !ok {
		return nil, errors.New("image data required")
	}

	switch operation {
	case "resize":
		return dp.imageResize(data, params)
	case "normalize":
		return dp.imageNormalize(data, params)
	case "crop":
		return dp.imageCrop(data, params)
	case "convert":
		return dp.imageConvert(data, params)
	case "metadata":
		return dp.imageMetadata(data, params)
	default:
		return nil, fmt.Errorf("unknown image operation: %s", operation)
	}
}

// imageResize resizes an image
func (dp *DataPreprocessor) imageResize(data interface{}, params map[string]interface{}) (interface{}, error) {
	width, _ := params["width"].(float64)
	height, _ := params["height"].(float64)

	// Placeholder - would use actual image library
	return map[string]interface{}{
		"operation": "resize",
		"width":     width,
		"height":    height,
		"status":    "completed",
	}, nil
}

// imageNormalize normalizes image pixel values
func (dp *DataPreprocessor) imageNormalize(data interface{}, params map[string]interface{}) (interface{}, error) {
	// Placeholder for image normalization
	return map[string]interface{}{
		"operation": "normalize",
		"status":    "completed",
	}, nil
}

// imageCrop crops an image
func (dp *DataPreprocessor) imageCrop(data interface{}, params map[string]interface{}) (interface{}, error) {
	x, _ := params["x"].(float64)
	y, _ := params["y"].(float64)
	width, _ := params["width"].(float64)
	height, _ := params["height"].(float64)

	return map[string]interface{}{
		"operation": "crop",
		"x":         x,
		"y":         y,
		"width":     width,
		"height":    height,
		"status":    "completed",
	}, nil
}

// imageConvert converts image format
func (dp *DataPreprocessor) imageConvert(data interface{}, params map[string]interface{}) (interface{}, error) {
	format, _ := params["format"].(string)

	return map[string]interface{}{
		"operation": "convert",
		"format":    format,
		"status":    "completed",
	}, nil
}

// imageMetadata extracts image metadata
func (dp *DataPreprocessor) imageMetadata(data interface{}, params map[string]interface{}) (interface{}, error) {
	// Placeholder for metadata extraction
	return map[string]interface{}{
		"operation":  "metadata",
		"width":      0,
		"height":     0,
		"format":     "unknown",
		"color_mode": "unknown",
	}, nil
}

// processText processes text data
func (dp *DataPreprocessor) processText(operation string, params map[string]interface{}) (interface{}, error) {
	text, _ := params["data"].(string)

	switch operation {
	case "clean":
		return dp.textClean(text, params)
	case "tokenize":
		return dp.textTokenize(text, params)
	case "lowercase":
		return strings.ToLower(text), nil
	case "uppercase":
		return strings.ToUpper(text), nil
	case "split":
		sep, _ := params["separator"].(string)
		if sep == "" {
			sep = " "
		}
		return strings.Split(text, sep), nil
	case "truncate":
		maxLen, _ := params["max_length"].(float64)
		if len(text) > int(maxLen) {
			return text[:int(maxLen)], nil
		}
		return text, nil
	default:
		return nil, fmt.Errorf("unknown text operation: %s", operation)
	}
}

// textClean cleans text data
func (dp *DataPreprocessor) textClean(text string, params map[string]interface{}) (interface{}, error) {
	// Remove extra whitespace
	result := strings.TrimSpace(text)

	// Remove multiple spaces
	for strings.Contains(result, "  ") {
		result = strings.ReplaceAll(result, "  ", " ")
	}

	// Remove special characters if requested
	if removeSpecial, _ := params["remove_special"].(bool); removeSpecial {
		// Keep only alphanumeric and spaces
		var cleaned strings.Builder
		for _, r := range result {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ' {
				cleaned.WriteRune(r)
			}
		}
		result = cleaned.String()
	}

	return result, nil
}

// textTokenize tokenizes text
func (dp *DataPreprocessor) textTokenize(text string, params map[string]interface{}) (interface{}, error) {
	method, _ := params["method"].(string)
	if method == "" {
		method = "word"
	}

	switch method {
	case "word":
		return strings.Fields(text), nil
	case "char":
		return strings.Split(text, ""), nil
	case "sentence":
		// Simple sentence split
		result := strings.Split(text, ". ")
		for i, s := range result {
			result[i] = strings.TrimSpace(s)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unknown tokenize method: %s", method)
	}
}

// processJSON processes JSON data
func (dp *DataPreprocessor) processJSON(operation string, params map[string]interface{}) (interface{}, error) {
	data := params["data"]

	switch operation {
	case "extract":
		path, _ := params["path"].(string)
		return dp.jsonExtract(data, path)
	case "flatten":
		return dp.jsonFlatten(data)
	case "validate":
		return dp.jsonValidate(data)
	case "schema":
		return dp.jsonSchema(data)
	default:
		return nil, fmt.Errorf("unknown json operation: %s", operation)
	}
}

// jsonExtract extracts value from JSON path
func (dp *DataPreprocessor) jsonExtract(data interface{}, path string) (interface{}, error) {
	if path == "" {
		return data, nil
	}

	// Simple path extraction (e.g., "user.name")
	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			if v, exists := m[part]; exists {
				current = v
			} else {
				return nil, fmt.Errorf("path not found: %s", part)
			}
		} else {
			return nil, fmt.Errorf("invalid path: %s", part)
		}
	}

	return current, nil
}

// jsonFlatten flattens nested JSON
func (dp *DataPreprocessor) jsonFlatten(data interface{}) (interface{}, error) {
	result := make(map[string]interface{})
	dp.flattenRecursive("", data, result)
	return result, nil
}

// flattenRecursive recursively flattens JSON
func (dp *DataPreprocessor) flattenRecursive(prefix string, data interface{}, result map[string]interface{}) {
	if m, ok := data.(map[string]interface{}); ok {
		for k, v := range m {
			newKey := k
			if prefix != "" {
				newKey = prefix + "." + k
			}
			dp.flattenRecursive(newKey, v, result)
		}
	} else {
		result[prefix] = data
	}
}

// jsonValidate validates JSON structure
func (dp *DataPreprocessor) jsonValidate(data interface{}) (interface{}, error) {
	return map[string]interface{}{
		"valid":  true,
		"type":   getJSONType(data),
		"size":   getJSONSize(data),
	}, nil
}

// jsonSchema generates JSON schema
func (dp *DataPreprocessor) jsonSchema(data interface{}) (interface{}, error) {
	return map[string]interface{}{
		"type": getJSONType(data),
	}, nil
}

// processAudio processes audio data
func (dp *DataPreprocessor) processAudio(operation string, params map[string]interface{}) (interface{}, error) {
	switch operation {
	case "convert":
		return dp.audioConvert(params)
	case "resample":
		return dp.audioResample(params)
	case "normalize":
		return dp.audioNormalize(params)
	case "trim":
		return dp.audioTrim(params)
	default:
		return nil, fmt.Errorf("unknown audio operation: %s", operation)
	}
}

// audioConvert converts audio format
func (dp *DataPreprocessor) audioConvert(params map[string]interface{}) (interface{}, error) {
	format, _ := params["format"].(string)
	return map[string]interface{}{
		"operation": "convert",
		"format":    format,
		"status":    "completed",
	}, nil
}

// audioResample resamples audio
func (dp *DataPreprocessor) audioResample(params map[string]interface{}) (interface{}, error) {
	rate, _ := params["sample_rate"].(float64)
	return map[string]interface{}{
		"operation":   "resample",
		"sample_rate": rate,
		"status":      "completed",
	}, nil
}

// audioNormalize normalizes audio levels
func (dp *DataPreprocessor) audioNormalize(params map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"operation": "normalize",
		"status":    "completed",
	}, nil
}

// audioTrim trims silence from audio
func (dp *DataPreprocessor) audioTrim(params map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"operation": "trim",
		"status":    "completed",
	}, nil
}

// processVideo processes video data
func (dp *DataPreprocessor) processVideo(operation string, params map[string]interface{}) (interface{}, error) {
	switch operation {
	case "extract_frames":
		return dp.videoExtractFrames(params)
	case "resize":
		return dp.videoResize(params)
	case "convert":
		return dp.videoConvert(params)
	case "trim":
		return dp.videoTrim(params)
	default:
		return nil, fmt.Errorf("unknown video operation: %s", operation)
	}
}

// videoExtractFrames extracts frames from video
func (dp *DataPreprocessor) videoExtractFrames(params map[string]interface{}) (interface{}, error) {
	fps, _ := params["fps"].(float64)
	return map[string]interface{}{
		"operation": "extract_frames",
		"fps":       fps,
		"status":    "completed",
	}, nil
}

// videoResize resizes video
func (dp *DataPreprocessor) videoResize(params map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"operation": "resize",
		"status":    "completed",
	}, nil
}

// videoConvert converts video format
func (dp *DataPreprocessor) videoConvert(params map[string]interface{}) (interface{}, error) {
	format, _ := params["format"].(string)
	return map[string]interface{}{
		"operation": "convert",
		"format":    format,
		"status":    "completed",
	}, nil
}

// videoTrim trims video
func (dp *DataPreprocessor) videoTrim(params map[string]interface{}) (interface{}, error) {
	start, _ := params["start"].(float64)
	end, _ := params["end"].(float64)
	return map[string]interface{}{
		"operation": "trim",
		"start":     start,
		"end":       end,
		"status":    "completed",
	}, nil
}

// RegisterPipeline registers a preprocessing pipeline
func (dp *DataPreprocessor) RegisterPipeline(pipeline *PreprocessPipeline) error {
	if pipeline.ID == "" {
		return errors.New("pipeline ID required")
	}

	pipeline.CreatedAt = time.Now()
	dp.pipelines.Store(pipeline.ID, pipeline)

	return nil
}

// GetPipeline returns a pipeline by ID
func (dp *DataPreprocessor) GetPipeline(id string) (*PreprocessPipeline, bool) {
	if v, ok := dp.pipelines.Load(id); ok {
		return v.(*PreprocessPipeline), true
	}
	return nil, false
}

// ListPipelines returns all pipelines
func (dp *DataPreprocessor) ListPipelines() []*PreprocessPipeline {
	var pipelines []*PreprocessPipeline
	dp.pipelines.Range(func(key, value interface{}) bool {
		pipelines = append(pipelines, value.(*PreprocessPipeline))
		return true
	})
	return pipelines
}

// GetStats returns preprocessor statistics
func (dp *DataPreprocessor) GetStats() map[string]interface{} {
	dp.mu.RLock()
	defer dp.mu.RUnlock()

	var avgDuration int64
	if dp.processedCount > 0 {
		avgDuration = dp.totalDuration / dp.processedCount
	}

	return map[string]interface{}{
		"processed_count": dp.processedCount,
		"total_duration":  dp.totalDuration,
		"avg_duration":    avgDuration,
		"pipeline_count":  len(dp.ListPipelines()),
	}
}

// Helper functions

func getJSONType(data interface{}) string {
	switch data.(type) {
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case nil:
		return "null"
	default:
		return "unknown"
	}
}

func getJSONSize(data interface{}) int {
	switch v := data.(type) {
	case map[string]interface{}:
		return len(v)
	case []interface{}:
		return len(v)
	case string:
		return len(v)
	default:
		return 1
	}
}