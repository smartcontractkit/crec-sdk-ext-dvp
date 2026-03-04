//go:build wasip1

package main

import (
	"log/slog"

	"github.com/smartcontractkit/cre-sdk-go/cre"
	"github.com/smartcontractkit/cre-sdk-go/cre/wasm"

	workflows "github.com/smartcontractkit/crec-workflow-utils"
	wf "github.com/smartcontractkit/crec-sdk-ext-dvp/watcher/handler"
)

func main() {
	r := wasm.NewRunner(workflows.ParseWorkflowConfig)
	r.Run(func(cfg *workflows.Config, _ *slog.Logger, _ cre.SecretsProvider) (cre.Workflow[*workflows.Config], error) {
		return workflows.InitEventListenerWorkflow(cfg, wf.OnLog)
	})
}
