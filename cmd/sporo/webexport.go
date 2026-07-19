package main

import (
	"fmt"

	"github.com/spf13/cobra"
	sporo "sporo.dev/sporo"
	"sporo.dev/sporo/internal/recipe"
)

// webMirrorCmd regenerates the committed export forms the site serves as each recipe's `.md`
// mirror (web/src/data/exports/<slug>.md), byte-for-byte the `sporo export <slug>` output. It is
// Hidden because it is a build tool, not a user verb — the user-facing composer is `sporo export`;
// this one exists so `go generate` can refresh the committed mirror and a git-diff gate can prove
// the site never drifts from the binary's handover file. The composition itself lives once, in
// recipe.Export; this command only fans it across the corpus and writes the results.
func webMirrorCmd() *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:    "web-mirror",
		Short:  "Regenerate the site's committed recipe export mirror (used by `go generate`)",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := recipe.WriteMirror(sporo.Recipes, "", out)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "sporo web-mirror: wrote %d export form(s) to %s\n", n, out)
			return nil
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "directory for the <slug>.md export forms (used by `go generate`)")
	_ = cmd.MarkFlagRequired("out")
	return cmd
}
