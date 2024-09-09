/*
 * SPDX-FileCopyrightText: Bosch Rexroth AG
 *
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */
package packager

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	bar "github.com/schollz/progressbar/v3"
	"lo-stash.de.bosch.com/icx/tool.node-packager.git/pkg/utils"
)

const (
	progressInterval   = time.Duration(300 * time.Millisecond)
	prebuildLinuxArm64 = "linux-arm64"
	prebuildLinuxAmd64 = "linux-x64"
	defaultAuditLevel  = "high"
	defaultRegistry    = "https://registry.npmjs.org"
	spinnerTypeLinux   = 11 // ⣯ (not working on windows) -> ./vendor/github.com/schollz/progressbar/v3/spinners.go
)

type NodePackager struct {
	// API
	LibraryName string
	SrcDir      string
	Proxy       string
	Registry    string
	AuditLevel  string

	Verbose  bool
	NoAudit  bool
	AuditFix bool
	KeepTmp  bool

	// Private
	buildOn                   string
	libraryNameWithoutVersion string
	tarballPrefix             string
	tarballName               string
	tarballPath               string
	tarballSize               string
	tempDir                   string
	workingDir                string

	hasNativeModules                bool
	hasNativePrebuilds              bool
	hasAllNativePrebuildsLinuxArm64 bool
	hasAllNativePrebuildsLinuxAmd64 bool

	spinner   *bar.ProgressBar
	startTime time.Time
}

// pack packs the library and returns the path of the tarball.
func (p *NodePackager) Pack() (string, error) {

	// ******************************************************
	// INITIALIZATION
	// ******************************************************

	// Reset
	p.reset()

	// Validate
	if len(p.LibraryName) == 0 {
		return p.finish(fmt.Errorf("library not defined"))
	}

	if p.hasSrcDir() && !utils.FileExists(p.SrcDir) {
		return p.finish(fmt.Errorf("library source directory not found: %s", p.SrcDir))
	}

	// Start packing
	utils.Printfln("Packing: %s ...", p.LibraryName)

	// Provide progress if no verbosity selected
	if !p.Verbose {
		go func() {
			tick := time.NewTicker(progressInterval)
			for range tick.C {
				if err := p.spinner.Add(1); err != nil {
					return
				}
			}
		}()
	}

	// Get working directory
	err := p.getWorkingDir()
	if err != nil {
		return p.finish(err)
	}

	// Create new temp directory
	err = p.createTempDir()
	if err != nil {
		return p.finish(err)
	}

	// If source is given, copy to temp dir and use as-it-is
	if p.hasSrcDir() {
		p.spinner.Describe("copying ...")
		if err := utils.CopyDir(p.SrcDir, p.tempDir); err != nil {
			return p.finish(err)
		}
	}

	// ******************************************************
	// INSTALLATION
	// ******************************************************
	// If a source directory is not given, fetch library from npm repo
	if !p.hasSrcDir() {
		if err := p.npmInstall(p.LibraryName); err != nil {
			return p.finish(err)
		}
		if err := p.moveLibraryModule(); err != nil {
			return p.finish(err)
		}
	} else {
		if err := p.npmInstall(""); err != nil {
			return p.finish(err)
		}
	}

	// ******************************************************
	// AUDITING
	// ******************************************************
	if err := p.npmAuditFix(); err != nil {
		return p.finish(err)
	}

	if err := p.npmAudit(); err != nil {
		return p.finish(err)
	}

	// ******************************************************
	// MODIFYING
	// ******************************************************
	if err := p.modifyPackageJSON(); err != nil {
		return p.finish(err)
	}

	if err := p.scanNativeModules(); err != nil {
		return p.finish(err)
	}

	// ******************************************************
	// VALIDATING
	// ******************************************************
	if err := p.validateNativeModules(); err != nil {
		return p.finish(err)
	}

	// ******************************************************
	// PREPARING
	// ******************************************************
	if err := p.npmBundleDeps(); err != nil {
		return p.finish(err)
	}

	// ******************************************************
	// PACKING
	// ******************************************************
	if err := p.npmPack(); err != nil {
		return p.finish(err)
	}

	// ******************************************************
	// COPYING
	// ******************************************************
	if err := p.movePackage(); err != nil {
		return p.finish(err)
	}

	// Success
	return p.finish(nil)
}

