package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Sn0wo2/CatSync/internal/util"
)

func runCmd(command string, args ...string) (string, error) {
	out, err := exec.Command(command, args...).CombinedOutput()
	s := strings.TrimSpace(util.BytesToString(out))
	if err != nil {
		if strings.Contains(s, "No names found") {
			return "", nil
		}
		return "", fmt.Errorf("failed to run command '%s %s': %w\n%s", command, strings.Join(args, " "), err, s)
	}
	return s, nil
}

func must(out string, err error) string {
	if err != nil {
		panic(err)
	}
	return out
}

func executeStep(description string, command string, args ...string) {
	fmt.Println(description)
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func main() {
	must(runCmd("git", "fetch", "origin"))

	status := must(runCmd("git", "status", "--porcelain"))
	if status != "" {
		panic("Uncommitted changes found, please commit or stash them first.")
	}

	local := must(runCmd("git", "rev-parse", "@"))
	remote := must(runCmd("git", "rev-parse", "@{u}"))

	if local != remote {
		fmt.Println("Local branch is not up to date with remote, pulling...")
		must(runCmd("git", "pull"))
	}

	lastTag, err := runCmd("git", "describe", "--tags", "--abbrev=0")
	if err != nil {
		if lastTag == "" || strings.Contains(lastTag, "No names found") {
			lastTag = ""
		} else {
			panic(err)
		}
	}

	if lastTag == "" {
		fmt.Println("No tags found.")
	} else {
		fmt.Printf("Latest tag: %s\n", lastTag)
	}

	fmt.Print("Enter new tag: ")
	newTag, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		panic(err)
	}
	newTag = strings.TrimSpace(newTag)
	if newTag == "" {
		panic("No tag entered, aborting.")
	}

	executeStep(fmt.Sprintf("Tagging %s...", newTag), "git", "tag", newTag)
	executeStep(fmt.Sprintf("Pushing tag %s...", newTag), "git", "push", "origin", newTag)

	fmt.Printf("Successfully tagged and pushed %s.\n", newTag)
}
