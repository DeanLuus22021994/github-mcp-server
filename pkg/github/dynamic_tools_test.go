package github

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/github/github-mcp-server/pkg/toolsets"
	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/require"
)

func Test_ListAvailableTool(t *testing.T) {
	// Verify tool definition
	toolsetGroup := toolsets.NewToolsetGroup(false)

	// Create toolsets with different states
	toolset1 := toolsets.NewToolset("toolset1", "Test toolset 1")
	toolset1.Enabled = true
	toolset2 := toolsets.NewToolset("toolset2", "Test toolset 2")
	toolset3 := toolsets.NewToolset("toolset3", "Test toolset 3")
	toolset3.Enabled = true

	// Add the toolsets to the toolset group
	toolsetGroup.AddToolset(toolset1)
	toolsetGroup.AddToolset(toolset2)
	toolsetGroup.AddToolset(toolset3)

	// Test ListAvailableToolsets tool definition
	tool, _ := ListAvailableToolsets(toolsetGroup, translations.NullTranslationHelper)
	require.Equal(t, "list_available_toolsets", tool.Name)
	require.NotEmpty(t, tool.Description)
	require.Empty(t, tool.InputSchema.Required) // No required parameters

	tests := []struct {
		name            string
		toolsetGroup    *toolsets.ToolsetGroup
		requestArgs     map[string]interface{}
		expectError     bool
		expectedErrMsg  string
		expectedEnabled map[string]bool
	}{
		{
			name: "regular toolset group",
			toolsetGroup: func() *toolsets.ToolsetGroup {
				fs := toolsets.NewToolsetGroup(false)

				// Create toolsets
				toolset1 := toolsets.NewToolset("toolset1", "Test toolset 1")
				toolset1.Enabled = true
				toolset2 := toolsets.NewToolset("toolset2", "Test toolset 2")
				toolset3 := toolsets.NewToolset("toolset3", "Test toolset 3")
				toolset3.Enabled = true

				// Add toolsets to group
				fs.AddToolset(toolset1)
				fs.AddToolset(toolset2)
				fs.AddToolset(toolset3)
				return fs
			}(),
			requestArgs: map[string]interface{}{},
			expectError: false,
			expectedEnabled: map[string]bool{
				"toolset1": true,
				"toolset2": false,
				"toolset3": true,
			},
		},
		{
			name:            "empty toolset group",
			toolsetGroup:    toolsets.NewToolsetGroup(false),
			requestArgs:     map[string]interface{}{},
			expectError:     false,
			expectedEnabled: map[string]bool{},
		},
		{
			name: "toolset group with everything enabled",
			toolsetGroup: func() *toolsets.ToolsetGroup {
				fs := toolsets.NewToolsetGroup(false)

				// Create toolsets
				toolset1 := toolsets.NewToolset("toolset1", "Test toolset 1")
				toolset2 := toolsets.NewToolset("toolset2", "Test toolset 2")

				// Add toolsets to group
				fs.AddToolset(toolset1)
				fs.AddToolset(toolset2)

				_ = fs.EnableToolset("all")
				return fs
			}(),
			requestArgs: map[string]interface{}{},
			expectError: false,
			expectedEnabled: map[string]bool{
				"toolset1": true,
				"toolset2": true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Get tool handler
			_, handler := ListAvailableToolsets(tc.toolsetGroup, translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)

			// Parse result and get text content
			textContent := getTextResult(t, result)

			// Unmarshal the result to verify toolsets
			var returnedToolsets map[string]bool
			err = json.Unmarshal([]byte(textContent.Text), &returnedToolsets)
			require.NoError(t, err)

			// Verify the toolsets match what we expect
			require.Equal(t, tc.expectedEnabled, returnedToolsets)
		})
	}
}

