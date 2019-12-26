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
	version           = "0.3.0"
	packageName       = "github.com/gusmin/gate"
	golintPackage     = "golang.org/x/lint/golint"
	securegateKeysDir = ".sgsh"
	releaseDir        = "scripts/release"
	releaseBinDir     = "scripts/release/bin"
	releaseBinName    = "securegate-gate"
	configFile        = "config.json"
	configDir         = "/etc/securegate/gate"
	logDir            = "/var/log/securegate/gate"
	secureGateLibDir  = "/var/lib/securegate/gate"
	translations      = "./translations/"
	translationsDir   = "/var/lib/securegate/gate/translations"
	translationsTar   = "translations.tgz"
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

// Release Type to create release namespace
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

// Linux Create release tarbal with linux binary
func (Release) Linux() error {
	fmt.Println("[>] Creating Linux release")
	fmt.Println("[>] Setting up environment")
	var err error
	err = os.Setenv("GOOS", "linux")
	if err != nil {
		log.Fatal(err)
	}
	err = os.Setenv("GOARCH", "amd64")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("[>] Building")
	err = goBuild("-o", releaseBinDir+"/"+releaseBinName)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("[>] Creating translations archive")
	out, err := exec.Command("/usr/bin/tar", "-zcf", releaseDir+"/"+translationsTar, translations).Output()
	if len(out) != 0 {
		fmt.Println(out)
	}
	if err != nil {
		log.Fatal(err)
	}
	err = os.Chdir(releaseDir)
	if err != nil {
		return err
	}
	fmt.Println("[>] Creating release version " + version)
	err = sh.Run("./release.sh", version)
	if err != nil {
		return err
	}
	fmt.Println("[+] Release available in directory ./scripts/release/releases/")
	return nil
}

// Install A custom install step if you need your bin someplace other than go/bin
func Install() error {
	fmt.Println("[>] Installing")

	makeSGSHDir()
	makeLogDir()
	installTranslations()
	initConfiguration()

	return goInstall(packageName)
}

// Vet Run go vet linter
func Vet() error {
	fmt.Println("[>] Code analysis")
	return goVet("./...")
}

// Golint Run golint linter
func Golint() error {
	golintInstall()
	fmt.Println("[>] Linting")
	return goLint("./...")
}

// Lint Run all linters in parallel (e.g. go vet, golint...)
func Lint() {
	mg.Deps(Vet, Golint)
}

// Test Run tests
func Test() error {
	fmt.Println("[>] Testing")
	return goTest("-v", "-race", "-coverprofile=cover.out", "./...")
}

// Check Run tests and linter
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
	fmt.Println("[>] Checking for existing linter")
	if err := goList(golintPackage); err != nil {
		fmt.Println("[-] Linter does not exists")
		fmt.Println("[+] Installing golint linter")
		if err := goGet("-u", golintPackage); err != nil {
			log.Fatal(err)
		}
	}
}

func initConfiguration() {
	fmt.Println("[>] Checking for existing configuration file config.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Println("[x] You need to copy config.json.template to config.json and " +
			"complete the configuration before launching the installation")
		log.Fatal(err)
	} else {
		fmt.Println("[>] Configuration file already exists")
	}
	fmt.Println("[>] Checking for existing configuration directory in /etc/securegate")
	ok, err := exists(configDir)
	if err != nil {
		log.Fatal(err)
	}
	if !ok {
		fmt.Println("[+] Creating directory for Secure Gate configuration files")
		out, err := exec.Command("/bin/sh", "-c", "sudo mkdir -p "+configDir).Output()
		if len(out) != 0 {
			fmt.Print(out)
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("[+] Configuration directory created")
	} else {
		fmt.Println("[>] Config directory already exists")
	}
	fmt.Println("[>] Initialize basic configuration for Secure Gate")
	out, err := exec.Command("/bin/sh", "-c", "sudo cp "+configFile+" "+configDir+"/"+configFile).Output()
	if len(out) != 0 {
		fmt.Println(out)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("[+] Basic configuration created")
}

func makeSGSHDir() {
	home := os.Getenv("HOME")
	finalPath := path.Join(home, securegateKeysDir)
	fmt.Println("[>] Checking for existing secret keys directory in $HOME")
	ok, err := exists(finalPath)
	if err != nil {
		log.Fatal(err)
	}
	if !ok {
		fmt.Println("[>] Creating new secret keys directory in $HOME")
		err := os.Mkdir(finalPath, 0777)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("[+] Secret keys directory created")
	} else {
		fmt.Println("[>] Secret keys directory already exists")
	}
}

func makeLogDir() {
	user := os.Getenv("USER")
	fmt.Println("[>] Checking for existing log directory")
	ok, err := exists(logDir)
	if err != nil {
		log.Fatal(err)
	}
	if !ok {
		fmt.Println("[>] Creating log directory with correct access right")
		out, err := exec.Command("/bin/sh", "-c", "sudo mkdir -p "+logDir).Output()
		if len(out) != 0 {
			fmt.Print(out)
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("[+] Directory created")
	} else {
		fmt.Println("[>] Directory already exists")
	}
	fmt.Println("[>] Update access right of the directory to current user")
	out, err := exec.Command("/bin/sh", "-c", "sudo chown -R "+user+": "+logDir).Output()
	if len(out) != 0 {
		fmt.Print(out)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func installTranslations() {
	user := os.Getenv("USER")
	fmt.Println("[>] Installing translations")
	fmt.Println("[>] Checking for existing translations directory")
	ok, err := exists(translationsDir)
	if err != nil {
		log.Fatal(err)
	}
	if !ok {
		fmt.Println("[>] Creating translations directory with correct access right")
		out, err := exec.Command("/bin/sh", "-c", "sudo mkdir -p "+translationsDir).Output()
		if len(out) != 0 {
			fmt.Print(out)
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("[+] Directory created")
	} else {
		fmt.Println("[>] Directory already exists")
	}
	fmt.Println("[>] Update access right of the directory to current user")
	out, err := exec.Command("/bin/sh", "-c", "sudo chown -R "+user+": "+secureGateLibDir).Output()
	if len(out) != 0 {
		fmt.Print(out)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("[+] Directory updated")
	fmt.Println("[>] Update translations")
	out, err = exec.Command("/bin/sh", "-c", "sudo cp -R "+translations+"* "+translationsDir).Output()
	if len(out) != 0 {
		fmt.Print(out)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("[+] Translations updated")
}