// reset resets to default values.
func (p *NodePackager) reset() {

	p.hasAllNativePrebuildsLinuxAmd64 = true // Default
	p.hasAllNativePrebuildsLinuxArm64 = true // Default
	p.buildOn = fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	p.SrcDir = os.ExpandEnv(p.SrcDir)

	versionIndex := strings.LastIndex(p.LibraryName, "@")
	if versionIndex <= 0 {
		// no version present
		p.libraryNameWithoutVersion = p.LibraryName
	} else {
		p.libraryNameWithoutVersion = p.LibraryName[:versionIndex]
	}
	p.tarballPrefix = strings.ReplaceAll(strings.ReplaceAll(p.libraryNameWithoutVersion, "/", "-"), "@", "")

	if runtime.GOOS == "linux" {
		p.spinner = bar.NewOptions(-1, bar.OptionSpinnerType(spinnerTypeLinux))
	} else {
		p.spinner = bar.NewOptions(-1)
	}
	p.startTime = time.Now()

	if len(p.AuditLevel) == 0 {
		p.AuditLevel = defaultAuditLevel
	}

	if len(p.Registry) == 0 {
		p.Registry = defaultRegistry
	}
}

// finish finishes the operation with or without any error.
func (p *NodePackager) finish(err error) (string, error) {

	// Clean up
	if err := p.clean(); err != nil {
		utils.Printfln("failed to cleanup: %v", err.Error()) // we don't care about errors here
	}

	// Failed
	if err != nil {
		return "", err
	}

	if err := p.spinner.Close(); err != nil {
		if p.Verbose {
			log.Printf("failed to close spinner: %v", err.Error()) // we don't care about errors here
		}
	}
	p.spinner.Describe("")

	// Succeeded
	var summary string
	if p.hasNativeModules {
		summary = "Native "
	}
	summary = fmt.Sprintf("%sLibrary '%s' [%s] successfully packed in %s.", summary, p.tarballName, p.tarballSize, time.Since(p.startTime))
	fmt.Println(summary)

	return p.tarballPath, nil
}

// clean cleans up the used temp directory.
func (p *NodePackager) clean() error {

	if !utils.FileExists(p.tempDir) {
		return nil
	}

	// Clean up (don't care about errors)
	// Debug: We keep the working and temp dir
	if !p.KeepTmp {
		p.spinner.Describe("cleaning ...")

		if err := p.removeDir(p.tempDir); err != nil {
			return err
		}
	}

	return nil
}

// npmInstall runs 'npm install [pkg]' or just npm install if no library given.
func (p *NodePackager) npmInstall(packageName string) error {
	p.spinner.Describe("downloading ...")

	npmArgs := []string{
		"install",
		fmt.Sprintf("--prefix=%s", p.tempDir),
		fmt.Sprintf("--registry=%s", p.Registry),
		"--no-fund",
		"--omit=dev",
		"--no-audit"}

	if len(p.Proxy) > 0 {
		npmArgs = append(npmArgs, fmt.Sprintf("--proxy=%s", p.Proxy))
	}

	if p.Verbose {
		npmArgs = append(npmArgs, "--verbose")
	}

	if len(packageName) > 0 {
		npmArgs = append(npmArgs, packageName)
	}

	return p.execute(exec.Command("npm", npmArgs...))
}

