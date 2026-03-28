/*
Copyright © 2026 Patrick Hermann patrick.hermann@sva.de
*/

package internal

import (
	"github.com/fatih/color"
	goVersion "go.hein.dev/go-version"
)

var (
	date    = "unknown"
	commit  = "unknown"
	output  = "yaml"
	version = "unset"
)

const logo = `

          ╦ ╦╔═╗  ╦═╗╦ ╦╔╗╔  ╔╦╗╦ ╦╦╔╗╔╔═╗╔═╗
          ║║║║╣   ╠╦╝║ ║║║║   ║ ╠═╣║║║║║ ╦╚═╗
          ╚╩╝╚═╝  ╩╚═╚═╝╝╚╝   ╩ ╩ ╩╩╝╚╝╚═╝╚═╝

`

const banner = `
                    ██╗    ██╗███████╗    ██████╗ ██╗   ██╗███╗   ██╗
                    ██║    ██║██╔════╝    ██╔══██╗██║   ██║████╗  ██║
                    ██║ █╗ ██║█████╗      ██████╔╝██║   ██║██╔██╗ ██║
                    ██║███╗██║██╔══╝      ██╔══██╗██║   ██║██║╚████║
                    ╚███╔███╔╝███████╗    ██║  ██║╚██████╔╝██║ ╚███║
                     ╚══╝╚══╝ ╚══════╝    ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚══╝

                ████████╗██╗  ██╗██╗███╗   ██╗ ██████╗ ███████╗
                ╚══██╔══╝██║  ██║██║████╗  ██║██╔════╝ ██╔════╝
                   ██║   ███████║██║██╔██╗ ██║██║  ███╗███████╗
                   ██║   ██╔══██║██║██║╚████║██║   ██║╚════██║
                   ██║   ██║  ██║██║██║ ╚███║╚██████╔╝███████║
                   ╚═╝   ╚═╝  ╚═╝╚═╝╚═╝  ╚══╝ ╚═════╝ ╚══════╝

`

// OUTPUT BANNER + VERSION OUTPUT
func PrintBanner() string {
	color.Cyan(banner)
	resp := goVersion.FuncWithOutput(false, version, commit, date, output)
	color.Magenta(resp + "\n")
	return resp
}
