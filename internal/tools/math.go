package tools

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/kringen/go-mcp-server/pkg/mcp"
)

// MathToolProvider provides basic mathematical operations
type MathToolProvider struct{}

// NewMathToolProvider creates a new math tool provider
func NewMathToolProvider() *MathToolProvider {
	return &MathToolProvider{}
}

// ListTools returns the list of available math tools
func (m *MathToolProvider) ListTools(ctx context.Context) ([]mcp.Tool, error) {
	return []mcp.Tool{
		{
			Name:        "add",
			Description: "Add two numbers together",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"a": map[string]interface{}{
						"type":        "number",
						"description": "First number",
					},
					"b": map[string]interface{}{
						"type":        "number",
						"description": "Second number",
					},
				},
				"required": []string{"a", "b"},
			},
		},
		{
			Name:        "multiply",
			Description: "Multiply two numbers",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"a": map[string]interface{}{
						"type":        "number",
						"description": "First number",
					},
					"b": map[string]interface{}{
						"type":        "number",
						"description": "Second number",
					},
				},
				"required": []string{"a", "b"},
			},
		},
		{
			Name:        "power",
			Description: "Calculate a number raised to a power",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"base": map[string]interface{}{
						"type":        "number",
						"description": "Base number",
					},
					"exponent": map[string]interface{}{
						"type":        "number",
						"description": "Exponent",
					},
				},
				"required": []string{"base", "exponent"},
			},
		},
	}, nil
}

// CallTool executes a math tool
func (m *MathToolProvider) CallTool(ctx context.Context, request mcp.ToolCallRequest) (*mcp.ToolCallResponse, error) {
	switch request.Name {
	case "add":
		return m.handleAdd(request)
	case "multiply":
		return m.handleMultiply(request)
	case "power":
		return m.handlePower(request)
	default:
		return &mcp.ToolCallResponse{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Unknown tool: %s", request.Name),
				},
			},
			IsError: true,
		}, nil
	}
}

func (m *MathToolProvider) handleAdd(request mcp.ToolCallRequest) (*mcp.ToolCallResponse, error) {
	a, err := getNumberArg(request.Arguments, "a")
	if err != nil {
		return errorResponse(err.Error()), nil
	}

	b, err := getNumberArg(request.Arguments, "b")
	if err != nil {
		return errorResponse(err.Error()), nil
	}

	result := a + b
	return &mcp.ToolCallResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("%.2f + %.2f = %.2f", a, b, result),
			},
		},
	}, nil
}

func (m *MathToolProvider) handleMultiply(request mcp.ToolCallRequest) (*mcp.ToolCallResponse, error) {
	a, err := getNumberArg(request.Arguments, "a")
	if err != nil {
		return errorResponse(err.Error()), nil
	}

	b, err := getNumberArg(request.Arguments, "b")
	if err != nil {
		return errorResponse(err.Error()), nil
	}

	result := a * b
	return &mcp.ToolCallResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("%.2f × %.2f = %.2f", a, b, result),
			},
		},
	}, nil
}

func (m *MathToolProvider) handlePower(request mcp.ToolCallRequest) (*mcp.ToolCallResponse, error) {
	base, err := getNumberArg(request.Arguments, "base")
	if err != nil {
		return errorResponse(err.Error()), nil
	}

	exponent, err := getNumberArg(request.Arguments, "exponent")
	if err != nil {
		return errorResponse(err.Error()), nil
	}

	result := math.Pow(base, exponent)
	return &mcp.ToolCallResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("%.2f^%.2f = %.2f", base, exponent, result),
			},
		},
	}, nil
}

// Helper functions
func getNumberArg(args map[string]interface{}, key string) (float64, error) {
	val, exists := args[key]
	if !exists {
		return 0, fmt.Errorf("missing required argument: %s", key)
	}

	switch v := val.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number format for %s: %s", key, v)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("invalid argument type for %s: expected number, got %T", key, v)
	}
}

func errorResponse(message string) *mcp.ToolCallResponse {
	return &mcp.ToolCallResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: message,
			},
		},
		IsError: true,
	}
}
