/*
Copyright © 2020 Christian González Di Antonio christian@slashdevops.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the application version",
	Long:  `Show the application version and its details.`,
	Run: func(cmd *cobra.Command, args []string) {
		exd, _ := cmd.Flags().GetBool("extended")
		if exd {
			fmt.Printf(
				"version: %s, "+
					"revision: %s, "+
					"branch: %s\n",
				conf.Version,
				conf.Revision,
				conf.Branch)
		} else {
			fmt.Println(conf.Version)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().BoolP("extended", "", false, "Show the extended version information")
}
