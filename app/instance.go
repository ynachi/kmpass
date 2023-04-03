package app

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// MinMemory is the minimal acceptable memory for a VM instance
const MinMemory = "2G"

// MinDisk is the minimal disk size that can be assigned to a VM instance
const MinDisk = "1G"

// Instance represents a multipass VM, typically created using the launch subcommand.
// Try multipass launch --help for more info.
type Instance struct {
	Cores int
	// Memory in bytes. Could be prefixed with K, M or G.
	// Should be more than min memory
	Memory string
	// Disk space in bytes. Could be prefixed with K, M or G.
	// Should be more than min disk
	Disk string
	Name string
	// CloudInitFile is a path to a cloud init file. Not mandatory.
	CloudInitFile string
	// Name of the cluster in which this instance belongs to
	Cluster string
	// Image is the name of the image to use, on 20.04 works with our k8s script for now
	Image string
}

// NewInstanceConfig returns a valid configuration of an instance or an error
// cloudinit is the path of a cloud init script to pass to the method
func NewInstanceConfig(cores int, memory string, disk string, image string, name string, cloudinit string) (*Instance, error) {
	vmconfig := new(Instance)
	if !validateMemoryFormat(memory) {
		return vmconfig, ErrMemFormat
	}
	if !validateMemoryFormat(disk) {
		return vmconfig, ErrMemFormat
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
	vmCores := strconv.Itoa(vm.Cores)
	cmdConfig := []string{"launch", vm.Image, "-n", vm.Name, "-d", vm.Disk, "-c", vmCores, "-m", vm.Memory, "--timeout", "600"}
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
		err = errors.New("non 0 status code encountered by the create process")
		Logger.Error("failed to start instance", err, "name", vm.Name, "cmdOutput")
		return err
	}
	Logger.Info("instance created with success", "name", vm.Name)
	return nil
}

// Transfer transfers a file to the temp folder of an Instance. It leverages multipass transfer command. dest is
// the name of the dest file. It will appear in the VM as /tmp/dest
func Transfer(vmName string, src string, dst string) error {
	cmdConfig := []string{"transfer", src, fmt.Sprintf("%s:/tmp/%s", vmName, dst)}
	cmd := exec.Command("multipass", cmdConfig...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		Logger.Error("failed to copy file to instance", err, "name", vmName, "src", src, "dst", dst)
		Logger.Debug("failed to copy file to instance", "stdout", out)
	}
	return err
}

// RunCmd run commands in a VM. It leverages multipass exec command. Returns the combined output (stderr + stdout) of
// the command and an error.
func RunCmd(vmName string, args []string) (string, error) {
	args = append([]string{"exec", vmName, "--"}, args...)
	cmd := exec.Command("multipass", args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// state returns the state of a VM: Running, Stopped, NotExist
// This method should not be run on a nil instance of vm. But if it is done, it will not return an error (we do not
// want this method to error). But a warning log will be fired in that case.
func (vm *Instance) state() string {
	if vm == nil {
		Logger.Warn("cannot get the state of a nil VM", "name", vm.Name)
		return "NotSupported"
	}
	args := []string{"info", vm.Name}
	cmd := exec.Command("multipass", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if !strings.Contains(string(out), "does not exist") {
			return "NotSupported"
		}
		return "NotExist"
	}
	return strings.Fields(string(out))[3]
}

// GetIP gets the IP address of an instance
func (vm *Instance) GetIP() (string, error) {
	if !vm.IsRunning() {
		return "", ErrVMNotRunning
	}
	args := []string{"info", vm.Name}
	cmd := exec.Command("multipass", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New("unexpected exception")
	}
	return strings.Fields(string(out))[5], nil
}

// IsRunning checks whether a VM is running
func (vm *Instance) IsRunning() bool {
	return vm.state() == "Running"
}

// Exist checks whether a VM is already created on the host
func (vm *Instance) Exist() bool {
	return vm.state() != "NotExist"
}

// IsStopped checks whether a VM is stopped
func (vm *Instance) IsStopped() bool {
	return vm.state() == "Stopped"
}

// validateMemoryFormat checks if an instance memory or disk size is valid. eg: 4G
func validateMemoryFormat(size string) bool {
	re := regexp.MustCompile(`^[1-9][0-9]*(K|M|G|k|m|g)$`)
	return re.MatchString(size)
}
