/*
 * SPDX-FileCopyrightText: Bosch Rexroth AG
 *
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */
package packager

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"lo-stash.de.bosch.com/icx/tool.node-packager.git/pkg/utils"
)

const (
	TestLibDataView            = "node-red-contrib-data-view"
	TestLibOpcUa               = "node-red-contrib-opcua"
	TestLibNativeWithPrebuilds = "node-red-contrib-usb" //"node-red-contrib-modbus"

	TestLibVersioned       = TestLibDataView + "@0.0.3"
	TestLibScoped          = "@flowfuse/node-red-dashboard"
	TestLibScopedVersioned = TestLibScoped + "@1.16.0"
)

type NodePackagerTestSuite struct {
	p NodePackager

	suite.Suite
}

func TestNodePackagerTestSuite(t *testing.T) {
	// Create a new suite for tests
	//s = new(NodePackagerTestSuite)
	suite.Run(t, new(NodePackagerTestSuite))
}

// SetupSuite will run before the tests in the suite are run.
func (s *NodePackagerTestSuite) SetupSuite() {
	beforeAll(s)
}

// TearDownSuite will run after all the tests in the suite have been run.
func (s *NodePackagerTestSuite) TearDownSuite() {
	// We have to use afterAll instead in TestMain() to ensure proper shutdown for benchmarks and fuzzing, too.
	afterAll(s)
}

// SetupTest has a SetupTest method, which will run before each test in the suite.
func (s *NodePackagerTestSuite) SetupTest() {
	// Not used
}

// BeforeTest has a function to be executed right before the test starts and receives the suite and test names as input
func (s *NodePackagerTestSuite) BeforeTest(suiteName, testName string) {
	// Not used
}

// AfterTest has a function to be executed right after the test finishes and receives the suite and test names as input
func (s *NodePackagerTestSuite) AfterTest(suiteName, testName string) {

	// Clean up
	if utils.FileExists(s.p.tarballName) {
		s.NoError(os.Remove(s.p.tarballPath))
	}

	if utils.FileExists(s.p.tempDir) {
		s.NoError(os.RemoveAll(s.p.tempDir))
	}
}

// beforeAll ensures the test setup and starts the app which is subject of test (beforeAll).
func beforeAll(s *NodePackagerTestSuite) {
	// Not used
}

// afterAll runs after all tests executed.
func afterAll(s *NodePackagerTestSuite) {
	// Not used
}

func (s *NodePackagerTestSuite) TestPackDataView() {

	// Arrange
	s.p = NodePackager{
		LibraryName: TestLibDataView,
	}

	// Act
	tarball, err := s.p.Pack()

	// Assert
	s.NoError(err)
	s.NotEmpty(tarball)
	s.FileExists(tarball)
	s.Equal(".tgz", filepath.Ext(tarball))
	s.NoDirExists(s.p.tempDir)

	s.False(s.p.hasNativeModules)
	s.False(s.p.hasNativePrebuilds)
	s.False(s.p.hasAllNativePrebuildsLinuxAmd64)
	s.False(s.p.hasAllNativePrebuildsLinuxArm64)
}

func (s *NodePackagerTestSuite) TestPackNativeWithPrebuilds() {

	// Arrange
	s.p = NodePackager{
		LibraryName: TestLibNativeWithPrebuilds,
		NoAudit:     true,
	}

	// Act
	tarball, err := s.p.Pack()

	// Assert
	s.NoError(err)
	s.NotEmpty(tarball)
	s.FileExists(tarball)
	s.Equal(".tgz", filepath.Ext(tarball))
	s.NoDirExists(s.p.tempDir)

	s.True(s.p.hasNativeModules)
	s.True(s.p.hasNativePrebuilds)
	s.True(s.p.hasAllNativePrebuildsLinuxAmd64)
	s.True(s.p.hasAllNativePrebuildsLinuxArm64)
}

func (s *NodePackagerTestSuite) TestPackVersioned() {

	// Arrange
	s.p = NodePackager{
		LibraryName: TestLibVersioned,
		NoAudit:     true,
	}

	// Act
	tarball, err := s.p.Pack()

	// Assert
	s.NoError(err)
	s.NotEmpty(tarball)
	s.FileExists(tarball)
	s.Equal(".tgz", filepath.Ext(tarball))
	s.NoDirExists(s.p.tempDir)

	s.False(s.p.hasNativeModules)
	s.False(s.p.hasNativePrebuilds)
	s.False(s.p.hasAllNativePrebuildsLinuxAmd64)
	s.False(s.p.hasAllNativePrebuildsLinuxArm64)
}

func (s *NodePackagerTestSuite) TestPackScoped() {

	// Arrange
	s.p = NodePackager{
		LibraryName: TestLibScoped,
		NoAudit:     true,
	}

	// Act
	tarball, err := s.p.Pack()

	// Assert
	s.NoError(err)
	s.NotEmpty(tarball)
	s.FileExists(tarball)
	s.Equal(".tgz", filepath.Ext(tarball))
	s.NoDirExists(s.p.tempDir)

	s.False(s.p.hasNativeModules)
	s.False(s.p.hasNativePrebuilds)
	s.False(s.p.hasAllNativePrebuildsLinuxAmd64)
	s.False(s.p.hasAllNativePrebuildsLinuxArm64)
}

