package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// see https://www.kernel.org/releases.json
// NOTE: prefer cdn.kernel.org over www.kernel.org because the latter
// does not work with HTTP2: https://github.com/golang/go/issues/19035
var latest = "https://cdn.kernel.org/pub/linux/kernel/v4.x/linux-4.10.6.tar.xz"

const configAddendum = `
CONFIG_ARCH_BCM2835=y 
CONFIG_HW_RANDOM_BCM2835=y
CONFIG_DMA_BCM2835=y
# CONFIG_I2C_BCM2835 prints errors: i2c-bcm2835 3f805000.i2c: i2c transfer failed: 100
CONFIG_SPI_BCM2835=y
CONFIG_SPI_BCM2835AUX=y
CONFIG_SERIAL_8250_BCM2835AUX=y
CONFIG_BCM2835_WDT=y
CONFIG_SND_BCM2835_SOC_I2S=y
CONFIG_USB_USBNET=y
CONFIG_USB_NET_SMSC95XX=y
CONFIG_BCM2835_MBOX=y
CONFIG_RASPBERRYPI_FIRMWARE=y
CONFIG_RASPBERRYPI_POWER=y
CONFIG_IPV6=y

# For physically connecting the gokrazy bakery:
CONFIG_USB_SERIAL=y
CONFIG_USB_SERIAL_FTDI_SIO=y

# TODO: trim the settings below to the minimum set that works (taken from debian)
##
## file: arch/arm64/Kconfig
##
CONFIG_PCI=y
CONFIG_ARM64_ERRATUM_834220=y
#. Until we decide how/whether to handle this in userland as well
# CONFIG_ARM64_ERRATUM_843419 is not set
## choice: Virtual address space size
CONFIG_ARM64_VA_BITS_48=y
## end choice
CONFIG_SCHED_MC=y
CONFIG_SECCOMP=y
CONFIG_XEN=y
CONFIG_RANDOMIZE_BASE=y
CONFIG_RANDOMIZE_MODULE_REGION_FULL=y
CONFIG_ARM64_ACPI_PARKING_PROTOCOL=y
CONFIG_EFI_SECURE_BOOT_SECURELEVEL=y
CONFIG_COMPAT=y

##
## file: arch/arm64/crypto/Kconfig
##
CONFIG_ARM64_CRYPTO=y
CONFIG_CRYPTO_SHA1_ARM64_CE=y
CONFIG_CRYPTO_SHA2_ARM64_CE=y
CONFIG_CRYPTO_GHASH_ARM64_CE=y
CONFIG_CRYPTO_AES_ARM64_CE=y
CONFIG_CRYPTO_AES_ARM64_CE_CCM=y
CONFIG_CRYPTO_AES_ARM64_CE_BLK=y
# CONFIG_CRYPTO_AES_ARM64_NEON_BLK is not set
CONFIG_CRYPTO_CRC32_ARM64=y

##
## file: arch/arm64/kvm/Kconfig
##
CONFIG_VIRTUALIZATION=y
CONFIG_KVM=y

##
## file: arch/arm64/Kconfig.platforms
##
CONFIG_ARCH_BCM2835=y
CONFIG_ARCH_HISI=y
CONFIG_ARCH_MESON=y
CONFIG_ARCH_MVEBU=y
CONFIG_ARCH_QCOM=y
CONFIG_ARCH_SEATTLE=y
CONFIG_ARCH_TEGRA=y
CONFIG_ARCH_THUNDER=y
CONFIG_ARCH_VEXPRESS=y
CONFIG_ARCH_XGENE=y

##
## file: drivers/acpi/Kconfig
##
CONFIG_ACPI=y

##
## file: drivers/ata/Kconfig
##
CONFIG_SATA_AHCI_PLATFORM=y
CONFIG_AHCI_MVEBU=y
CONFIG_AHCI_TEGRA=y
CONFIG_AHCI_XGENE=y
CONFIG_SATA_AHCI_SEATTLE=y

##
## file: drivers/bluetooth/Kconfig
##
CONFIG_BT_HCIUART=y
CONFIG_BT_QCOMSMD=y

##
## file: drivers/bus/Kconfig
##
CONFIG_QCOM_EBI2=y
CONFIG_TEGRA_ACONNECT=y

##
## file: drivers/char/hw_random/Kconfig
##
CONFIG_HW_RANDOM_BCM2835=y
CONFIG_HW_RANDOM_HISI=y
CONFIG_HW_RANDOM_MSM=y
CONFIG_HW_RANDOM_XGENE=y
CONFIG_HW_RANDOM_MESON=y
CONFIG_HW_RANDOM_CAVIUM=y

##
## file: drivers/char/ipmi/Kconfig
##
CONFIG_IPMI_HANDLER=y
CONFIG_IPMI_DEVICE_INTERFACE=y
CONFIG_IPMI_SSIF=y

##
## file: drivers/clk/Kconfig
##
CONFIG_COMMON_CLK_XGENE=y

##
## file: drivers/clk/hisilicon/Kconfig
##
CONFIG_STUB_CLK_HI6220=y

##
## file: drivers/clk/qcom/Kconfig
##
CONFIG_COMMON_CLK_QCOM=y
CONFIG_MSM_GCC_8916=y
CONFIG_MSM_GCC_8996=y
CONFIG_MSM_MMCC_8996=y

##
## file: drivers/cpufreq/Kconfig
##
CONFIG_CPUFREQ_DT=y

##
## file: drivers/cpuidle/Kconfig.arm
##
CONFIG_ARM_CPUIDLE=y

##
## file: drivers/crypto/Kconfig
##
CONFIG_CRYPTO_DEV_QCE=y

##
## file: drivers/dma/Kconfig
##
CONFIG_DMADEVICES=y
CONFIG_DMA_BCM2835=y
CONFIG_K3_DMA=y
CONFIG_MV_XOR=y
CONFIG_MV_XOR_V2=y
CONFIG_TEGRA20_APB_DMA=y
CONFIG_TEGRA210_ADMA=y
CONFIG_XGENE_DMA=y

##
## file: drivers/dma/qcom/Kconfig
##
CONFIG_QCOM_BAM_DMA=y
CONFIG_QCOM_HIDMA_MGMT=y
CONFIG_QCOM_HIDMA=y

##
## file: drivers/edac/Kconfig
##
CONFIG_EDAC=y
CONFIG_EDAC_MM_EDAC=y
CONFIG_EDAC_XGENE=y

##
## file: drivers/extcon/Kconfig
##
CONFIG_EXTCON=y
CONFIG_EXTCON_QCOM_SPMI_MISC=y
CONFIG_EXTCON_USB_GPIO=y

##
## file: drivers/firmware/Kconfig
##
CONFIG_RASPBERRYPI_FIRMWARE=y

##
## file: drivers/gpio/Kconfig
##
CONFIG_GPIOLIB=y
CONFIG_GPIO_PL061=y
CONFIG_GPIO_XGENE=y
CONFIG_GPIO_XGENE_SB=y
CONFIG_GPIO_PCA953X=y
CONFIG_GPIO_PCA953X_IRQ=y
CONFIG_GPIO_MAX77620=y

##
## file: drivers/gpu/drm/Kconfig
##
CONFIG_DRM=y

##
## file: drivers/gpu/drm/bridge/adv7511/Kconfig
##
CONFIG_DRM_I2C_ADV7511=y

##
## file: drivers/gpu/drm/hisilicon/kirin/Kconfig
##
CONFIG_DRM_HISI_KIRIN=y

##
## file: drivers/gpu/drm/meson/Kconfig
##
CONFIG_DRM_MESON=y

##
## file: drivers/gpu/drm/msm/Kconfig
##
CONFIG_DRM_MSM=y
CONFIG_DRM_MSM_DSI=y
CONFIG_DRM_MSM_DSI_PLL=y
CONFIG_DRM_MSM_DSI_28NM_PHY=y
CONFIG_DRM_MSM_DSI_20NM_PHY=y

##
## file: drivers/gpu/drm/panel/Kconfig
##
CONFIG_DRM_PANEL_SIMPLE=y

##
## file: drivers/gpu/drm/tegra/Kconfig
##
CONFIG_DRM_TEGRA=y
CONFIG_DRM_TEGRA_STAGING=y

##
## file: drivers/gpu/drm/vc4/Kconfig
##
CONFIG_DRM_VC4=y

##
## file: drivers/gpu/host1x/Kconfig
##
CONFIG_TEGRA_HOST1X=y

##
## file: drivers/hwmon/Kconfig
##
CONFIG_SENSORS_XGENE=y

##
## file: drivers/hwspinlock/Kconfig
##
CONFIG_HWSPINLOCK_QCOM=y

##
## file: drivers/iio/adc/Kconfig
##
CONFIG_QCOM_SPMI_IADC=y
CONFIG_QCOM_SPMI_VADC=y

##
## file: drivers/input/keyboard/Kconfig
##
CONFIG_KEYBOARD_GPIO=y
CONFIG_KEYBOARD_TEGRA=y

##
## file: drivers/input/misc/Kconfig
##
CONFIG_INPUT_MISC=y
CONFIG_INPUT_PM8941_PWRKEY=y
CONFIG_INPUT_UINPUT=y
CONFIG_INPUT_HISI_POWERKEY=y

##
## file: drivers/iommu/Kconfig
##
CONFIG_TEGRA_IOMMU_SMMU=y
CONFIG_ARM_SMMU=y
CONFIG_ARM_SMMU_V3=y

##
## file: drivers/leds/Kconfig
##
CONFIG_LEDS_GPIO=y

##
## file: drivers/mailbox/Kconfig
##
CONFIG_MAILBOX=y
CONFIG_BCM2835_MBOX=y
CONFIG_HI6220_MBOX=y
CONFIG_XGENE_SLIMPRO_MBOX=y

##
## file: drivers/memory/tegra/Kconfig
##
CONFIG_TEGRA_MC=y

##
## file: drivers/mfd/Kconfig
##
CONFIG_MFD_CROS_EC=y
CONFIG_MFD_CROS_EC_I2C=y
CONFIG_MFD_CROS_EC_SPI=y
CONFIG_MFD_HI655X_PMIC=y
CONFIG_MFD_MAX77620=y
CONFIG_MFD_QCOM_RPM=y
CONFIG_MFD_SPMI_PMIC=y

##
## file: drivers/misc/Kconfig
##
CONFIG_QCOM_COINCELL=y

##
## file: drivers/misc/ti-st/Kconfig
##
CONFIG_TI_ST=y

##
## file: drivers/mmc/Kconfig
##
CONFIG_MMC=y

##
## file: drivers/mmc/host/Kconfig
##
CONFIG_MMC_ARMMMCI=y
CONFIG_MMC_QCOM_DML=y
CONFIG_MMC_SDHCI_PLTFM=y
CONFIG_MMC_SDHCI_TEGRA=y
CONFIG_MMC_SDHCI_IPROC=y
CONFIG_MMC_MESON_GX=y
CONFIG_MMC_SDHCI_MSM=y
CONFIG_MMC_SPI=y
CONFIG_MMC_DW=y
CONFIG_MMC_DW_K3=y

##
## file: drivers/mtd/spi-nor/Kconfig
##
CONFIG_SPI_HISI_SFC=y

##
## file: drivers/net/ethernet/Kconfig
##
CONFIG_FEALNX=y

##
## file: drivers/net/ethernet/3com/Kconfig
##
CONFIG_NET_VENDOR_3COM=y
CONFIG_VORTEX=y
CONFIG_TYPHOON=y

##
## file: drivers/net/ethernet/8390/Kconfig
##
CONFIG_NET_VENDOR_8390=y
CONFIG_NE2K_PCI=y

##
## file: drivers/net/ethernet/adaptec/Kconfig
##
CONFIG_NET_VENDOR_ADAPTEC=y
CONFIG_ADAPTEC_STARFIRE=y

##
## file: drivers/net/ethernet/amd/Kconfig
##
CONFIG_AMD_XGBE=y

##
## file: drivers/net/ethernet/apm/xgene/Kconfig
##
CONFIG_NET_XGENE=y

##
## file: drivers/net/ethernet/cavium/Kconfig
##
CONFIG_NET_VENDOR_CAVIUM=y
CONFIG_THUNDER_NIC_PF=y
CONFIG_THUNDER_NIC_VF=y
CONFIG_THUNDER_NIC_BGX=y
CONFIG_THUNDER_NIC_RGX=y

##
## file: drivers/net/ethernet/dec/tulip/Kconfig
##
CONFIG_NET_TULIP=y
CONFIG_DE2104X=y
CONFIG_TULIP=y
# CONFIG_TULIP_MWI is not set
# CONFIG_TULIP_MMIO is not set
CONFIG_WINBOND_840=y
CONFIG_DM9102=y

##
## file: drivers/net/ethernet/dlink/Kconfig
##
CONFIG_NET_VENDOR_DLINK=y
CONFIG_SUNDANCE=y
# CONFIG_SUNDANCE_MMIO is not set

##
## file: drivers/net/ethernet/hisilicon/Kconfig
##
CONFIG_NET_VENDOR_HISILICON=y
CONFIG_HIX5HD2_GMAC=y
CONFIG_HISI_FEMAC=y
CONFIG_HIP04_ETH=y
CONFIG_HNS=y
CONFIG_HNS_DSAF=y
CONFIG_HNS_ENET=y

##
## file: drivers/net/ethernet/intel/Kconfig
##
CONFIG_NET_VENDOR_INTEL=y
CONFIG_E100=y

##
## file: drivers/net/ethernet/natsemi/Kconfig
##
CONFIG_NET_VENDOR_NATSEMI=y
CONFIG_NATSEMI=y

##
## file: drivers/net/ethernet/realtek/Kconfig
##
CONFIG_8139CP=y
CONFIG_8139TOO=y

##
## file: drivers/net/ethernet/smsc/Kconfig
##
CONFIG_NET_VENDOR_SMSC=y
CONFIG_SMC91X=y
CONFIG_EPIC100=y
CONFIG_SMSC911X=y

##
## file: drivers/net/ethernet/stmicro/stmmac/Kconfig
##
CONFIG_STMMAC_ETH=y
CONFIG_STMMAC_PLATFORM=y
CONFIG_DWMAC_GENERIC=y
CONFIG_DWMAC_IPQ806X=y
CONFIG_DWMAC_MESON=y

##
## file: drivers/net/fddi/Kconfig
##
CONFIG_FDDI=y
CONFIG_SKFP=y

##
## file: drivers/net/phy/Kconfig
##
CONFIG_MDIO_HISI_FEMAC=y
CONFIG_MDIO_THUNDER=y
CONFIG_MDIO_XGENE=y
CONFIG_MESON_GXL_PHY=y

##
## file: drivers/net/wireless/ath/wcn36xx/Kconfig
##
CONFIG_WCN36XX=y

##
## file: drivers/net/wireless/ti/Kconfig
##
CONFIG_WLAN_VENDOR_TI=y
CONFIG_WILINK_PLATFORM_DATA=y

##
## file: drivers/net/wireless/ti/wl1251/Kconfig
##
CONFIG_WL1251=y
CONFIG_WL1251_SPI=y
CONFIG_WL1251_SDIO=y

##
## file: drivers/net/wireless/ti/wl12xx/Kconfig
##
CONFIG_WL12XX=y

##
## file: drivers/net/wireless/ti/wl18xx/Kconfig
##
CONFIG_WL18XX=y

##
## file: drivers/net/wireless/ti/wlcore/Kconfig
##
CONFIG_WLCORE=y
CONFIG_WLCORE_SPI=y
CONFIG_WLCORE_SDIO=y

##
## file: drivers/nvmem/Kconfig
##
CONFIG_QCOM_QFPROM=y

##
## file: drivers/pci/host/Kconfig
##
CONFIG_PCI_AARDVARK=y
CONFIG_PCI_HOST_GENERIC=y
CONFIG_PCI_XGENE=y
CONFIG_PCI_HISI=y
CONFIG_PCIE_QCOM=y
CONFIG_PCI_HOST_THUNDER_PEM=y
CONFIG_PCI_HOST_THUNDER_ECAM=y
CONFIG_PCIE_ARMADA_8K=y

##
## file: drivers/phy/Kconfig
##
CONFIG_PHY_HI6220_USB=y
CONFIG_PHY_QCOM_APQ8064_SATA=y
CONFIG_PHY_QCOM_IPQ806X_SATA=y
CONFIG_PHY_XGENE=y
CONFIG_PHY_QCOM_UFS=y
CONFIG_PHY_MESON8B_USB2=y

##
## file: drivers/phy/tegra/Kconfig
##
CONFIG_PHY_TEGRA_XUSB=y

##
## file: drivers/pinctrl/Kconfig
##
CONFIG_PINCTRL_AMD=y
CONFIG_PINCTRL_SINGLE=y
CONFIG_PINCTRL_MAX77620=y

##
## file: drivers/pinctrl/qcom/Kconfig
##
CONFIG_PINCTRL_MSM8916=y
CONFIG_PINCTRL_MSM8996=y
CONFIG_PINCTRL_QCOM_SPMI_PMIC=y
CONFIG_PINCTRL_QCOM_SSBI_PMIC=y

##
## file: drivers/platform/chrome/Kconfig
##
CONFIG_CHROME_PLATFORMS=y

##
## file: drivers/power/reset/Kconfig
##
CONFIG_POWER_RESET_HISI=y
CONFIG_POWER_RESET_MSM=y
CONFIG_POWER_RESET_VEXPRESS=y
CONFIG_POWER_RESET_XGENE=y
CONFIG_POWER_RESET_SYSCON=y
CONFIG_POWER_RESET_SYSCON_POWEROFF=y

##
## file: drivers/power/supply/Kconfig
##
CONFIG_BATTERY_BQ27XXX=y
CONFIG_CHARGER_QCOM_SMBB=y

##
## file: drivers/pwm/Kconfig
##
CONFIG_PWM=y
CONFIG_PWM_BCM2835=y
CONFIG_PWM_MESON=y
CONFIG_PWM_TEGRA=y

##
## file: drivers/regulator/Kconfig
##
CONFIG_REGULATOR=y
CONFIG_REGULATOR_FIXED_VOLTAGE=y
CONFIG_REGULATOR_HI655X=y
CONFIG_REGULATOR_MAX77620=y
CONFIG_REGULATOR_PWM=y
CONFIG_REGULATOR_QCOM_RPM=y
CONFIG_REGULATOR_QCOM_SMD_RPM=y
CONFIG_REGULATOR_QCOM_SPMI=y

##
## file: drivers/remoteproc/Kconfig
##
CONFIG_QCOM_Q6V5_PIL=y
#. We want to enable this but it currently results in a dependency loop!
# CONFIG_QCOM_WCNSS_PIL is not set

##
## file: drivers/reset/Kconfig
##
CONFIG_RESET_CONTROLLER=y
CONFIG_RESET_MESON=y

##
## file: drivers/reset/hisilicon/Kconfig
##
CONFIG_COMMON_RESET_HI6220=y

##
## file: drivers/rtc/Kconfig
##
CONFIG_RTC_DRV_DS1307=y
CONFIG_RTC_DRV_MAX77686=y
CONFIG_RTC_DRV_EFI=y
CONFIG_RTC_DRV_PL031=y
CONFIG_RTC_DRV_PM8XXX=y
CONFIG_RTC_DRV_TEGRA=y
CONFIG_RTC_DRV_XGENE=y

##
## file: drivers/scsi/Kconfig
##
CONFIG_SCSI_DMX3191D=y

##
## file: drivers/scsi/hisi_sas/Kconfig
##
CONFIG_SCSI_HISI_SAS=y

##
## file: drivers/soc/bcm/Kconfig
##
CONFIG_RASPBERRYPI_POWER=y

##
## file: drivers/soc/qcom/Kconfig
##
CONFIG_QCOM_GSBI=y
CONFIG_QCOM_SMEM=y
CONFIG_QCOM_SMD=y
CONFIG_QCOM_SMD_RPM=y
CONFIG_QCOM_SMP2P=y
CONFIG_QCOM_SMSM=y
CONFIG_QCOM_WCNSS_CTRL=y

##
## file: drivers/soc/tegra/Kconfig
##
CONFIG_ARCH_TEGRA_132_SOC=y
CONFIG_ARCH_TEGRA_210_SOC=y

##
## file: drivers/spi/Kconfig
##
CONFIG_SPI_BCM2835=y
CONFIG_SPI_BCM2835AUX=y
CONFIG_SPI_MESON_SPIFC=y
CONFIG_SPI_QUP=y
CONFIG_SPI_TEGRA114=y
CONFIG_SPI_TEGRA20_SFLASH=y
CONFIG_SPI_TEGRA20_SLINK=y
CONFIG_SPI_THUNDERX=y

##
## file: drivers/spmi/Kconfig
##
CONFIG_SPMI=y
CONFIG_SPMI_MSM_PMIC_ARB=y

##
## file: drivers/thermal/Kconfig
##
CONFIG_THERMAL=y
CONFIG_CPU_THERMAL=y
CONFIG_HISI_THERMAL=y
CONFIG_QCOM_SPMI_TEMP_ALARM=y

##
## file: drivers/thermal/qcom/Kconfig
##
CONFIG_QCOM_TSENS=y

##
## file: drivers/thermal/tegra/Kconfig
##
CONFIG_TEGRA_SOCTHERM=y

##
## file: drivers/tty/serial/Kconfig
##
CONFIG_SERIAL_AMBA_PL010=y
CONFIG_SERIAL_AMBA_PL010_CONSOLE=y
CONFIG_SERIAL_AMBA_PL011=y
CONFIG_SERIAL_AMBA_PL011_CONSOLE=y
CONFIG_SERIAL_MESON=y
CONFIG_SERIAL_MESON_CONSOLE=y
CONFIG_SERIAL_TEGRA=y
CONFIG_SERIAL_MSM=y
CONFIG_SERIAL_MSM_CONSOLE=y
CONFIG_SERIAL_MVEBU_UART=y
CONFIG_SERIAL_MVEBU_CONSOLE=y

##
## file: drivers/tty/serial/8250/Kconfig
##
CONFIG_SERIAL_8250=y
CONFIG_SERIAL_8250_DEPRECATED_OPTIONS=y
CONFIG_SERIAL_8250_CONSOLE=y
CONFIG_SERIAL_8250_DMA=y
CONFIG_SERIAL_8250_NR_UARTS=4
CONFIG_SERIAL_8250_RUNTIME_UARTS=4
CONFIG_SERIAL_8250_EXTENDED=y
CONFIG_SERIAL_8250_SHARE_IRQ=y
CONFIG_SERIAL_8250_BCM2835AUX=y
CONFIG_SERIAL_8250_DW=y
# CONFIG_SERIAL_8250_EM is not set
CONFIG_SERIAL_OF_PLATFORM=y

##
## file: drivers/usb/chipidea/Kconfig
##
CONFIG_USB_CHIPIDEA=y
CONFIG_USB_CHIPIDEA_UDC=y
CONFIG_USB_CHIPIDEA_HOST=y

##
## file: drivers/usb/dwc2/Kconfig
##
CONFIG_USB_DWC2=y
## choice: DWC2 Mode Selection
CONFIG_USB_DWC2_DUAL_ROLE=y
## end choice

##
## file: drivers/usb/dwc3/Kconfig
##
CONFIG_USB_DWC3=y
## choice: DWC3 Mode Selection
CONFIG_USB_DWC3_DUAL_ROLE=y
## end choice

##
## file: drivers/usb/gadget/Kconfig
##
CONFIG_USB_GADGET=y

##
## file: drivers/usb/host/Kconfig
##
#. xhci-platform apparently does not build as module, so xhci_hcd can't be either
CONFIG_USB_XHCI_HCD=y
CONFIG_USB_XHCI_PLATFORM=y
CONFIG_USB_XHCI_TEGRA=y
CONFIG_USB_EHCI_HCD=y
CONFIG_USB_EHCI_MSM=y
CONFIG_USB_EHCI_TEGRA=y
CONFIG_USB_EHCI_HCD_PLATFORM=y
CONFIG_USB_OHCI_HCD=y
CONFIG_USB_OHCI_HCD_PLATFORM=y

##
## file: drivers/usb/misc/Kconfig
##
CONFIG_USB_HSIC_USB3503=y

##
## file: drivers/usb/phy/Kconfig
##
CONFIG_USB_MSM_OTG=y
CONFIG_USB_QCOM_8X16_PHY=y

##
## file: drivers/video/backlight/Kconfig
##
CONFIG_BACKLIGHT_GENERIC=y
CONFIG_BACKLIGHT_LP855X=y

##
## file: drivers/video/fbdev/Kconfig
##
CONFIG_FB_EFI=y

##
## file: drivers/virtio/Kconfig
##
CONFIG_VIRTIO_MMIO=y

##
## file: drivers/watchdog/Kconfig
##
CONFIG_TEGRA_WATCHDOG=y
CONFIG_QCOM_WDT=y
CONFIG_MESON_GXBB_WATCHDOG=y
CONFIG_MESON_WATCHDOG=y
CONFIG_BCM2835_WDT=y

##
## file: fs/pstore/Kconfig
##
CONFIG_PSTORE=y

##
## file: net/bluetooth/Kconfig
##
CONFIG_BT_LEDS=y

##
## file: sound/pci/hda/Kconfig
##
CONFIG_SND_HDA_TEGRA=y

##
## file: sound/soc/Kconfig
##
CONFIG_SND_SOC=y

##
## file: sound/soc/bcm/Kconfig
##
CONFIG_SND_BCM2835_SOC_I2S=y

##
## file: sound/soc/qcom/Kconfig
##
CONFIG_SND_SOC_QCOM=y
CONFIG_SND_SOC_APQ8016_SBC=y

##
## file: sound/soc/tegra/Kconfig
##
CONFIG_SND_SOC_TEGRA=y
CONFIG_SND_SOC_TEGRA_RT5640=y
CONFIG_SND_SOC_TEGRA_WM8753=y
CONFIG_SND_SOC_TEGRA_WM8903=y
CONFIG_SND_SOC_TEGRA_TRIMSLICE=y
CONFIG_SND_SOC_TEGRA_ALC5632=y
CONFIG_SND_SOC_TEGRA_MAX98090=y
CONFIG_SND_SOC_TEGRA_RT5677=y

`

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

	f, err := os.OpenFile(".config", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write([]byte(configAddendum)); err != nil {
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

	make := exec.Command("make", "Image.gz", "dtbs", "-j8")
	make.Env = append(os.Environ(),
		"ARCH=arm64",
		"CROSS_COMPILE=aarch64-linux-gnu-",
		"KBUILD_BUILD_USER=gokrazy",
		"KBUILD_BUILD_HOST=docker",
		"KBUILD_BUILD_TIMESTAMP=Wed Mar  1 20:57:29 UTC 2017",
	)
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
	if err := exec.Command("tar", "xf", filepath.Base(latest)).Run(); err != nil {
		log.Fatal("untar: %v", err)
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

	if err := copyFile("/tmp/buildresult/rpi-3-b.dtb", "arch/arm64/boot/dts/broadcom/bcm2837-rpi-3-b.dtb"); err != nil {
		log.Fatal(err)
	}
}
