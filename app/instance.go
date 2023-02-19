package app

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
)

// MinMemory is the minimal acceptable memory for a VM instance
const MinMemory = "2G"

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
	// CloudInitFile is a path to a cloud init file. Not mandatory.
	CloudInitFile string
	// Name of the cluster in which this instance belongs to
	Cluster string
	// Image is the name of the image to use, on 20.04 works with our k8s script for now
	Image string
}

// NewInstanceConfig returns a valid configuration of an instance or an error
// cloudinit is the path of a cloud init script to pass to the method
func NewInstanceConfig(cores string, memory string, disk string, image string, name string, cloudinit string) (*Instance, error) {
	vmconfig := new(Instance)
	if !validateMemoryFormat(memory) {
		return vmconfig, ErrMemFormat
	}
	if !validateMemoryFormat(disk) {
		return vmconfig, ErrMemFormat
	}
	_, err := strconv.Atoi(cores)
	if err != nil {
		return vmconfig, ErrInvalidCoreFmt
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
		err = errors.New("non 0 status code encountered by the create process")
		Logger.Error("failed to start instance", err, "name", vm.Name, "cmdOutput")
		return err
	}
	Logger.Info("instance created with success", "name", vm.Name)
	return nil
}

// Transfer transfers a file to the temp folder of an Instance. It leverages multipass transfer command. dest is
// the name of the dest file. It will appear in the VM as /tmp/dest
func (vm *Instance) Transfer(src string, dst string) error {
	if vm == nil {
		return errors.New("cannot create vm from nil config")
	}
	cmdConfig := []string{"transfer", src, fmt.Sprintf("%s:/tmp/%s", vm.Name, dst)}
	cmd := exec.Command("multipass", cmdConfig...)
	err := cmd.Start()
	if err != nil {
		Logger.Error("failed to copy file to instance", err, "name", vm.Name, "src", src, "dst", dst)
		return err
	}
	err = cmd.Wait()
	if err != nil {
		Logger.Error("failed to copy file to instance", err, "name", vm.Name, "src", src, "dst", dst)
		return err
	}
	if cmd.ProcessState.ExitCode() != 0 {
		err = errors.New("non 0 status code encountered by the file copy command")
		Logger.Error("failed to copy file to instance", err, "name", vm.Name, "src", src, "dst", dst)
		return err
	}
	Logger.Info("file copied with success", "name", vm.Name, "src", src, "dst", dst)
	return nil
}

// validateMemoryFormat checks if an instance memory or disk size is valid. eg: 4G
func validateMemoryFormat(size string) bool {
	re := regexp.MustCompile(`^[1-9][0-9]*(K|M|G|k|m|g)$`)
	return re.MatchString(size)
}