// npmBundleDeps installs module 'https://www.npmjs.com/package/bundle-deps' and runs 'bundle-deps' for current library to bundle the dependencies.
// 'bundle-deps' tells 'npm pack', which 'node_modules' to embed in the tarball, otherwise 'npm pack' omits the 'node_modules' of the library at all.
// 'bundle-deps' has to be executed in the library directory, because it searches for an existing 'package.json' in current directory.
func (p *NodePackager) npmBundleDeps() error {

	p.spinner.Describe("prepare bundleing ...")

	npmArgs := []string{
		"install",
		fmt.Sprintf("--prefix=%s", p.tempDir),
		fmt.Sprintf("--registry=%s", p.Registry),
		"bundle-deps",
		"--no-save",
		"--ignore-scripts",
		"--no-audit"}

	if len(p.Proxy) > 0 {
		npmArgs = append(npmArgs, fmt.Sprintf("--proxy=%s", p.Proxy))
	}
	if p.Verbose {
		npmArgs = append(npmArgs, "--verbose")
	}

	// Install bundle-deps
	err := p.execute(exec.Command("npm", npmArgs...))
	if err != nil {
		return err
	}

	p.spinner.Describe("bundleing ...")

	// Execute bundle-deps
	nodeArgs := []string{
		path.Join(p.tempDir, "node_modules", "bundle-deps", "bundle-deps.js")}

	// Reset working dir
	defer func() {
		if p.Verbose {
			utils.Printfln("reset working dir to %s ...", p.workingDir)
		}

		if err := os.Chdir(p.workingDir); err != nil {
			p.spinner.Describe(fmt.Sprintf("failed to reset working dir: %v", err))
			utils.Warnfln("failed to reset working dir: %v", err)
		}
	}()

	if p.Verbose {
		utils.Printfln("changing working dir to %s ...", p.tempDir)
	}

	// We need to change current working dir to temp dir to make bundle-deps work correctly,
	// which needs to be run at package.json level.
	if err := os.Chdir(p.tempDir); err != nil {
		return err
	}

	return p.execute(exec.Command("node", nodeArgs...))
}

// npmAuditFix runs 'npm audit fix' in production mode.
func (p *NodePackager) npmAuditFix() error {

	if !p.AuditFix {
		utils.Infofln("vulnerability fix disabled: enable with option '--audit-fix'")
		return nil
	}

	p.spinner.Describe(fmt.Sprintf("vulnerability fix (level: %v) ...", p.AuditLevel))

	npmArgs := []string{
		"audit",
		"fix",
		fmt.Sprintf("--prefix=%s", p.tempDir),
		fmt.Sprintf("--registry=%s", defaultRegistry),
		fmt.Sprintf("--audit-level=%s", p.AuditLevel),
		"--only=prod",
		"--omit=dev",
	}

	if len(p.Proxy) > 0 {
		npmArgs = append(npmArgs, fmt.Sprintf("--proxy=%s", p.Proxy))
	}

	if p.Verbose {
		npmArgs = append(npmArgs, "--verbose")
	}

	if err := p.execute(exec.Command("npm", npmArgs...)); err != nil {
		return fmt.Errorf("vulnerability fix (level: %v) failed: try again with a decreased 'audit-level'", p.AuditLevel)
	}

	return nil
}

// npmAudit runs 'npm audit' with audit level 'high' in production mode.
func (p *NodePackager) npmAudit() error {

	if p.NoAudit {
		utils.Warnfln("vulnerability audit disabled")
		return nil
	}

	p.spinner.Describe(fmt.Sprintf("vulnerability audit (level: %v) ...", p.AuditLevel))

	npmArgs := []string{
		"audit",
		fmt.Sprintf("--prefix=%s", p.tempDir),
		fmt.Sprintf("--registry=%s", defaultRegistry),
		fmt.Sprintf("--audit-level=%s", p.AuditLevel),
		"--only=prod",
		"--omit=dev",
	}

	if len(p.Proxy) > 0 {
		npmArgs = append(npmArgs, fmt.Sprintf("--proxy=%s", p.Proxy))
	}

	if p.Verbose {
		npmArgs = append(npmArgs, "--verbose")
	}

	if err := p.execute(exec.Command("npm", npmArgs...)); err != nil {
		return fmt.Errorf("vulnerability audit (level: %v) failed: try again with '--no-audit' or increased 'audit-level'", p.AuditLevel)
	}

	return p.execute(exec.Command("npm", npmArgs...))
}

