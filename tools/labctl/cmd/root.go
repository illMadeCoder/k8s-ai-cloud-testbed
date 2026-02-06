package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "labctl",
	Short: "Lab control tool for experiment-driven learning",
	Long: `labctl provides interactive tutorial experiences for illm-k8s-ai-lab experiments.

It reads Experiment CRs from the hub cluster, extracts kubeconfigs and service
endpoints, and renders rich terminal tutorials using bubbletea.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(tutorialCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(kubeconfigCmd)
}
