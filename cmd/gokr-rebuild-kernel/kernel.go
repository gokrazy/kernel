package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"
	"time"
)

const dockerFileContents = `
FROM debian:stretch

RUN apt-get update && apt-get install -y crossbuild-essential-arm64 bc libssl-dev bison flex

COPY gokr-build-kernel /usr/bin/gokr-build-kernel
{{- range $idx, $path := .Patches }}
COPY {{ $path }} /usr/src/{{ $path }}
{{- end }}

RUN echo 'builduser:x:{{ .Uid }}:{{ .Gid }}:nobody:/:/bin/sh' >> /etc/passwd && \
    chown -R {{ .Uid }}:{{ .Gid }} /usr/src

USER builduser
WORKDIR /usr/src
ENTRYPOINT /usr/bin/gokr-build-kernel
`

var dockerFileTmpl = template.Must(template.New("dockerfile").
	Funcs(map[string]interface{}{
		"basename": func(path string) string {
			return filepath.Base(path)
		},
	}).
	Parse(dockerFileContents))

var patchFiles = []string{
	"0001-Revert-add-index-to-the-ethernet-alias.patch",
	// serial
	"0101-expose-UART0-ttyAMA0-on-GPIO-14-15-disable-UART1-tty.patch",
	"0102-expose-UART0-ttyAMA0-on-GPIO-14-15-disable-UART1-tty.patch",
	"0103-expose-UART0-ttyAMA0-on-GPIO-14-15-disable-UART1-tty.patch",
	"0104-bluetooth.patch",
	"0105-bluetooth-cm4.patch",
	// spi
	"0201-enable-spidev.patch",
	// logo
	"0001-gokrazy-logo.patch",
}

func copyFile(dest, src string) error {
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	st, err := in.Stat()
	if err != nil {
		return err
	}
	if err := out.Chmod(st.Mode()); err != nil {
		return err
	}
	return out.Close()
}

var gopath = mustGetGopath()

func mustGetGopath() string {
	gopathb, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		log.Panic(err)
	}
	return strings.TrimSpace(string(gopathb))
}

func find(filename string) (string, error) {
	if _, err := os.Stat(filename); err == nil {
		return filename, nil
	}

	path := filepath.Join(gopath, "src", "github.com", "gokrazy", "kernel", filename)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("could not find file %q (looked in . and %s)", filename, path)
}

func getContainerExecutable() (string, error) {
	// Probe podman first, because the docker binary might actually
	// be a thin podman wrapper with podman behavior.
	choices := []string{"podman", "docker"}
	for _, exe := range choices {
		p, err := exec.LookPath(exe)
		if err != nil {
			continue
		}
		resolved, err := filepath.EvalSymlinks(p)
		if err != nil {
			return "", err
		}
		return resolved, nil
	}
	return "", fmt.Errorf("none of %v found in $PATH", choices)
}

var fatalErr error

func fatal(e error) {
	log.Println(e)
	fatalErr = e
}

