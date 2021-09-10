package main

import (
	"fmt"
	"github.com/shirou/gopsutil/process"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)


func main() {
	processes, err := process.Processes()

	if err != nil {
		log.Println("Memory unreadable", err)
		os.Exit(1)
	}


	regexList := createRegexList([]string{"DisplayLinkUserAgent", "Wacom.*Driver", ".*Teams.*"})
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(len(regexList) * len(processes))

	for _, singleProcess := range processes {
		for _, regex := range regexList {
			go searchAndDestroy(waitGroup, singleProcess, regex)
		}
	}
	waitGroup.Wait()

	cleanTeamsCache()

	launchApplication("/Applications/DisplayLink Manager.app")
	launchApplication("/Applications/Microsoft Teams.app")
}

func searchAndDestroy(waitGroup *sync.WaitGroup, currentProcess *process.Process, processNameRegex *regexp.Regexp) {
	defer waitGroup.Done()
	processName, err := currentProcess.Name()

	if err != nil {
		log.Println("Process name unreadable", err)
	}

	if processNameRegex.MatchString(processName) {
		log.Printf(`Found '%s' on PID '%d'`, processName, currentProcess.Pid)
		currentProcess.Kill()
	}
}

func launchApplication(applicationPath string) {
	log.Printf("Launching %s", applicationPath)
	command := exec.Command("open", applicationPath)
	_, err := command.Output()
	command.Run()

	if err != nil {
		log.Println("Error executing command", err)
	}
}

func cleanTeamsCache() {
	command := exec.Command("whoami")
	usernameOutput , err := command.Output()
	command.Run()

	if err != nil {
		log.Println("Error executing command", err)
	}

	username := strings.TrimSpace(string(usernameOutput))

	teamsPath := fmt.Sprintf("/Users/%s/Library/Application Support/Microsoft/Teams", username)
	log.Printf("Deleting %s", teamsPath)
	err = os.RemoveAll(teamsPath)
	if err != nil {
		log.Fatal(err)
	}

}

func createRegexList(regexStringList []string) []*regexp.Regexp {
	var regexList []*regexp.Regexp

	for _, regex := range regexStringList {


		compiledRegex, err := regexp.Compile(regex)
		if err != nil {
			log.Println("Invalid Regex", err)
			os.Exit(2)
		}

		regexList = append(regexList, compiledRegex)
	}

	return regexList
}