// npmPack packs the package to a tarball (*.tgz).
func (p *NodePackager) npmPack() error {
	p.spinner.Describe("packing ...")

	npmArgs := []string{
		"pack",
		fmt.Sprintf("--prefix=%s", p.tempDir),
		fmt.Sprintf("--registry=%s", p.Registry),
	}

	if len(p.Proxy) > 0 {
		npmArgs = append(npmArgs, fmt.Sprintf("--proxy=%s", p.Proxy))
	}

	if p.Verbose {
		npmArgs = append(npmArgs, "--verbose")
	}

	// Reset working dir
	defer func() {
		if p.Verbose {
			utils.Printfln("reset working dir to %s ...", p.workingDir)
		}

		if err := os.Chdir(p.workingDir); err != nil {
			utils.Warnfln("failed to reset working dir: %v", err)
		}
	}()

	if p.Verbose {
		utils.Printfln("changing working dir to %s ...", p.tempDir)
	}

	// We need to change current working dir to temp dir to make bundle-deps work correctly,
	// which needs to be run at package.json level.
	if err := os.Chdir(p.tempDir); err != nil {
		return err
	}

	err := p.execute(exec.Command("npm", npmArgs...))
	if err != nil {
		return err
	}
	return nil
}

// moveLibraryModule moves the library from node_modules to the temp directory (one level up) and removes the original one.
func (p *NodePackager) moveLibraryModule() error {
	p.spinner.Describe("preparing library ...")

	subModulePath := path.Join(p.tempDir, "node_modules", p.libraryNameWithoutVersion)
	entries, err := os.ReadDir(subModulePath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		src := path.Join(subModulePath, entry.Name())
		dst := path.Join(p.tempDir, entry.Name())

		// We Overwrite all files, but keep existing 'package.json'
		if entry.Name() != "package.json" {
			if utils.FileExists(dst) {
				continue
			}
		}

		if p.Verbose {
			utils.Printfln("moving %s -> %s ...", src, dst)
		}

		// Rename (Move)
		if err := os.Rename(src, dst); err != nil {
			return err
		}
	}

	return p.removeDir(subModulePath)
}

// scanNativeModules scans the library for components having native bindings in 'node_modules' recursively.
// - It detects native components.
// - It modifies the 'package.json' of all native components to prevent a native binding rebuild on library import executing 'npm install'.
func (p *NodePackager) scanNativeModules() error {

	p.spinner.Describe("scanning native components ...")

	err := filepath.WalkDir(p.tempDir,
		func(path string, info fs.DirEntry, err error) error {

			if !info.IsDir() {
				return nil
			}

			// skip: blacklist
			if info.Name() == "docs" ||
				info.Name() == "test" ||
				info.Name() == "examples" {
				return filepath.SkipDir
			}

			if info.Name() == "prebuilds" {
				p.hasNativePrebuilds = true // prebuilds detected
				if err := p.checkNativePrebuilds(path); err != nil {
					return err
				}
			}

			return p.removeNativeBindingRebuild(path)
		})

	return err
}

// checkNativePrebuilds checks if all native modules come with matching prebuilds for 'linux_arm64' and 'linux_amd64'.
func (p *NodePackager) checkNativePrebuilds(prebuildsDir string) error {

	// Read prebuilds sub directories, named by Node.js 'os_arch' pattern
	prebuilds, err := os.ReadDir(prebuildsDir)
	if err != nil {
		return err
	}

	m := make(map[string]bool)
	for _, prebuild := range prebuilds {
		m[prebuild.Name()] = true
	}

	// Check for linux_arm64 prebuilds
	linuxArm64Prebuild := m[prebuildLinuxArm64]
	if !linuxArm64Prebuild {
		p.hasAllNativePrebuildsLinuxArm64 = false
	}

	// Check for linux_amd64 prebuilds
	linuxAmd64Prebuild := m[prebuildLinuxAmd64]
	if !linuxAmd64Prebuild {
		p.hasAllNativePrebuildsLinuxAmd64 = false
	}

	return nil
}

