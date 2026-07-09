package main

import (
	"crypto/sha512"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	distribution string
	architecture string
	role         string
	force        bool
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	operation := os.Args[1]
	switch operation {
	case "version", "-v", "--version":
		runVersion()
		return
	case "-h", "--help", "help":
		usage()
		return
	}

	if err := checkRequiredCommands(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	switch operation {
	case "create", "update", "login":
		run(operation, os.Args[2:])
	default:
		fmt.Printf("unknown command: %s\n\n", operation)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("chroot-pbuilder simplifies chroot environment management using pbuilder.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  chroot-pbuilder <create|update|login> --distribution <dist> [options] [-- pbuilder options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --distribution  Distribution (required)")
	fmt.Println("  --architecture  Architecture (default: amd64)")
	fmt.Println("  --role          Role")
	fmt.Println("  --force         Overwrite existing base.tgz (create only)")
}

func runVersion() {
	fmt.Printf("chroot-pbuilder %s (commit %s, built %s)\n", version, commit, date)
}

func run(operation string, args []string) {
	fs := flag.NewFlagSet(operation, flag.ExitOnError)
	fs.StringVar(&distribution, "distribution", "", "Distribution (required)")
	fs.StringVar(&architecture, "architecture", "amd64", "Architecture")
	fs.StringVar(&role, "role", "", "Role")
	if operation == "create" {
		fs.BoolVar(&force, "force", false, "Overwrite existing base.tgz")
	}

	fs.Parse(args)

	if distribution == "" {
		fmt.Println("Error: --distribution is required")
		os.Exit(1)
	}

	if operation == "create" {
		runCreate(fs.Args())
		return
	}

	runPbuilder(operation, fs.Args())
}

func checkRequiredCommands() error {
	commands := []string{"sudo", "/usr/sbin/pbuilder"}
	for _, cmd := range commands {
		if _, err := exec.LookPath(cmd); err != nil {
			return fmt.Errorf("required command not found: %s", cmd)
		}
	}
	return nil
}

func runCreate(args []string) {
	baseTgz := getBaseTgzPath(resolveRole())

	if _, err := os.Stat(baseTgz); err == nil {
		if !force {
			fmt.Printf("baseTgz already exists at %s. Use --force to overwrite.\n", baseTgz)
			return
		}
		fmt.Printf("Force flag set. Removing existing baseTgz at %s.\n", baseTgz)
		err = os.Remove(baseTgz)
		if err != nil {
			fmt.Printf("Error removing existing baseTgz: %v\n", err)
			return
		}
	}

	runPbuilder("create", args)
}

func runPbuilder(operation string, args []string) {
	effectiveRole := resolveRole()
	baseTgz := getBaseTgzPath(effectiveRole)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		os.Exit(1)
	}

	bindMountDir := filepath.Join(homeDir, ".chroot-pbuilder", fmt.Sprintf("%s-%s-%s", distribution, architecture, effectiveRole))
	err = os.MkdirAll(bindMountDir, 0755)
	if err != nil {
		fmt.Println("Error creating bind mount directory:", err)
		os.Exit(1)
	}

	pbuilderArgs := []string{
		operation,
		"--basetgz", baseTgz,
		"--distribution", distribution,
		"--architecture", architecture,
		"--bindmounts", bindMountDir,
	}

	pbuilderArgs = append(pbuilderArgs, args...)

	pbuilderCmd := exec.Command("sudo", append([]string{"pbuilder"}, pbuilderArgs...)...)
	pbuilderCmd.Stdout = os.Stdout
	pbuilderCmd.Stderr = os.Stderr
	pbuilderCmd.Stdin = os.Stdin

	err = pbuilderCmd.Run()
	if err != nil {
		fmt.Printf("Error running pbuilder %s: %v\n", operation, err)
		os.Exit(1)
	}
}

func resolveRole() string {
	if role != "" {
		return role
	}

	hash := sha512.Sum512([]byte(distribution + "-" + architecture))
	return fmt.Sprintf("%x", hash)[:10]
}

func getBaseTgzPath(effectiveRole string) string {
	baseDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		os.Exit(1)
	}

	return filepath.Join(baseDir, fmt.Sprintf("%s-%s-%s.tgz", distribution, architecture, effectiveRole))
}
