package toolsets

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ResourceTemplate struct {
	templateResource mcp.ResourceTemplate
	handler          server.ResourceTemplateHandlerFunc
}

func NewResourceTemplate(templateResource mcp.ResourceTemplate, handler server.ResourceTemplateHandlerFunc) ResourceTemplate {
	return ResourceTemplate{
		templateResource: templateResource,
		handler:          handler,
	}
}

func NewServerTool(tool mcp.Tool, handler server.ToolHandlerFunc) server.ServerTool {
	return server.ServerTool{Tool: tool, Handler: handler}
}

type Toolset struct {
	Name              string
	Description       string
	Enabled           bool
	readOnly          bool
	writeTools        []server.ServerTool
	readTools         []server.ServerTool
	resourceTemplates []ResourceTemplate
}

func (t *Toolset) GetActiveTools() []server.ServerTool {
	if t.Enabled {
		if t.readOnly {
			return t.readTools
		}
		return append(t.readTools, t.writeTools...)
	}
	return nil
}

func (t *Toolset) GetAvailableTools() []server.ServerTool {
	if t.readOnly {
		return t.readTools
	}
	return append(t.readTools, t.writeTools...)
}

func (t *Toolset) RegisterTools(s *server.MCPServer) {
	if !t.Enabled {
		return
	}
	for _, tool := range t.readTools {
		s.AddTool(tool.Tool, tool.Handler)
	}
	if !t.readOnly {
		for _, tool := range t.writeTools {
			s.AddTool(tool.Tool, tool.Handler)
		}
	}
	// resources are read-only
	for _, resource := range t.resourceTemplates {
		s.AddResourceTemplate(resource.templateResource, resource.handler)
	}
}

func (t *Toolset) SetReadOnly() {
	// Set the toolset to read-only
	t.readOnly = true
}

func (t *Toolset) AddWriteTools(tools ...server.ServerTool) *Toolset {
	// Silently ignore if the toolset is read-only to avoid any breach of that contract
	if !t.readOnly {
		t.writeTools = append(t.writeTools, tools...)
	}
	return t
}

func (t *Toolset) AddReadTools(tools ...server.ServerTool) *Toolset {
	t.readTools = append(t.readTools, tools...)
	return t
}

func (t *Toolset) AddTemplateResources(resources ...ResourceTemplate) *Toolset {
	t.resourceTemplates = append(t.resourceTemplates, resources...)
	return t
}

type ToolsetGroup struct {
	Toolsets     map[string]*Toolset
	everythingOn bool
	readOnly     bool
}

func NewToolsetGroup(readOnly bool) *ToolsetGroup {
	return &ToolsetGroup{
		Toolsets:     make(map[string]*Toolset),
		everythingOn: false,
		readOnly:     readOnly,
	}
}

func (tg *ToolsetGroup) AddToolset(ts *Toolset) {
	if tg.readOnly {
		ts.SetReadOnly()
	}
	tg.Toolsets[ts.Name] = ts
}

func NewToolset(name string, description string) *Toolset {
	return &Toolset{
		Name:        name,
		Description: description,
		Enabled:     false,
		readOnly:    false,
	}
}

func (tg *ToolsetGroup) IsEnabled(name string) bool {
	// If everythingOn is true, all features are enabled
	if tg.everythingOn {
		return true
	}

	feature, exists := tg.Toolsets[name]
	if !exists {
		return false
	}
	return feature.Enabled
}

func (tg *ToolsetGroup) EnableToolsets(names []string) error {
	for _, name := range names {
		err := tg.EnableToolset(name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tg *ToolsetGroup) EnableToolset(name string) error {
	// Special case for "all"
	if name == "all" {
		tg.everythingOn = true
		return nil
	}

	toolset, exists := tg.Toolsets[name]
	if !exists {
		return fmt.Errorf("toolset %s does not exist", name)
	}
	toolset.Enabled = true
	tg.Toolsets[name] = toolset
	return nil
}

func (tg *ToolsetGroup) RegisterTools(s *server.MCPServer) {
	for _, toolset := range tg.Toolsets {
		toolset.RegisterTools(s)
	}
}
