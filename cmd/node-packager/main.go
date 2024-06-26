/*
 * SPDX-FileCopyrightText: Bosch Rexroth AG
 *
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */
package main

import (
	"flag"
	"fmt"
	"os"

	"lo-stash.de.bosch.com/icx/tool.node-packager.git/pkg/packager"
	"lo-stash.de.bosch.com/icx/tool.node-packager.git/pkg/utils"
)

const (
	Version       = "1.0.0"
	Name          = "Node-RED Node Packager"
	Description   = "Packs a Node-RED library to a tarball (*.tgz) for Node-RED palette import."
	Examples      = "Examples:\n   node-packager node-red-contrib-ctrlx-automation\n   node-packager --verbose node-red-contrib-ctrlx-automation\n   node-packager --no-audit node-red-contrib-ctrlx-automation\n   node-packager --audit-fix node-red-contrib-ctrlx-automation"
	Prerequisites = "Prerequisites:\n - Ensure Node.js to be installed on your system."
	Installation  = "Import:\n - Navigate to the login page and open the Node-RED Editor.\n - Click on 'Manage palette' in the left Node-RED menu.\n - Click on tab 'Installation'.\n - Click the 'Upload Button' and choose your packed library for upload.\n - In some cases you have to restart the Node-RED App to apply the changes."
	Notes         = "IMPORTANT:\n - Imported libraries override libraries, shipped with the Node-RED App.\n - This libraries won't be updated on future App updates.\n - To enable the library update, navigate to '{USERDIR}/node_modules/{INSTALLED_LIBRARY}'\n   and remove the installed library, before you run the update."
)

var (
	// Arguments
	argSrc        = flag.String("src", "", "Packs an already installed library directory, instead of fetching it from the npm registry.")
	argRegistry   = flag.String("registry", "https://registry.npmjs.org", "The npm registry.")
	argProxy      = flag.String("proxy", "", "Sets the proxy (e.g. if behind a corporate proxy).")
	argAuditLevel = flag.String("audit-level", "high", "The audit level used for security auditing and fixing: (low,moderate,high,critical).")
	argVerbose    = flag.Bool("verbose", false, "Enables verbosity.")
	argNoAudit    = flag.Bool("no-audit", false, "Disables vulnerability audit.")
	argAuditFix   = flag.Bool("audit-fix", false, "Enables vulnerability fix.")
	argKeepTmp    = flag.Bool("keep-tmp", false, "Prevents the used temporary directory from beeing removed after packing.")
	argVersion    = flag.Bool("version", false, "Prints the version.")
	argHelp       = flag.Bool("help", false, "Prints the help.")
)

func main() {

	flag.Parse()

	if *argVersion {
		fmt.Println(Version)
		os.Exit(0)
	}

	// Banner
	fmt.Printf("%s %s\n%s\n\n%s\n\n%s\n\n%s\n\n%s\n", Name, Version, Description, Examples, Prerequisites, Installation, Notes)
	fmt.Println()

	// Help
	if *argHelp {
		fmt.Println()
		flag.Usage()
		os.Exit(0)
	}

	// Run
	p := packager.NodePackager{
		LibraryName: flag.Arg(0),
		Proxy:       *argProxy,
		Registry:    *argRegistry,
		SrcDir:      *argSrc,
		NoAudit:     *argNoAudit,
		AuditFix:    *argAuditFix,
		AuditLevel:  *argAuditLevel,
		Verbose:     *argVerbose,
		KeepTmp:     *argKeepTmp,
	}

	if _, err := p.Pack(); err != nil {
		fmt.Println()
		fmt.Println()
		flag.Usage()
		utils.Errorfln("%v", err.Error())
		os.Exit(1)
	}
}
