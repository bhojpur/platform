// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// urlCmd represents the url command
var urlCmd = &cobra.Command{
	Use:   "url [port]",
	Short: "Prints the URL of this Bhojpur.NET Platform application",
	Long: `Prints the URL of this application. This command can print the URL of
the current application itself, or of a service running in this application on a
particular port. For example:
    bp url 8080
will print the URL of a service/server exposed on port 8080.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println(os.Getenv("BHOJPUR_APPLICATION_URL"))
			return
		}

		port, err := strconv.ParseUint(args[0], 10, 16)
		if err != nil {
			fmt.Fprintf(os.Stderr, "port \"%s\" is not a valid number\n", args[0])
			return
		}

		fmt.Println(GetApplicationURL(int(port)))
	},
}

func init() {
	rootCmd.AddCommand(urlCmd)
}

func GetApplicationURL(port int) (url string) {
	wsurl := os.Getenv("BHOJPUR_APPLICATION_URL")
	if port == 0 {
		return wsurl
	}

	serviceurl := wsurl
	serviceurl = strings.Replace(serviceurl, "https://", fmt.Sprintf("https://%d-", port), -1)
	serviceurl = strings.Replace(serviceurl, "http://", fmt.Sprintf("http://%d-", port), -1)
	return serviceurl
}
