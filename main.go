package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"

	"github.com/go-git/go-git/v5"
	"github.com/manifoldco/promptui"
	"github.com/pterm/pterm"
	"golang.org/x/sys/windows/registry"
)

func check_teardown(steamPath string) bool {
	teardowmPath := steamPath + "\\steamapps\\common\\Teardown"
	if _, err := os.Stat(teardowmPath); os.IsNotExist(err) {
		prompt := promptui.Select{
			Label: "Is it in a different path?",
			Items: []string{"Yes", "No"},
		}

		_, result, _ := prompt.Run()

		if result == "Yes" {
			prompt := promptui.Prompt{
				Label: "Teardown root directory path",
			}

			_, err := prompt.Run()
			if err == nil {
				return true
			}
		} else {
			return false
		}
	} else {
		return true
	}

	// Default to false if something goes wrong
	return false
}

func sledge_clone() {
	spinner, _ := pterm.DefaultSpinner.Start("Checking if Seldge folder exists")
	if _, err := os.Stat("sledge"); os.IsNotExist(err) {
		spinner.UpdateText("Sledge folder not found")
		os.Mkdir("sledge", 0755)
	}

	if _, err := os.Stat("sledge\\sledge1"); os.IsNotExist(err) {
		spinner.UpdateText("Creating sledge1 folder")
		os.Mkdir("sledge\\sledge1", 0755)
		spinner.UpdateText("Downloading Sledge")
		git.PlainClone(".\\sledge\\sledge1", false, &git.CloneOptions{
			URL:               "https://github.com/44lr/sledge.git",
			RecurseSubmodules: 10,
		})
	}

	if _, err := os.Stat("sledge\\sledge2"); os.IsNotExist(err) {
		spinner.UpdateText("Creating sledge2 folder")
		os.Mkdir("sledge\\sledge2", 0755)
		git.PlainClone(".\\sledge\\sledge2", false, &git.CloneOptions{
			URL:               "https://github.com/44lr/sledge.git",
			RecurseSubmodules: 10,
		})
	}

	spinner.Success("Sledge cloned")
}

func check_prerequisites(prerequisites [5]string) {
	p, _ := pterm.DefaultProgressbar.WithTotal(len(prerequisites)).WithTitle("Checking prerequisites").Start()

	var steamPath string

	for i := 0; i < p.Total; i++ {
		p.UpdateTitle("Checking " + prerequisites[i])

		switch cmd := prerequisites[i]; cmd {
		case "cmake":
			cmakeCheck := exec.Command(`cmd.exe`, `/C`, `cmake --version`)
			if err := cmakeCheck.Run(); err != nil {
				pterm.Error.Println("cmake could not be found.\nInstall cmake-3.23.0-windows-x86_64.msi from https://cmake.org/download/.\nMake sure you select to be included in your PATH!")
			} else {
				pterm.Success.Println("cmake found!")
				p.Increment()
			}
		case "openSSL":
			k, openSSLErr := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\OpenSSL (64-bit)_is1`, registry.QUERY_VALUE)
			if openSSLErr != nil {
				pterm.Error.Println("OpenSSL could not be found. Install Win64 OpenSSL from https://slproweb.com/products/Win32OpenSSL.html.\nMake sure you install the full version (not the Light version).")
			} else {
				pterm.Success.Println("OpenSSL found!")
				p.Increment()
			}
			defer k.Close()
		case "dotnet":
			dotnetCheck := exec.Command(`cmd.exe`, `/C`, `dotnet --version`)
			if err := dotnetCheck.Run(); err != nil {
				pterm.Error.Println(".NET 6.0 SDK could not be found.\nInstall .NET 6.0 SDK from https://dotnet.microsoft.com/en-us/download/dotnet/6.0")
			} else {
				pterm.Success.Println(".NET 6.0 SDK found!")
				p.Increment()
			}
		case "steam":
			k, _ := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Valve\Steam`, registry.QUERY_VALUE)
			steamInstallPath, _, err := k.GetStringValue("InstallPath")
			steamPath = steamInstallPath
			if err != nil {
				pterm.Error.Println("Steam could not be found, please ensure it is installed.")
			} else {
				pterm.Success.Println("Steam found!")
				p.Increment()
			}
		case "teardown":
			if check_teardown(steamPath) {
				pterm.Success.Println("Teardown found!")
				p.Increment()
			} else {
				pterm.Error.Println("Teardown could not be found.")
			}
		}
	}

	p.Stop()

	if p.Current < 5 {
		pterm.Println()
		pterm.Error.Println("Please ensure you have all the pre-requisites before continuing")
		os.Exit(1)
	}
}

func sledge_build(sledgeName string) {
	cmd := exec.Command(`.\Create project.bat`)
	cmd.Dir = "C:\\Users\\alexandargyurov\\Desktop\\teardownM-cli\\sledge\\" + sledgeName

	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)

	cmd.Stdout = mw
	cmd.Stderr = mw

	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func teardownM_clone() {
	folderName := "teardownM-client"
	spinner, _ := pterm.DefaultSpinner.Start("Checking if " + folderName + " folder exists")
	if _, err := os.Stat(folderName); os.IsNotExist(err) {
		spinner.UpdateText(folderName + " folder not found")
		os.Mkdir(folderName, 0755)
		spinner.UpdateText("Downloading " + folderName)

	}
}

func main() {
	pterm.EnableDebugMessages()

	pterm.DefaultBigText.WithLetters(
		pterm.NewLettersFromStringWithStyle("teardown", pterm.NewStyle(pterm.FgCyan)),
		pterm.NewLettersFromStringWithStyle("M", pterm.NewStyle(pterm.FgLightMagenta))).
		Render()

	pterm.DefaultSection.Println("Checking prerequisites")

	prerequisites := [5]string{"cmake", "openSSL", "dotnet", "steam", "teardown"}
	check_prerequisites(prerequisites)

	pterm.DefaultSection.Println("Sledge 1")
	sledge_clone()
	pterm.Warning.Println("Currently cannot detect if Sledge build script has passed or failed. Please see console output.")
	sledge_build("sledge1")

	pterm.DefaultSection.Println("Sledge 2")
	sledge_build("sledge2")

	pterm.DefaultSection.Println("TeardownM")
	teardownM_clone()
}
