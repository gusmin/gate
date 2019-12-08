// +build mage

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/magefile/mage/sh"

	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
)

const (
	version           = "1.0.1-beta"
	packageName       = "github.com/gusmin/gate"
	golintPackage     = "golang.org/x/lint/golint"
	securegateKeysDir = ".sgsh"
	releaseDir        = "scripts/release"
	releaseBinDir     = "scripts/release/bin"
	releaseBinName    = "securegate-gate"
	configFile        = "config.json"
	configDir         = "/etc/securegate/gate"
)

var (
	goCmd     = mg.GoCmd()
	goGet     = sh.RunCmd(goCmd, "get")
	goList    = sh.RunCmd(goCmd, "list")
	goBuild   = sh.RunCmd(goCmd, "build")
	goInstall = sh.RunCmd(goCmd, "install")
	goTest    = sh.RunCmd(goCmd, "test")
	goVet     = sh.RunCmd(goCmd, "vet")
	goLint    = sh.RunCmd("golint")
)

type Release mg.Namespace

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build
func init() {
	// We want to use Go 1.11 modules even if the source lives inside GOPATH.
	// The default is "auto".
	err := os.Setenv("GO111MODULE", "on")
	if err != nil {
		log.Fatal(err)
	}
}

// Create release tarbal with linux binary
func (Release) Linux() error {
	fmt.Println("Building...")
	env := map[string]string{
		"GOOS":   "linux",
		"GOARCH": "amd64",
	}
	err := sh.RunWith(env, goCmd, "build", "-o", releaseBinDir+"/"+releaseBinName)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = os.Chdir(releaseDir)
	if err != nil {
		return err
	}
	fmt.Println("Creating release...")
	return sh.Run("./release.sh", version)
}

// A custom install step if you need your bin someplace other than go/bin
func Install() error {
	fmt.Println("Installing...")

	makeSGSHDir()
	initConfig()
	golintInstall()

	return goInstall(packageName)
}

// Run go vet linter
func Vet() error {
	return goVet("./...")
}

// Run golint linter
func Golint() error {
	return goLint("./...")
}

// Run all linters in parallel (e.g. go vet, golint...)
func Lint() {
	fmt.Println("Linting...")
	mg.Deps(Vet, Golint)
}

// Run tests
func Test() error {
	fmt.Println("Testing...")
	return goTest("-v", "./...")
}

// Run tests and linter
func Check() {
	mg.Deps(Lint)
	mg.Deps(Test)
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func golintInstall() {
	fmt.Println("Checking for existing linter...")
	if err := goList(golintPackage); err != nil {
		fmt.Println("Linter does not exists")
		fmt.Println("Installing golint linter...")
		if err := goGet("-u", golintPackage); err != nil {
			log.Fatal(err)
		}
	}
}

func initConfig() {
	fmt.Println("Checking for existing config file config.json...")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Println("You need to copy config.json.templatem to config.json and " +
			"complete the configuration before launching the installation")
		log.Fatal(err)
	} else {
		fmt.Println("Exists.")
	}
	fmt.Println("Checking for existing config directory in /etc/...")
	ok, err := exists(configDir)
	if err != nil {
		log.Fatal(err)
	}
	if !ok {
		fmt.Println("Making directory for Secure Gate config files...")
		out, err := exec.Command("/bin/sh", "-c", "sudo mkdir -p "+configDir).Output()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(out)
	} else {
		fmt.Println("Already exists.")
	}
	fmt.Println("Initialize basic configuration for Secure Gate...")
	err = sh.Run("/bin/sh", "-c", "sudo cp "+configFile+" "+configDir+"/"+configFile)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Done.")
}

func makeSGSHDir() {
	home := os.Getenv("HOME")
	finalPath := path.Join(home, securegateKeysDir)
	fmt.Println("Checking for existing secret keys directory in $HOME...")
	ok, err := exists(finalPath)
	if err != nil {
		log.Fatal(err)
	}
	if !ok {
		fmt.Println("Making new secret keys directory in $HOME...")
		err := os.Mkdir(finalPath, 0777)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("Already exists.")
	}
}
