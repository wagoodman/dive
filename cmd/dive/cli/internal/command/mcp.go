package command

import (
	"github.com/anchore/clio"
	"github.com/spf13/cobra"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/mcp"
	"github.com/wagoodman/dive/cmd/dive/cli/internal/options"
)

type mcpOptions struct {
	options.Application `yaml:",inline" mapstructure:",squash"`
}

func MCP(app clio.Application, id clio.Identification) *cobra.Command {
	opts := &mcpOptions{
		Application: options.DefaultApplication(),
	}
	return app.SetupCommand(&cobra.Command{
		Use:   "mcp",
		Short: "Start the Model Context Protocol (MCP) server.",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := mcp.NewServer(id, opts.MCP)
			return mcp.Run(id, s, opts.MCP)
		},
	}, opts)
}