// validateNativeModules checks if native modules detected and all requirments met to potentially run on the target.
func (p *NodePackager) validateNativeModules() error {

	// Only if native modules detected
	if !p.hasNativeModules {
		p.hasAllNativePrebuildsLinuxArm64 = false
		p.hasAllNativePrebuildsLinuxAmd64 = false
		return nil
	}

	p.spinner.Describe("validating native modules ...")

	if p.Verbose {
		utils.Printfln("Native modules: %v", p.hasNativeModules)
		utils.Printfln("Native prebuilds: %v", p.hasNativePrebuilds)
		utils.Printfln("All prebuilds available ('linux_arm64'): %v", p.hasAllNativePrebuildsLinuxArm64)
		utils.Printfln("All prebuilds available ('linux_amd64'): %v", p.hasAllNativePrebuildsLinuxAmd64)
	}

	if p.hasNativePrebuilds {
		// If not all native prebuild is shipped for a supported OS target, raise warning.
		if !p.hasAllNativePrebuildsLinuxArm64 {
			utils.Warnfln("this native library contains not all prebuilds for 'linux_arm64' and therefore must be packed on a OS with same architecture")
		}
		if !p.hasAllNativePrebuildsLinuxAmd64 {
			utils.Warnfln("this native library contains not all prebuilds for 'linux_amd64' and therefore must be packed on a OS with same architecture")
		}
	} else {
		// The library can only be installed on targets with matching OS and architecture.
		utils.Warnfln("this native library was compiled for '%v' and therefore can only be installed on same architecture", p.buildOn)
	}

	return nil
}

// modifyPackageJSON modifies the 'package.json'.
func (p *NodePackager) modifyPackageJSON() error {

	p.spinner.Describe("modifying ...")

	packageJSONPath := path.Join(p.tempDir, "package.json")

	// Read 'package.json' into a map
	var jsonMap map[string]interface{}
	bytes, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(bytes, &jsonMap); err != nil {
		return err
	}

	// Remove any optional 'files' element, which defines partial content of the node,
	// because need to pack everything (including node_modules folder and our scripts).
	delete(jsonMap, "files")

	// Ensure 'version', because mandatory
	if _, exist := jsonMap["version"]; !exist {
		jsonMap["version"] = "0.0.0"
	}

	// Ensure 'name', because mandatory
	if _, exist := jsonMap["name"]; !exist {
		jsonMap["name"] = p.libraryNameWithoutVersion
	}

	// Write modified map back to 'package.json'
	buffer, err := json.MarshalIndent(jsonMap, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(packageJSONPath, buffer, utils.DefaultFilePermissionsFiles)
}

// removeNativeBindingRebuild removes native binding rebuild commands in the 'package.json' in given package path.
func (p *NodePackager) removeNativeBindingRebuild(dir string) error {

	// Delete an existing 'binding.gyp' file
	bindingsGypPath := path.Join(dir, "binding.gyp")
	if utils.FileExists(bindingsGypPath) {
		p.hasNativeModules = true // Native module detected
		if err := os.Remove(bindingsGypPath); err != nil {
			return err
		}
	}

	// Remove node-gyp related scripts from an existing 'package.json'
	packageJSONPath := path.Join(dir, "package.json")
	if !utils.FileExists(packageJSONPath) {
		return nil
	}

	// Read 'package.json' into a map
	bytes, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return err
	}

	var jsonMap map[string]interface{}

	// Unmarshal to map
	if err := json.Unmarshal(bytes, &jsonMap); err != nil {
		return err
	}

	// Check for scripts
	scripts, ok := jsonMap["scripts"].(map[string]interface{})
	if !ok {
		return nil
	}

	// Remove scripts which contains a node-gyp rebuild
	modified := false
	scriptNames := []string{
		"prebuild",
		"rebuild",
		"preinstall",
		"postinstall",
		"install",
	}

	for _, scriptName := range scriptNames {
		cmd, hasScript := scripts[scriptName]
		if !hasScript {
			continue
		}
		if !strings.Contains(cmd.(string), "node-gyp rebuild") {
			continue
		}

		p.hasNativeModules = true // Native module detected
		delete(scripts, scriptName)
		modified = true
	}

	// Check if anything has been modified
	if !modified {
		return nil
	}

	// Marshal to buffer
	buffer, err := json.MarshalIndent(jsonMap, "", "    ")
	if err != nil {
		return err
	}

	// Write modified map to 'package.json'
	return os.WriteFile(packageJSONPath, buffer, utils.DefaultFilePermissionsFiles)
}

