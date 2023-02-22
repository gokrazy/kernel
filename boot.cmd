setenv bootargs "console=ttySAC2,115200n8 root=/dev/mmcblk0p2 rootwait panic=10 oops=panic init=/gokrazy/init"

fatload mmc 2:1 0x40008000 vmlinuz

fatload mmc 2:1 0x42000000 cmdline.txt

env import -t 0x42000000 ${filesize}

fatload mmc 2:1 0x44000000 sun50i-h6-tanix-tx6.dtb

fdt addr 0x44000000

bootz 0x40008000 - 0x44000000