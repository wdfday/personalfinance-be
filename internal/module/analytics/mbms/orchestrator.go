package mbms

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// modelOrchestrator implements ModelOrchestrator interface
type modelOrchestrator struct {
	registry ModelRegistry
	resolver DependencyResolver
	cache    ResultCache
	logger   *zap.Logger
}

// NewModelOrchestrator tạo orchestrator mới
func NewModelOrchestrator(
	registry ModelRegistry,
	resolver DependencyResolver,
	cache ResultCache,
	logger *zap.Logger,
) ModelOrchestrator {
	return &modelOrchestrator{
		registry: registry,
		resolver: resolver,
		cache:    cache,
		logger:   logger,
	}
}

// ExecutePipeline chạy một chuỗi models theo thứ tự phụ thuộc
func (o *modelOrchestrator) ExecutePipeline(
	ctx context.Context,
	modelNames []string,
	inputs map[string]interface{},
) (map[string]*ModelResult, error) {

	// Step 1: Resolve dependencies để lấy execution order
	orderedModels, err := o.resolver.Resolve(modelNames)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	o.logger.Info("Resolved model execution order",
		zap.Strings("requested", modelNames),
		zap.Strings("ordered", orderedModels))

	// Step 2: Execute models theo thứ tự
	results := make(map[string]*ModelResult)
	modelOutputs := make(map[string]interface{}) // Store outputs để pass cho dependent models

	for _, modelName := range orderedModels {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("pipeline execution cancelled: %w", ctx.Err())
		default:
		}

		o.logger.Info("Executing model", zap.String("model", modelName))

		// Get input cho model này
		var modelInput interface{}
		if input, exists := inputs[modelName]; exists {
			modelInput = input
		} else {
			// Nếu không có explicit input, tạo input từ outputs của dependencies
			modelInput = o.prepareInputFromDependencies(modelName, modelOutputs)
		}

		// Execute model
		result, err := o.ExecuteSingle(ctx, modelName, modelInput)
		if err != nil {
			return nil, fmt.Errorf("failed to execute model '%s': %w", modelName, err)
		}

		// Store result và output
		results[modelName] = result
		modelOutputs[modelName] = result.Output

		o.logger.Info("Model execution completed",
			zap.String("model", modelName),
			zap.Duration("duration", result.Metadata.Duration),
			zap.String("status", result.Metadata.Status))
	}

	return results, nil
}

// ExecuteSingle chạy một model đơn lẻ
func (o *modelOrchestrator) ExecuteSingle(
	ctx context.Context,
	modelName string,
	input interface{},
) (*ModelResult, error) {

	// Get model từ registry
	model, err := o.registry.Get(modelName)
	if err != nil {
		return nil, err
	}

	// Generate cache key
	cacheKey, err := o.generateCacheKey(modelName, input)
	if err != nil {
		o.logger.Warn("Failed to generate cache key",
			zap.String("model", modelName),
			zap.Error(err))
	} else {
		// Try get từ cache
		if o.cache != nil {
			cachedResult, err := o.cache.Get(ctx, cacheKey)
			if err == nil && cachedResult != nil {
				o.logger.Info("Cache hit",
					zap.String("model", modelName),
					zap.String("cache_key", cacheKey))
				return cachedResult, nil
			}
		}
	}

	// Cache miss - execute model
	executionID := uuid.New().String()
	startTime := time.Now()

	metadata := ModelMetadata{
		ModelName:   modelName,
		ExecutionID: executionID,
		StartTime:   startTime,
	}

	// Validate input
	if err := model.Validate(ctx, input); err != nil {
		metadata.EndTime = time.Now()
		metadata.Duration = time.Since(startTime)
		metadata.Status = "failed"
		metadata.ErrorMessage = fmt.Sprintf("validation failed: %v", err)

		return &ModelResult{
			Output:   nil,
			Metadata: metadata,
		}, fmt.Errorf("validation failed: %w", err)
	}

	// Snapshot input for auditing
	if inputJSON, err := json.Marshal(input); err == nil {
		metadata.InputSnapshot = inputJSON
	}

	// Execute model
	output, err := model.Execute(ctx, input)

	// Update metadata
	metadata.EndTime = time.Now()
	metadata.Duration = time.Since(startTime)

	if err != nil {
		metadata.Status = "failed"
		metadata.ErrorMessage = err.Error()

		result := &ModelResult{
			Output:   nil,
			Metadata: metadata,
		}

		return result, err
	}

	// Success
	metadata.Status = "success"

	// Snapshot output for auditing
	if outputJSON, err := json.Marshal(output); err == nil {
		metadata.OutputSnapshot = outputJSON
	}

	result := &ModelResult{
		Output:   output,
		Metadata: metadata,
	}

	// Cache result
	if o.cache != nil && cacheKey != "" {
		// Cache for 1 hour by default
		if err := o.cache.Set(ctx, cacheKey, result, time.Hour); err != nil {
			o.logger.Warn("Failed to cache result",
				zap.String("model", modelName),
				zap.Error(err))
		} else {
			o.logger.Info("Result cached",
				zap.String("model", modelName),
				zap.String("cache_key", cacheKey))
		}
	}

	// Update registry metadata stats
	if err := o.updateModelStats(modelName, metadata.Duration); err != nil {
		o.logger.Warn("Failed to update model stats", zap.Error(err))
	}

	return result, nil
}

// ResolveDependencies delegates to resolver
func (o *modelOrchestrator) ResolveDependencies(modelNames []string) ([]string, error) {
	return o.resolver.Resolve(modelNames)
}

// prepareInputFromDependencies tạo input từ outputs của dependencies
// Helper method để tự động wire inputs
func (o *modelOrchestrator) prepareInputFromDependencies(
	modelName string,
	dependencyOutputs map[string]interface{},
) interface{} {

	// Get model để biết dependencies
	model, err := o.registry.Get(modelName)
	if err != nil {
		return nil
	}

	deps := model.Dependencies()
	if len(deps) == 0 {
		return nil
	}

	// Tạo composite input từ dependency outputs
	compositeInput := make(map[string]interface{})
	for _, dep := range deps {
		if output, exists := dependencyOutputs[dep]; exists {
			compositeInput[dep] = output
		}
	}

	return compositeInput
}

// generateCacheKey generates unique cache key cho model execution
func (o *modelOrchestrator) generateCacheKey(modelName string, input interface{}) (string, error) {
	// Serialize input to JSON
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	// Create cache key: model_name:input_hash
	// Simple approach - in production có thể dùng hash function
	key := fmt.Sprintf("model:%s:%s", modelName, string(inputJSON))

	// Truncate nếu quá dài (Redis key limit)
	if len(key) > 1024 {
		key = key[:1024]
	}

	return key, nil
}

// updateModelStats updates execution statistics trong registry metadata
func (o *modelOrchestrator) updateModelStats(modelName string, duration time.Duration) error {
	meta, err := o.registry.GetMetadata(modelName)
	if err != nil {
		return err
	}

	// Calculate new average exec time (exponential moving average)
	alpha := 0.2 // Weight for new observation
	newAvg := alpha*float64(duration.Milliseconds()) + (1-alpha)*meta.AverageExecTime

	// This would require adding UpdateMetadata method to public interface
	// For now, we log it
	o.logger.Debug("Model stats",
		zap.String("model", modelName),
		zap.Duration("duration", duration),
		zap.Float64("avg_exec_time", newAvg))

	return nil
}
