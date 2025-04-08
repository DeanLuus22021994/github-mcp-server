package github

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/github/github-mcp-server/pkg/toolsets"
	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func ToolsetEnum(toolsetGroup *toolsets.ToolsetGroup) mcp.PropertyOption {
	toolsetNames := make([]string, 0, len(toolsetGroup.Toolsets))
	for name := range toolsetGroup.Toolsets {
		toolsetNames = append(toolsetNames, name)
	}
	return mcp.Enum(toolsetNames...)
}

func EnableToolset(s *server.MCPServer, toolsetGroup *toolsets.ToolsetGroup, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("enable_toolset",
			mcp.WithDescription(t("TOOL_ENABLE_TOOLSET_DESCRIPTION", "Enable one of the sets of tools the GitHub MCP server provides, to access the tools and accomplish your goals")),
			mcp.WithString("toolset",
				mcp.Required(),
				mcp.Description("The name of the toolset to enable"),
				ToolsetEnum(toolsetGroup),
			),
		),
		func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// We need to convert the toolsets back to a map for JSON serialization
			toolsetName, err := requiredParam[string](request, "toolset")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			toolset := toolsetGroup.Toolsets[toolsetName]
			if toolset == nil {
				return mcp.NewToolResultError(fmt.Sprintf("Toolset %s not found", toolsetName)), nil
			}
			if toolset.Enabled {
				return mcp.NewToolResultText(fmt.Sprintf("Toolset %s is already enabled", toolsetName)), nil
			}

			toolset.Enabled = true

			// caution: this currently affects the global tools and notifies all clients:
			//
			// Send notification to all initialized sessions
			// s.sendNotificationToAllClients("notifications/tools/list_changed", nil)
			s.AddTools(toolset.GetActiveTools()...)

			return mcp.NewToolResultText(fmt.Sprintf("Toolset %s enabled", toolsetName)), nil
		}
}

func ListAvailableToolsets(toolsetGroup *toolsets.ToolsetGroup, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("list_available_toolsets",
			mcp.WithDescription(t("TOOL_LIST_AVAILABLE_FEATURES_DESCRIPTION", "List all available toolsets this GitHub MCP server can offer, providing the enabled status of each. Use this when you think a task could be achieved with a GitHub product, but you don't think the currently available tools could achieve it")),
		),
		func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// We need to convert the toolsetGroup back to a map for JSON serialization
			featureMap := make(map[string]bool)
			for name := range toolsetGroup.Toolsets {
				featureMap[name] = toolsetGroup.IsEnabled(name)
			}

			r, err := json.Marshal(featureMap)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal features: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

func GetToolsetsTools(toolsetGroup *toolsets.ToolsetGroup, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_toolset_tools",
			mcp.WithDescription(t("TOOL_GET_TOOLSET_TOOLS_DESCRIPTION", "Lists all the capabilities that are enabled when you enabled a toolset, use this to get clarity on whether enabling a toolset would help you to complete a task")),
			mcp.WithString("toolset",
				mcp.Required(),
				mcp.Description("The name of the toolset you want to get the tools for"),
				ToolsetEnum(toolsetGroup),
			),
		),
		func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// We need to convert the toolsetGroup back to a map for JSON serialization
			toolsetName, err := requiredParam[string](request, "toolset")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			toolset := toolsetGroup.Toolsets[toolsetName]
			if toolset == nil {
				return mcp.NewToolResultError(fmt.Sprintf("Toolset %s not found", toolsetName)), nil
			}

			tools := make(map[string]string)

			for _, st := range toolset.GetActiveTools() {
				tools[st.Tool.Name] = st.Tool.Description
			}

			r, err := json.Marshal(tools)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal features: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}
