package main

import (
	"crypto/sha512"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	distribution string
	architecture string
	role         string
	force        bool
)

func main() {
	if err := checkRequiredCommands(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var rootCmd = &cobra.Command{
		Use:   "chroot-pbuilder",
		Short: "A tool to simplify chroot environment creation and management using pbuilder",
	}

	var createCmd = &cobra.Command{
		Use:   "create [additional pbuilder options]",
		Short: "Create a new chroot environment",
		Run:   runCreate,
	}

	var updateCmd = &cobra.Command{
		Use:   "update [additional pbuilder options]",
		Short: "Update an existing chroot environment",
		Run:   runUpdate,
	}

	var loginCmd = &cobra.Command{
		Use:   "login [additional pbuilder options]",
		Short: "Log in to a chroot environment",
		Run:   runLogin,
	}

	rootCmd.PersistentFlags().StringVarP(&distribution, "distribution", "d", "", "Distribution (required)")
	rootCmd.PersistentFlags().StringVarP(&architecture, "architecture", "a", "amd64", "Architecture")
	rootCmd.PersistentFlags().StringVarP(&role, "role", "r", "", "Role")
	createCmd.Flags().BoolVarP(&force, "force", "f", false, "Force creation even if baseTgz exists")

	rootCmd.MarkPersistentFlagRequired("distribution")

	rootCmd.AddCommand(createCmd, updateCmd, loginCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
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

func runCreate(cmd *cobra.Command, args []string) {
	baseTgz := getBaseTgzPath()

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

func runUpdate(cmd *cobra.Command, args []string) {
	runPbuilder("update", args)
}

func runLogin(cmd *cobra.Command, args []string) {
	runPbuilder("login", args)
}

func runPbuilder(operation string, args []string) {
	baseTgz := getBaseTgzPath()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		os.Exit(1)
	}

	bindMountDir := filepath.Join(homeDir, ".chroot-pbuilder", fmt.Sprintf("%s-%s-%s", distribution, architecture, role))
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

func getBaseTgzPath() string {
	if role == "" {
		hash := sha512.Sum512([]byte(distribution + "-" + architecture))
		role = fmt.Sprintf("%x", hash)[:10]
	}

	baseDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		os.Exit(1)
	}

	return filepath.Join(baseDir, fmt.Sprintf("%s-%s-%s.tgz", distribution, architecture, role))
}
