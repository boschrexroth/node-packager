# Node-Packager

Packs a Node-RED library to a tarball (*.tgz) for ctrlX Node-RED palette import.

_IMPORTANT__:

Imported libraries override libraries, shipped with the Node-RED App.
This libraries won't be updated on future App updates.

To enable the library update, navigate to

    {USERDIR}/node_modules/{INSTALLED_LIBRARY} 

and remove the installed library, before you run the update

## Prerequisites

### Node.js
Please ensure [Node.js](https://nodejs.org/en/download) to be installed on your system. We recommend to use the latest LTS version.

### Proxy
If you're behind a corporate proxy and encounter any issues, try a proxy like [Px](https://github.com/genotrance/px/releases) and edit your _.npmrc_ file located in your _HOME_ directory as following.

__Example (Px):__

    proxy=http://localhost:3128
    http-proxy=http://localhost:3128
    https-proxy=http://localhost:3128
    maxsockets=1

## Usage

    node-packager [-options] library
Please see

    node-packager --help

## Examples

Packs a library:

    node-packager node-red-contrib-ctrlx-automation

Packs a library with verbose output:

    node-packager --verbose node-red-contrib-ctrlx-automation

Packs a library without vulnerability audit:

    node-packager --no-audit node-red-contrib-ctrlx-automation 

Packs a library with vulnerability fix:

    node-packager --audit-fix node-red-contrib-ctrlx-automation 

## Native packages

Native packages must be packed on a Linux OS with same architecture as the OS the library should be imported, if not shipped with binary prebuilds for all OS supported architectures.

The _node-packager_ detects missing prebuilds and outputs a _WARNING_, if the  tarball can only be imported on a OS with same OS architecture it was build.

# Tarball Import
Install your packed tarball in Node-RED

- Navigate to the login page and open the Node-RED editor.
- Click on _Manage palette_ in the left Node-RED menu.
- Click on tab _Installation_.
- Click the _Upload Button_ and choose your packed library for upload.

In some cases you have to restart the Node-RED App to apply the changes. 

## Options

| Name         |  Description    
|--------------|------------
| src          | Packs an already installed library directory, instead of fetching it from the npm registry.
| registry     | The node registry. (default "https://registry.npmjs.org")
| proxy        | Sets the proxy (e.g. if behind a corporate proxy).
| audit-level  | The audit level used for security auditing and fixing: (low,moderate,high,critical). (default "high")
| no-audit     | Disables vulnerability audit.
| audit-fix    | Enables vulnerability fix.
| verbose      | Enables verbosity.
| keep-tmp     | Prevents the used temporary directory from beeing removed after packing.
| version      | Prints the version.
| help         | Prints the help. 


## Important directions for use

### Areas of use and application

The content (e.g. source code and related documents) of this repository is intended to be used for configuration, parameterization, programming or diagnostics in combination with selected Bosch Rexroth ctrlX AUTOMATION devices.
Additionally, the specifications given in the "Areas of Use and Application" for ctrlX AUTOMATION devices used with the content of this repository do also apply.

### Unintended use

Any use of the source code and related documents of this repository in applications other than those specified above or under operating conditions other than those described in the documentation and the technical specifications is considered as "unintended". Furthermore, this software must not be used in any application areas not expressly approved by Bosch Rexroth.

## About

SPDX-FileCopyrightText: Copyright (c) 2024 Bosch Rexroth AG

<https://www.boschrexroth.com/en/dc/imprint/>

## Licenses

SPDX-License-Identifier: AGPL-3.0-or-later