func main() {
	defer func() {
		if fatalErr != nil {
			os.Exit(1)
		}
	}()

	var overwriteContainerExecutable = flag.String("overwrite_container_executable",
		"",
		"E.g. docker or podman to overwrite the automatically detected container executable")
	flag.Parse()
	executable, err := getContainerExecutable()
	if err != nil {
		fatal(err)
		return
	}
	if *overwriteContainerExecutable != "" {
		executable = *overwriteContainerExecutable
	}
	execName := filepath.Base(executable)
	// We explicitly use /tmp, because Docker only allows volume mounts under
	// certain paths on certain platforms, see
	// e.g. https://docs.docker.com/docker-for-mac/osxfs/#namespaces for macOS.
	tmp, err := os.MkdirTemp("/tmp", "gokr-rebuild-kernel")
	if err != nil {
		fatal(err)
		return
	}
	defer os.RemoveAll(tmp)

	cmd := exec.Command("go", "install", "github.com/gokrazy/kernel/cmd/gokr-build-kernel")
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOBIN="+tmp)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fatal(fmt.Errorf("%v: %v", cmd.Args, err))
	}

	buildPath := filepath.Join(tmp, "gokr-build-kernel")

	var patchPaths []string
	for _, filename := range patchFiles {
		path, err := find(filename)
		if err != nil {
			fatal(err)
			return
		}
		patchPaths = append(patchPaths, path)
	}

	kernelPath, err := find("vmlinuz")
	if err != nil {
		fatal(err)
		return
	}
	dtbPath, err := find("bcm2710-rpi-3-b.dtb")
	if err != nil {
		fatal(err)
		return
	}
	dtbPlusPath, err := find("bcm2710-rpi-3-b-plus.dtb")
	if err != nil {
		fatal(err)
		return
	}
	dtbZero2WPath, err := find("bcm2710-rpi-zero-2.dtb")
	if err != nil {
		fatal(err)
		return
	}
	dtbCM3Path, err := find("bcm2710-rpi-cm3.dtb")
	if err != nil {
		fatal(err)
		return
	}
	dtb4Path, err := find("bcm2711-rpi-4-b.dtb")
	if err != nil {
		fatal(err)
		return
	}

	// Copy all files into the temporary directory so that docker
	// includes them in the build context.
	for _, path := range patchPaths {
		if err := copyFile(filepath.Join(tmp, filepath.Base(path)), path); err != nil {
			fatal(err)
			return
		}
	}

	u, err := user.Current()
	if err != nil {
		fatal(err)
		return
	}
	dockerFile, err := os.Create(filepath.Join(tmp, "Dockerfile"))
	if err != nil {
		fatal(err)
		return
	}

	if err := dockerFileTmpl.Execute(dockerFile, struct {
		Uid       string
		Gid       string
		BuildPath string
		Patches   []string
	}{
		Uid:       u.Uid,
		Gid:       u.Gid,
		BuildPath: buildPath,
		Patches:   patchFiles,
	}); err != nil {
		fatal(err)
		return
	}

	if err := dockerFile.Close(); err != nil {
		fatal(err)
		return
	}

	log.Printf("building %s container for kernel compilation", execName)

	dockerBuild := exec.Command(execName,
		"build",
		"--rm=true",
		"--tag=gokr-rebuild-kernel",
		".")
	dockerBuild.Dir = tmp
	dockerBuild.Stdout = os.Stdout
	dockerBuild.Stderr = os.Stderr
	if err := dockerBuild.Run(); err != nil {
		fatal(fmt.Errorf("%s build: %v (cmd: %v)", execName, err, dockerBuild.Args))
		return
	}

	log.Printf("compiling kernel")

	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	go func() {
		<-signalChan
		cancel()
		log.Println("Stopping ...")
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	rand.Seed(time.Now().UnixNano())
	randBytes := make([]byte, 4)
	rand.Read(randBytes)
	containerId := "compilekernel-" + hex.EncodeToString(randBytes)

	var dockerRun *exec.Cmd
	if execName == "podman" {
		dockerRun = exec.CommandContext(ctx, executable,
			"run",
			"--rm",
			"--userns=keep-id",
			"--rm",
			"--name", containerId,
			"--volume", tmp+":/tmp/buildresult:Z",
			"gokr-rebuild-kernel")
	} else {
		dockerRun = exec.CommandContext(ctx, executable,
			"run",
			"--rm",
			"--name", containerId,
			"--volume", tmp+":/tmp/buildresult:Z",
			"gokr-rebuild-kernel")
	}
	defer func() {
		if !dockerRun.ProcessState.Success() {
			exec.Command(
				executable,
				"stop", containerId,
			).Run()
		}
	}()
	dockerRun.Dir = tmp
	dockerRun.Stdout = os.Stdout
	dockerRun.Stderr = os.Stderr
	if err := dockerRun.Run(); err != nil {
		fatal(fmt.Errorf("%s run: %v (cmd: %v)", execName, err, dockerRun.Args))
		return
	}

	if err := copyFile(kernelPath, filepath.Join(tmp, "vmlinuz")); err != nil {
		fatal(err)
		return
	}

	if err := copyFile(dtbPath, filepath.Join(tmp, "bcm2710-rpi-3-b.dtb")); err != nil {
		fatal(err)
		return
	}

	// Until the Raspberry Pi Zero 2 W DTB is built by the kernel, use bcm2710-rpi-3-b.dtb:
	if err := copyFile(dtbZero2WPath, filepath.Join(tmp, "bcm2710-rpi-3-b.dtb")); err != nil {
		fatal(err)
		return
	}

	if err := copyFile(dtbPlusPath, filepath.Join(tmp, "bcm2710-rpi-3-b-plus.dtb")); err != nil {
		fatal(err)
		return
	}

	if err := copyFile(dtbCM3Path, filepath.Join(tmp, "bcm2710-rpi-cm3.dtb")); err != nil {
		fatal(err)
		return
	}

	if err := copyFile(dtb4Path, filepath.Join(tmp, "bcm2711-rpi-4-b.dtb")); err != nil {
		fatal(err)
		return
	}
}
