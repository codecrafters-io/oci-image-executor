package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/codecrafters-io/oci-image-executor/internal"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/openlyinc/pointy"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	log "github.com/sirupsen/logrus"
)

func main() {
	config := internal.ParseConfig()
	config.ValidatePathsExist()

	rootFSBuilder, err := internal.NewRootFSBuilder(config)
	if err != nil {
		fmt.Println(err)
		os.Exit(11)
	}

	fmt.Println("Before RootFS Build", time.Now().Format("2006.01.02 15:04:05.50"))

	rootFSPath, err := rootFSBuilder.Build()
	if err != nil {
		fmt.Println(err)
		os.Exit(11)
	}

	fmt.Println("After RootFS Build", time.Now().Format("2006.01.02 15:04:05.50"))

	defer rootFSBuilder.Cleanup()

	rootFSBuilder.PrintDebugInfo()

	if err := runVMM(context.Background(), rootFSPath); err != nil {
		log.Fatalf(err.Error())
	}

	fmt.Printf("no error!")
}

// Run a vmm with a given set of options
func runVMM(ctx context.Context, rootFSPath string) error {
	vmmCtx, vmmCancel := context.WithCancel(ctx)
	defer vmmCancel()

	logger := log.New()

	machineOpts := []firecracker.Opt{
		firecracker.WithLogger(log.NewEntry(logger)),
	}

	firecrackerBinary, err := exec.LookPath("firecracker")
	if err != nil {
		return err
	}

	finfo, err := os.Stat(firecrackerBinary)
	if os.IsNotExist(err) {
		return fmt.Errorf("binary %q does not exist: %v", firecrackerBinary, err)
	}

	if err != nil {
		return fmt.Errorf("failed to stat binary, %q: %v", firecrackerBinary, err)
	}

	if finfo.IsDir() {
		return fmt.Errorf("binary, %q, is a directory", firecrackerBinary)
	} else if finfo.Mode()&0111 == 0 {
		return fmt.Errorf("binary, %q, is not executable. Check permissions of binary", firecrackerBinary)
	}

	cmd := firecracker.VMCommandBuilder{}.
		WithBin(firecrackerBinary).
		WithSocketPath("/tmp/firecracker.socket").
		WithStdin(os.Stdin).
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		Build(ctx)

	machineOpts = append(machineOpts, firecracker.WithProcessRunner(cmd))

	var vcpuCount int64 = 2
	var memSize int64 = 1024
	var htEnabled bool = false

	config := firecracker.Config{
		Drives: []models.Drive{
			{
				DriveID:      pointy.String("rootfs"),
				IsReadOnly:   pointy.Bool(false),
				IsRootDevice: pointy.Bool(true),
				PathOnHost:   pointy.String(rootFSPath),
			},
		},
		KernelImagePath: "/root/firecracker-resources/vmlinux.bin",
		KernelArgs:      "ro console=ttyS0 noapic reboot=k panic=1 tsc=reliable 8250.nr_uarts=0 ipv6.disable=1 pci=off i8042.noaux i8042.nomux i8042.nopnp i8042.dumbkbd init=/init",
		LogLevel:        "Debug",
		LogPath:         "/tmp/firecracker-logs",
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  &vcpuCount,
			MemSizeMib: &memSize,
			HtEnabled:  &htEnabled,
		},
		NetworkInterfaces: []firecracker.NetworkInterface{
			{
				CNIConfiguration: &firecracker.CNIConfiguration{
					NetworkName: "fcnet",
					IfName:      "veth0",
				},
			},
		},
		SocketPath: "/tmp/firecracker.socket",
	}
	if err := config.Validate(); err != nil {
		return err
	}

	m, err := firecracker.NewMachine(vmmCtx, config, machineOpts...)
	if err != nil {
		return fmt.Errorf("failed creating machine: %s", err)
	}

	if err := m.Start(vmmCtx); err != nil {
		return fmt.Errorf("failed to start machine: %v", err)
	}
	defer m.StopVMM()

	installSignalHandlers(vmmCtx, m)

	// wait for the VMM to exit
	if err := m.Wait(vmmCtx); err != nil {
		return fmt.Errorf("wait returned an error %s", err)
	}
	log.Printf("Start machine was happy")
	return nil
}

// Install custom signal handlers:
func installSignalHandlers(ctx context.Context, m *firecracker.Machine) {
	go func() {
		// Clear some default handlers installed by the firecracker SDK:
		signal.Reset(os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

		for {
			switch s := <-c; {
			case s == syscall.SIGTERM || s == os.Interrupt:
				log.Printf("Caught signal: %s, requesting clean shutdown", s.String())
				m.Shutdown(ctx)
			case s == syscall.SIGQUIT:
				log.Printf("Caught signal: %s, forcing shutdown", s.String())
				m.StopVMM()
			}
		}
	}()
}
