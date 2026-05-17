package recommender

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"os"
	"sync"

	ort "github.com/yalue/onnxruntime_go"
)

var envOnce sync.Once
var envInitErr error

type Service struct {
	session      *ort.DynamicAdvancedSession
	logger       *slog.Logger
	enabled      bool
	mu           sync.RWMutex
	inputName    string
	outputName   string
	embeddingDim int64
}

func New(modelPath string, logger *slog.Logger, embeddingDim int64) (*Service, error) {
	s := &Service{
		logger:       logger,
		enabled:      false,
		embeddingDim: embeddingDim,
	}

	if modelPath == "" {
		logger.Warn("REC_MODEL_PATH not set, running in fallback mode")
		return s, nil
	}

	soPath := os.Getenv("ONNXRUNTIME_SHARED_LIBRARY_PATH")
	if soPath == "" {
		soPath = "/usr/lib/onnxruntime/libonnxruntime.so"
	}
	ort.SetSharedLibraryPath(soPath)

	envOnce.Do(func() {
		envInitErr = ort.InitializeEnvironment()
	})
	if envInitErr != nil {
		return nil, fmt.Errorf("onnxruntime init failed: %w", envInitErr)
	}

	inputs, outputs, err := ort.GetInputOutputInfo(modelPath)
	if err != nil {
		return nil, fmt.Errorf("get model I/O info: %w", err)
	}
	if len(inputs) == 0 || len(outputs) == 0 {
		return nil, fmt.Errorf("model has no inputs/outputs")
	}
	s.inputName = inputs[0].Name
	s.outputName = outputs[0].Name

	session, err := ort.NewDynamicAdvancedSession(
		modelPath,
		[]string{s.inputName},
		[]string{s.outputName},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create onnx session: %w", err)
	}

	s.session = session

	s.enabled = true
	logger.Info("ONNX model loaded", "input", s.inputName, "output", s.outputName, "dim", embeddingDim)
	return s, nil
}

func (s *Service) RunInference(_ context.Context, userFeatures []float32) ([]float32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.enabled || s.session == nil {
		return nil, nil
	}

	expectedLen := int(s.embeddingDim)
	if len(userFeatures) != expectedLen {
		return nil, fmt.Errorf(
			"input size mismatch: expected %d, got %d",
			expectedLen,
			len(userFeatures),
		)
	}

	// normalize input EXACTLY like training
	var norm float32
	for _, v := range userFeatures {
		norm += v * v
	}

	norm = float32(math.Sqrt(float64(norm)))

	input := make([]float32, len(userFeatures))
	copy(input, userFeatures)

	if norm > 0 {
		for i := range input {
			input[i] /= norm
		}
	}

	inputShape := ort.NewShape(1, s.embeddingDim)

	inputTensor, err := ort.NewTensor(inputShape, input)
	if err != nil {
		return nil, fmt.Errorf("create input tensor: %w", err)
	}
	defer inputTensor.Destroy()

	outputShape := ort.NewShape(1, s.embeddingDim)

	outputTensor, err := ort.NewEmptyTensor[float32](outputShape)
	if err != nil {
		return nil, fmt.Errorf("create output tensor: %w", err)
	}
	defer outputTensor.Destroy()

	err = s.session.Run(
		[]ort.Value{inputTensor},
		[]ort.Value{outputTensor},
	)

	if err != nil {
		return nil, fmt.Errorf("inference failed: %w", err)
	}

	outputData := outputTensor.GetData()

	result := make([]float32, len(outputData))
	copy(result, outputData)

	// normalize output
	var outNorm float32
	for _, v := range result {
		outNorm += v * v
	}

	outNorm = float32(math.Sqrt(float64(outNorm)))

	if outNorm > 0 {
		for i := range result {
			result[i] /= outNorm
		}
	}

	return result, nil
}

func (s *Service) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.session != nil {
		s.session.Destroy()
		s.session = nil
	}

	s.enabled = false
}
