FIREWALLDIR = $(abspath ./)
BTFFILE = /sys/kernel/btf/vmlinux
BPFTOOL = $(shell which bpftool)
ARCH := $(shell uname -m | sed 's/x86_64/amd64/g; s/aarch64/arm64/g')
CBPFFILENAME = main.bpf.o
VMLINUXH = ./vmlinux.h
LIBBPFPATH = $(BASEDIR)/libbpfso


one-make:
	make c2bpf
	make bpf2go

test:
	echo $(FIREWALLDIR)

bpf2go:
	CC=clang \
		CGO_CFLAGS="-I$(LIBBPFPATH)/" \
		CGO_LDFLAGS="-lelf -lz $(LIBBPFPATH)/libbpf.a" \
		GOOS=linux GOARCH=$(ARCH) \
		go build \
		-tags netgo -ldflags '-w -extldflags "-static"' \
		-o rss_stat \
		./main.go

c2bpf:
	clang -g -O2 -c -target bpf -o $(CBPFFILENAME) main.bpf.c

con-vmlinux:
	@if [ ! -f $(BTFFILE) ]; then \
		echo "ERROR: kernel does not seem to support BTF"; \
		exit 1; \
    fi
	# 获取 uname -r 输出的版本号
	kernel_version=$(uname -r)

    # 使用 awk 命令提取主版本号和次版本号
	major_version=$(echo "$kernel_version" | awk -F '.' '{print $1}')
	minor_version=$(echo "$kernel_version" | awk -F '.' '{print $2}')

    # 判断是否小于 5.8
	if [ "$major_version" -lt 5 ] || ([ "$major_version" -eq 5 ] && [ "$minor_version" -lt 8 ]); then
		echo "Kernel version is less than 5.8"
	else
		echo "Kernel version is equal to or greater than 5.8"
	fi

	@if [ ! -f $(VMLINUXH) ]; then \
		if [ ! $(BPFTOOL) ]; then \
			echo "ERROR: could not find bpftool, but can be installed with: \n"; \
			     "	  sudo apt install linux-tools-generic"; \
			exit 1; \
		fi; \
		echo "INFO: generating $(VMLINUXH) from $(BTFFILE)"; \
		$(BPFTOOL) btf dump file $(BTFFILE) format c > $(VMLINUXH); \
	fi
