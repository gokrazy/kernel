package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	_ "embed"
)

// see https://www.kernel.org/releases.json
var latest = "https://cdn.kernel.org/pub/linux/kernel/v6.x/linux-6.8.tar.xz"

//go:embed config.addendum.txt
var configAddendum []byte

func downloadKernel() error {
	out, err := os.Create(filepath.Base(latest))
	if err != nil {
		return err
	}
	defer out.Close()
	resp, err := http.Get(latest)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if got, want := resp.StatusCode, http.StatusOK; got != want {
		return fmt.Errorf("unexpected HTTP status code for %s: got %d, want %d", latest, got, want)
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}
	return out.Close()
}

func applyPatches(srcdir string) error {
	patches, err := filepath.Glob("*.patch")
	if err != nil {
		return err
	}
	for _, patch := range patches {
		log.Printf("applying patch %q", patch)
		f, err := os.Open(patch)
		if err != nil {
			return err
		}
		defer f.Close()
		cmd := exec.Command("patch", "-p1")
		cmd.Dir = srcdir
		cmd.Stdin = f
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
		f.Close()
	}

	return nil
}

func compile() error {
	defconfig := exec.Command("make", "ARCH=arm64", "defconfig")
	defconfig.Stdout = os.Stdout
	defconfig.Stderr = os.Stderr
	if err := defconfig.Run(); err != nil {
		return fmt.Errorf("make defconfig: %v", err)
	}

	// Change answers from mod to no if possible
	mod2noconfig := exec.Command("make", "ARCH=arm64", "mod2noconfig")
	mod2noconfig.Stdout = os.Stdout
	mod2noconfig.Stderr = os.Stderr
	if err := mod2noconfig.Run(); err != nil {
		return fmt.Errorf("make olddefconfig: %v", err)
	}

	f, err := os.OpenFile(".config", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(configAddendum); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	olddefconfig := exec.Command("make", "ARCH=arm64", "olddefconfig")
	olddefconfig.Stdout = os.Stdout
	olddefconfig.Stderr = os.Stderr
	if err := olddefconfig.Run(); err != nil {
		return fmt.Errorf("make olddefconfig: %v", err)
	}

	env := append(os.Environ(),
		"ARCH=arm64",
		"CROSS_COMPILE=aarch64-linux-gnu-",
		"KBUILD_BUILD_USER=gokrazy",
		"KBUILD_BUILD_HOST=docker",
		"KBUILD_BUILD_TIMESTAMP=Wed Mar  1 20:57:29 UTC 2017",
	)
	make := exec.Command("make", "Image.gz", "dtbs", "modules", "-j"+strconv.Itoa(runtime.NumCPU()))
	make.Env = env
	make.Stdout = os.Stdout
	make.Stderr = os.Stderr
	if err := make.Run(); err != nil {
		return fmt.Errorf("make: %v", err)
	}

	make = exec.Command("make", "INSTALL_MOD_PATH=/tmp/buildresult", "modules_install", "-j"+strconv.Itoa(runtime.NumCPU()))
	make.Env = env
	make.Stdout = os.Stdout
	make.Stderr = os.Stderr
	if err := make.Run(); err != nil {
		return fmt.Errorf("make: %v", err)
	}

	return nil
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

func main() {
	log.Printf("downloading kernel source: %s", latest)
	if err := downloadKernel(); err != nil {
		log.Fatal(err)
	}

	log.Printf("unpacking kernel source")
	untar := exec.Command("tar", "xf", filepath.Base(latest))
	untar.Stdout = os.Stdout
	untar.Stderr = os.Stderr
	if err := untar.Run(); err != nil {
		log.Fatalf("untar: %v", err)
	}

	srcdir := strings.TrimSuffix(filepath.Base(latest), ".tar.xz")

	log.Printf("applying patches")
	if err := applyPatches(srcdir); err != nil {
		log.Fatal(err)
	}

	if err := os.Chdir(srcdir); err != nil {
		log.Fatal(err)
	}

	log.Printf("compiling kernel")
	if err := compile(); err != nil {
		log.Fatal(err)
	}

	if err := copyFile("/tmp/buildresult/vmlinuz", "arch/arm64/boot/Image"); err != nil {
		log.Fatal(err)
	}

	if err := copyFile("/tmp/buildresult/bcm2710-rpi-3-b.dtb", "arch/arm64/boot/dts/broadcom/bcm2837-rpi-3-b.dtb"); err != nil {
		log.Fatal(err)
	}

	if err := copyFile("/tmp/buildresult/bcm2710-rpi-3-b-plus.dtb", "arch/arm64/boot/dts/broadcom/bcm2837-rpi-3-b-plus.dtb"); err != nil {
		log.Fatal(err)
	}

	if err := copyFile("/tmp/buildresult/bcm2710-rpi-cm3.dtb", "arch/arm64/boot/dts/broadcom/bcm2837-rpi-cm3-io3.dtb"); err != nil {
		log.Fatal(err)
	}

	if err := copyFile("/tmp/buildresult/bcm2711-rpi-4-b.dtb", "arch/arm64/boot/dts/broadcom/bcm2711-rpi-4-b.dtb"); err != nil {
		log.Fatal(err)
	}

	if err := copyFile("/tmp/buildresult/bcm2710-rpi-zero-2-w.dtb", "arch/arm64/boot/dts/broadcom/bcm2837-rpi-zero-2-w.dtb"); err != nil {
		log.Fatal(err)
	}

}