func Test_EnableToolset(t *testing.T) {
	// Create a mock MCP server for testing
	mockServer := &server.MCPServer{}

	// Verify tool definition
	toolsetGroup := toolsets.NewToolsetGroup(false)

	// Create toolsets
	toolset1 := toolsets.NewToolset("toolset1", "Test toolset 1")
	toolset2 := toolsets.NewToolset("toolset2", "Test toolset 2")

	// Add toolsets to the group
	toolsetGroup.AddToolset(toolset1)
	toolsetGroup.AddToolset(toolset2)

	// Test EnableToolset tool definition
	tool, _ := EnableToolset(mockServer, toolsetGroup, translations.NullTranslationHelper)
	require.Equal(t, "enable_toolset", tool.Name)
	require.NotEmpty(t, tool.Description)
	require.Equal(t, 1, len(tool.InputSchema.Required))
	require.Equal(t, "toolset", tool.InputSchema.Required[0])

	// Verify toolset enum has the correct values
	prop := tool.InputSchema.Properties["toolset"]
	propJson, err := json.Marshal(prop)

	require.NoError(t, err)
	require.NotEmpty(t, propJson)

	// Unmarshal the json to verify its values
	var propMap map[string]interface{}
	err = json.Unmarshal(propJson, &propMap)
	require.NoError(t, err)
	require.NotNil(t, propMap["enum"])
	enum := propMap["enum"].([]interface{})

	// get the string values from the []interface{} and compare them
	stringEnum := make([]string, len(enum))
	for i, v := range enum {
		stringEnum[i] = v.(string)
	}

	require.ElementsMatch(t, []string{"toolset1", "toolset2"}, stringEnum)

	tests := []struct {
		name           string
		toolsetGroup   *toolsets.ToolsetGroup
		requestArgs    map[string]interface{}
		expectError    bool
		expectedErrMsg string
		expectedResult string
	}{
		{
			name: "enable valid toolset",
			toolsetGroup: func() *toolsets.ToolsetGroup {
				group := toolsets.NewToolsetGroup(false)

				// Create toolset
				toolset1 := toolsets.NewToolset("toolset1", "Test toolset 1")

				// Add toolset to group
				group.AddToolset(toolset1)
				return group
			}(),
			requestArgs:    map[string]interface{}{"toolset": "toolset1"},
			expectError:    false,
			expectedResult: "Toolset toolset1 enabled",
		},
		{
			name: "toolset already enabled",
			toolsetGroup: func() *toolsets.ToolsetGroup {
				group := toolsets.NewToolsetGroup(false)

				// Create toolset
				toolset1 := toolsets.NewToolset("toolset1", "Test toolset 1")
				toolset1.Enabled = true

				// Add toolset to group
				group.AddToolset(toolset1)
				return group
			}(),
			requestArgs:    map[string]interface{}{"toolset": "toolset1"},
			expectError:    false,
			expectedResult: "Toolset toolset1 is already enabled",
		},
		{
			name: "non-existent toolset",
			toolsetGroup: func() *toolsets.ToolsetGroup {
				return toolsets.NewToolsetGroup(false)
			}(),
			requestArgs:    map[string]interface{}{"toolset": "non-existent"},
			expectError:    true,
			expectedErrMsg: "Toolset non-existent not found",
		},
		{
			name: "missing required parameter",
			toolsetGroup: func() *toolsets.ToolsetGroup {
				return toolsets.NewToolsetGroup(false)
			}(),
			requestArgs:    map[string]interface{}{},
			expectError:    true,
			expectedErrMsg: "missing required parameter: toolset",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock server for each test
			mockServer := &server.MCPServer{}

			// Get tool handler
			_, handler := EnableToolset(mockServer, tc.toolsetGroup, translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				require.True(t, result.IsError)
				content, err := json.Marshal(result.Content)
				require.NoError(t, err)
				require.Contains(t, string(content), tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			textContent := getTextResult(t, result)
			require.Equal(t, tc.expectedResult, textContent.Text)

			// If we enabled a toolset, verify it's now enabled
			if tc.requestArgs["toolset"] != nil && !tc.expectError && !strings.Contains(tc.expectedResult, "already enabled") {
				toolsetName := tc.requestArgs["toolset"].(string)
				require.True(t, tc.toolsetGroup.IsEnabled(toolsetName))
			}
		})
	}
}

func Test_GetToolsetsTools(t *testing.T) {
	var nilHandler server.ToolHandlerFunc = func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return nil, nil
	}

	// Verify tool definition
	toolsetGroup := toolsets.NewToolsetGroup(false)

	// Create toolsets
	toolset1 := toolsets.NewToolset("toolset1", "Test toolset 1")
	toolset1.Enabled = true
	toolset2 := toolsets.NewToolset("toolset2", "Test toolset 2")

	// Add toolsets to the group
	toolsetGroup.AddToolset(toolset1)
	toolsetGroup.AddToolset(toolset2)

	// Add some mock toolsets.NewServerTools to the toolset
	toolset1.AddReadTools(
		toolsets.NewServerTool(mcp.NewTool("tool1", mcp.WithDescription("Tool 1 description")), nilHandler),
		toolsets.NewServerTool(mcp.NewTool("tool2", mcp.WithDescription("Tool 2 description")), nilHandler),
	)

	// Test GetToolsetsTools tool definition
	tool, _ := GetToolsetsTools(toolsetGroup, translations.NullTranslationHelper)
	require.Equal(t, "get_toolset_tools", tool.Name)
	require.NotEmpty(t, tool.Description)
	require.Equal(t, 1, len(tool.InputSchema.Required))
	require.Equal(t, "toolset", tool.InputSchema.Required[0])

	// Verify toolset enum has the correct values
	prop := tool.InputSchema.Properties["toolset"]
	propJson, err := json.Marshal(prop)
	require.NoError(t, err)
	require.NotEmpty(t, propJson)

	// Unmarshal the json to verify its values
	var propMap map[string]interface{}
	err = json.Unmarshal(propJson, &propMap)
	require.NoError(t, err)
	require.NotNil(t, propMap["enum"])
	enum := propMap["enum"].([]interface{})

	// get the string values from the []interface{} and compare them
	stringEnum := make([]string, len(enum))
	for i, v := range enum {
		stringEnum[i] = v.(string)
	}

	require.ElementsMatch(t, []string{"toolset1", "toolset2"}, stringEnum)

	tests := []struct {
		name           string
		toolsetGroup   *toolsets.ToolsetGroup
		requestArgs    map[string]interface{}
		expectError    bool
		expectedErrMsg string
		expectedTools  map[string]string
	}{
		{
			name: "get tools for valid toolset with tools",
			toolsetGroup: func() *toolsets.ToolsetGroup {
				group := toolsets.NewToolsetGroup(false)

				// Create toolset
				toolset1 := toolsets.NewToolset("toolset1", "Test toolset 1")
				toolset1.Enabled = true

				// Add toolset to group
				group.AddToolset(toolset1)

				// Add tools to the toolset
				toolset1.AddReadTools(
					toolsets.NewServerTool(mcp.NewTool("tool1", mcp.WithDescription("Tool 1 description")), nilHandler),
					toolsets.NewServerTool(mcp.NewTool("tool2", mcp.WithDescription("Tool 2 description")), nilHandler),
				)
				return group
			}(),
			requestArgs: map[string]interface{}{"toolset": "toolset1"},
			expectError: false,
			expectedTools: map[string]string{
				"tool1": "Tool 1 description",
				"tool2": "Tool 2 description",
			},
		},
		{
			name: "get tools for valid toolset without tools",
			toolsetGroup: func() *toolsets.ToolsetGroup {
				group := toolsets.NewToolsetGroup(false)

				// Create toolset
				toolset2 := toolsets.NewToolset("toolset2", "Test toolset 2")
				toolset2.Enabled = true

				// Add toolset to group
				group.AddToolset(toolset2)
				return group
			}(),
			requestArgs:   map[string]interface{}{"toolset": "toolset2"},
			expectError:   false,
			expectedTools: map[string]string{},
		},
		{
			name: "non-existent toolset",
			toolsetGroup: func() *toolsets.ToolsetGroup {
				return toolsets.NewToolsetGroup(false)
			}(),
			requestArgs:    map[string]interface{}{"toolset": "non-existent"},
			expectError:    true,
			expectedErrMsg: "Toolset non-existent not found",
		},
		{
			name: "missing required parameter",
			toolsetGroup: func() *toolsets.ToolsetGroup {
				return toolsets.NewToolsetGroup(false)
			}(),
			requestArgs:    map[string]interface{}{},
			expectError:    true,
			expectedErrMsg: "missing required parameter: toolset",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Get tool handler
			_, handler := GetToolsetsTools(tc.toolsetGroup, translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				require.True(t, result.IsError)
				content, err := json.Marshal(result.Content)
				require.NoError(t, err)
				require.Contains(t, string(content), tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			textContent := getTextResult(t, result)

			// Parse the JSON result to verify tools
			var returnedTools map[string]string
			err = json.Unmarshal([]byte(textContent.Text), &returnedTools)
			require.NoError(t, err)

			// Verify the returned tools match what we expect
			require.Equal(t, tc.expectedTools, returnedTools)
		})
	}
}
