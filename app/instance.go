package app

import (
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strconv"

	"golang.org/x/exp/slog"
)

// Application wide logger
var Logger = slog.New(slog.NewJSONHandler(os.Stdout))

// init logger
func init() {
	slog.SetDefault(Logger)
}

// MinMemory is the minimal acceptable memory for a VM instance
const MinMemory = "512"

// MinDisk is the minimal disk size that can be assigned to a VM instance
const MinDisk = "1G"

// Instance represents a multipass VM, typically created using the launch subcommand.
// Try multipass launch --help for more info.
type Instance struct {
	Cores string
	// Memory in bytes. Could be prefixed with K, M or G.
	// Should be more than min memory
	Memory string
	// Disk space in bytes. Could be prefixed with K, M or G.
	// Should be more than min disk
	Disk string
	Name string
	// CloudInitFile is a path to a cloud init file. Not mendatory.
	CloudInitFile string
	// Name of the cluster in wich this instance belongs to
	Cluster string
	// Image is the name of the image to use, on 20.04 works with our k8s script for now
	Image string
}

// New returns a valid configuration of an instance or an error
// cloudinit is the path of a cloud init script to pass to the method
func New(cores string, memory string, disk string, image string, name string, cloudinit string) (*Instance, error) {
	vmconfig := new(Instance)
	if !validate(memory) {
		return vmconfig, errors.New("invalid memory format")
	}
	if !validate(disk) {
		return vmconfig, errors.New("invalid memory format")
	}
	_, err := strconv.Atoi(cores)
	if err != nil {
		return vmconfig, errors.New("invalid core format")
	}
	vmconfig = &Instance{
		Cores:         cores,
		Memory:        memory,
		Disk:          disk,
		Name:          name,
		CloudInitFile: cloudinit,
		Cluster:       "",
		Image:         image,
	}
	return vmconfig, nil
}

// Create creates a multipass instance
func (vm *Instance) Create() error {
	if vm == nil {
		return errors.New("cannot create vm from nil config")
	}
	cmdConfig := []string{"launch", vm.Image, "-n", vm.Name, "-d", vm.Disk, "-c", vm.Cores, "-m", vm.Memory, "--timeout", "600"}
	if vm.CloudInitFile != "" {
		cmdConfig = append(cmdConfig, "--cloud-init", vm.CloudInitFile)
	}
	cmd := exec.Command("multipass", cmdConfig...)
	err := cmd.Start()
	if err != nil {
		Logger.Error("failed to start instance", err, "name", vm.Name)
		return err
	}
	err = cmd.Wait()
	if err != nil {
		Logger.Error("failed to start instance", err, "name", vm.Name)
		return err
	}
	if cmd.ProcessState.ExitCode() != 0 {
		err = errors.New("non 0 status code encontered by the create process")
		Logger.Error("failed to start instance", err, "name", vm.Name, "cmdOutput")
		return err
	}
	Logger.Info("instance created with sucess", "name", vm.Name)
	return nil
}

// validate checks if an instance memory or disk size is valid
func validate(size string) bool {
	re := regexp.MustCompile(`\d*(K|M|G|k|m|g)`)
	return re.MatchString(size)
}

// checkMpassCmd checks if multipass command is available
func checkMpassCmd() error {
	_, err := exec.LookPath("multipass")
	return err
}
