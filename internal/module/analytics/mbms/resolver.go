package mbms

import (
	"fmt"
)

// dependencyResolver implements DependencyResolver interface
type dependencyResolver struct {
	registry ModelRegistry
}

// NewDependencyResolver tạo resolver mới
func NewDependencyResolver(registry ModelRegistry) DependencyResolver {
	return &dependencyResolver{
		registry: registry,
	}
}

// Resolve returns ordered list of models to execute
// Input: target models (models user muốn chạy)
// Output: ordered list bao gồm cả target models và dependencies
func (r *dependencyResolver) Resolve(targetModels []string) ([]string, error) {
	// Step 1: Build dependency graph
	// graph[node] = list of nodes that depend on it
	// Ví dụ: graph["goal_prioritization"] = ["tradeoff", "budget"]
	//        (tradeoff và budget depend on goal_prioritization)
	graph := make(map[string][]string)

	// inDegree[node] = số dependencies của node đó
	inDegree := make(map[string]int)

	// allNodes tracks tất cả nodes trong graph
	allNodes := make(map[string]bool)

	// Populate graph bắt đầu từ target models
	err := r.buildGraph(targetModels, graph, inDegree, allNodes)
	if err != nil {
		return nil, err
	}

	// Step 2: Topological sort using Kahn's algorithm
	result, err := r.topologicalSort(graph, inDegree, allNodes)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// buildGraph builds dependency graph recursively
func (r *dependencyResolver) buildGraph(
	models []string,
	graph map[string][]string,
	inDegree map[string]int,
	allNodes map[string]bool,
) error {

	for _, modelName := range models {
		// Skip nếu đã processed
		if allNodes[modelName] {
			continue
		}

		// Mark as seen
		allNodes[modelName] = true

		// Get model từ registry
		model, err := r.registry.Get(modelName)
		if err != nil {
			return fmt.Errorf("failed to get model '%s': %w", modelName, err)
		}

		// Get dependencies
		deps := model.Dependencies()

		// Initialize in-degree nếu chưa có
		if _, exists := inDegree[modelName]; !exists {
			inDegree[modelName] = 0
		}

		// Process each dependency
		for _, dep := range deps {
			// Add edge from dep to modelName trong graph
			// Meaning: modelName depends on dep
			graph[dep] = append(graph[dep], modelName)

			// Increment in-degree của modelName
			inDegree[modelName]++

			// Recursively build graph cho dependency
			// Vì dependency có thể có dependencies của nó
			err := r.buildGraph([]string{dep}, graph, inDegree, allNodes)
			if err != nil {
				return err
			}
		}

		// Initialize empty list trong graph nếu chưa có
		// (để handle models without outgoing edges)
		if _, exists := graph[modelName]; !exists {
			graph[modelName] = []string{}
		}
	}

	return nil
}

// topologicalSort performs Kahn's algorithm
func (r *dependencyResolver) topologicalSort(
	graph map[string][]string,
	inDegree map[string]int,
	allNodes map[string]bool,
) ([]string, error) {

	// Result list sẽ chứa thứ tự execution
	result := make([]string, 0, len(allNodes))

	// Queue chứa nodes có in-degree = 0
	// Những nodes này không depend on anything, có thể run ngay
	queue := make([]string, 0)

	// Find tất cả nodes với in-degree = 0
	for node := range allNodes {
		if inDegree[node] == 0 {
			queue = append(queue, node)
		}
	}

	// Process queue
	for len(queue) > 0 {
		// Dequeue (FIFO)
		current := queue[0]
		queue = queue[1:]

		// Add vào result
		result = append(result, current)

		// Process tất cả nodes that depend on current
		for _, dependent := range graph[current] {
			// Giảm in-degree vì current đã được processed
			inDegree[dependent]--

			// Nếu in-degree becomes 0, add vào queue
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// Check for cycles
	// Nếu result length < total nodes → có cycle
	if len(result) != len(allNodes) {
		// Find nodes that are in cycle
		cycleNodes := make([]string, 0)
		for node := range allNodes {
			found := false
			for _, r := range result {
				if r == node {
					found = true
					break
				}
			}
			if !found {
				cycleNodes = append(cycleNodes, node)
			}
		}

		return nil, fmt.Errorf("circular dependency detected involving: %v", cycleNodes)
	}

	return result, nil
}

// GetDependencyTree returns dependency tree cho visualization
// Useful để debug hoặc display cho user
func (r *dependencyResolver) GetDependencyTree(modelName string) (*DependencyNode, error) {
	model, err := r.registry.Get(modelName)
	if err != nil {
		return nil, err
	}

	node := &DependencyNode{
		Name:         modelName,
		Dependencies: make([]*DependencyNode, 0),
	}

	// Recursively build tree
	for _, dep := range model.Dependencies() {
		depNode, err := r.GetDependencyTree(dep)
		if err != nil {
			return nil, err
		}
		node.Dependencies = append(node.Dependencies, depNode)
	}

	return node, nil
}

// ValidateDependencies validates tất cả dependencies are resolvable
// Useful để check lúc startup
func (r *dependencyResolver) ValidateDependencies() error {
	// Get tất cả models
	allModels := r.registry.List()

	// Try resolve từng model
	for _, modelName := range allModels {
		_, err := r.Resolve([]string{modelName})
		if err != nil {
			return fmt.Errorf("validation failed for model '%s': %w", modelName, err)
		}
	}

	return nil
}