func (s *NodePackagerTestSuite) TestPackScopedVersioned() {

	// Arrange
	s.p = NodePackager{
		LibraryName: TestLibScopedVersioned,
		NoAudit:     true,
	}

	// Act
	tarball, err := s.p.Pack()

	// Assert
	s.NoError(err)
	s.NotEmpty(tarball)
	s.FileExists(tarball)
	s.Equal(".tgz", filepath.Ext(tarball))
	s.NoDirExists(s.p.tempDir)

	s.False(s.p.hasNativeModules)
	s.False(s.p.hasNativePrebuilds)
	s.False(s.p.hasAllNativePrebuildsLinuxAmd64)
	s.False(s.p.hasAllNativePrebuildsLinuxArm64)
}

func (s *NodePackagerTestSuite) TestPackOpcUa() {

	s.T().Skip()

	// Arrange
	s.p = NodePackager{
		LibraryName: TestLibOpcUa,
		Verbose:     true,
	}

	// Act
	tarball, err := s.p.Pack()

	// Assert
	s.NoError(err)
	s.NotEmpty(tarball)
	s.FileExists(tarball)
	s.Equal(".tgz", filepath.Ext(tarball))
	s.NoDirExists(s.p.tempDir)

	s.False(s.p.hasNativeModules)
	s.False(s.p.hasNativePrebuilds)
	s.False(s.p.hasAllNativePrebuildsLinuxAmd64)
	s.False(s.p.hasAllNativePrebuildsLinuxArm64)
}

func (s *NodePackagerTestSuite) TestPackKeepTmp() {

	// Arrange
	s.p = NodePackager{
		LibraryName: TestLibDataView,
		KeepTmp:     true,
	}

	// Act
	tarball, err := s.p.Pack()

	// Assert
	s.NoError(err)
	s.NotEmpty(tarball)
	s.FileExists(tarball)
	s.Equal(".tgz", filepath.Ext(tarball))
	s.DirExists(s.p.tempDir)
}

func (s *NodePackagerTestSuite) TestPackAuditFix() {

	// Arrange
	s.p = NodePackager{
		LibraryName: TestLibDataView,
		AuditFix:    true,
	}

	// Act
	tarball, err := s.p.Pack()

	// Assert
	s.NoError(err)
	s.NotEmpty(tarball)
	s.FileExists(tarball)
	s.Equal(".tgz", filepath.Ext(tarball))
	s.NoDirExists(s.p.tempDir)
}

func (s *NodePackagerTestSuite) TestPackNoAudit() {

	// Arrange
	s.p = NodePackager{
		LibraryName: TestLibDataView,
		NoAudit:     true,
	}

	// Act
	tarball, err := s.p.Pack()

	// Assert
	s.NoError(err)
	s.NotEmpty(tarball)
	s.FileExists(tarball)
	s.Equal(".tgz", filepath.Ext(tarball))
	s.NoDirExists(s.p.tempDir)
}

func (s *NodePackagerTestSuite) TestPackVerbose() {

	// Arrange
	s.p = NodePackager{
		LibraryName: TestLibDataView,
		Verbose:     true,
	}

	// Act
	tarball, err := s.p.Pack()

	// Assert
	s.NoError(err)
	s.NotEmpty(tarball)
	s.FileExists(tarball)
	s.Equal(".tgz", filepath.Ext(tarball))
	s.NoDirExists(s.p.tempDir)
}

func (s *NodePackagerTestSuite) TestPackSrcDir() {

	// Arrange

	// Install a node module as source to temp directory
	tmpDir := s.T().TempDir()
	srcDir := path.Join(tmpDir, TestLibDataView)

	args := []string{
		"install",
		TestLibDataView,
		fmt.Sprintf("--prefix=%s", srcDir),
		"--no-fund",
		"--omit=dev",
		"--no-audit"}

	s.NoError(s.p.execute(exec.Command("npm", args...))) //nolint:gosec

	s.p = NodePackager{
		LibraryName: TestLibDataView,
		SrcDir:      srcDir,
	}

	// Act
	tarball, err := s.p.Pack()

	// Assert
	s.NoError(err)
	s.NotEmpty(tarball)
	s.FileExists(tarball)
	s.Equal(".tgz", filepath.Ext(tarball))
	s.NoDirExists(s.p.tempDir)
}

func (s *NodePackagerTestSuite) TestPackNameMissing() {

	// Arrange
	s.p = NodePackager{}

	// Act
	tarball, err := s.p.Pack()

	// Assert
	s.Error(err)
	s.Empty(tarball)
	s.NoFileExists(tarball)
}

func (s *NodePackagerTestSuite) TestInvalidSrcDir() {

	// Arrange
	s.p = NodePackager{
		LibraryName: TestLibDataView,
		SrcDir:      path.Join(s.T().TempDir(), "42"),
	}

	// Act
	tarball, err := s.p.Pack()

	// Assert
	s.Error(err)
	s.Empty(tarball)
	s.NoFileExists(tarball)
}