// findTarball searches for the packed tarball.
func (p *NodePackager) findTarball() (string, string, error) {

	tarballPath := ""
	err := filepath.Walk(p.tempDir, func(path string, fi os.FileInfo, err error) error {

		//We expect the generated tarball to be placed at top level (temp dir)
		if fi.IsDir() && (path != p.tempDir) {
			return filepath.SkipDir
		}

		if strings.HasPrefix(fi.Name(), p.tarballPrefix) && filepath.Ext(fi.Name()) == ".tgz" {
			tarballPath = path
			return filepath.SkipAll
		}

		return nil
	})

	if err != nil {
		return "", "", err
	}

	// tarball not found
	if len(tarballPath) == 0 {
		return "", "", errors.New("tarball not found")
	}

	// tarball found
	_, tarballName := filepath.Split(tarballPath)
	return tarballPath, tarballName, nil
}

// movePackage moves the tarball to the working directory.
func (p *NodePackager) movePackage() error {

	// Find the tarball
	tarballPath, tarballName, err := p.findTarball()
	if err != nil {
		return err
	}

	// Move the tarball to working dir
	if p.Verbose {
		utils.Printfln("moving %s -> %s ...", tarballPath, p.workingDir)
	}
	to := path.Join(p.workingDir, tarballName)
	err = os.Rename(tarballPath, to)
	if err != nil {
		return err
	}

	p.tarballPath = to
	p.tarballName = tarballName
	p.tarballSize, err = utils.HumanizedFileSize(p.tarballPath)
	return err
}

// getWorkingDir returns the working dir.
func (p *NodePackager) getWorkingDir() error {
	workingDir, err := os.Getwd()
	if err != nil {
		return err
	}

	p.workingDir = workingDir
	if p.Verbose {
		utils.Printfln("working directory: %s", p.workingDir)
	}
	return nil
}

// createTempDir creates and returns the temp dir.
func (p *NodePackager) createTempDir() error {

	// Create temp dir
	tempDir, err := os.MkdirTemp(p.workingDir, p.libraryNameToFolderName(p.libraryNameWithoutVersion)+"-")
	if err != nil {
		return err
	}

	if p.Verbose {
		utils.Printfln("temp directory: %s", tempDir)
	}

	p.tempDir = tempDir
	return nil
}

// removeDir removes the given directory.
func (p *NodePackager) removeDir(dir string) error {

	if p.Verbose {
		utils.Printfln("removing %s ...", dir)
	}

	return os.RemoveAll(dir)
}

// execute excecutes a command and forwards the console output.
func (p *NodePackager) execute(cmd *exec.Cmd) error {

	if p.Verbose {
		utils.Printfln("executing %s %v ...", cmd.Path, strings.Join(cmd.Args, " "))
		// redirect the command output to the console
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

// libraryNameToFolderName returns the library name as a folder name by replacing all file path chars ('/', ':') and '@' with underscores '_'.
func (p *NodePackager) libraryNameToFolderName(name string) string {
	folder := filepath.ToSlash(name)
	folder = strings.ReplaceAll(folder, "@", "_")
	folder = strings.ReplaceAll(folder, string(os.PathSeparator), "_")
	folder = strings.ReplaceAll(folder, string(os.PathListSeparator), "_")
	return folder
}

// hasSrcDir returns a value, indicating whether a a source directory is given.
func (p *NodePackager) hasSrcDir() bool {
	return len(p.SrcDir) > 0
}
