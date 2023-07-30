export BASEDIR := $(abspath ./)

GIT := $(shell which git || /bin/false)

LIBBPF_PATH = $(BASEDIR)/third_party/libbpf
LIBBPF_SRC = $(abspath ./$(LIBBPF_PATH)/src)
LIBBPF_OBJ = $(abspath ./$(LIBBPF_PATH)/lib/libbpf.a)

FIREWALLDIR = $(BASEDIR)/internal/firewall
BPFPROGEAMDIR = $(FIREWALLDIR)

define FOREACH_BUILD_BPF
	BUILDERR=0; \
	for DIR in $(BPFPROGEAMDIR); do \
		$(MAKE) -j1 -C $$DIR $(1) || BUILDERR=1; \
	done; \
	if [ $$BUILDERR -eq 1 ]; then \
		exit 1; \
	fi
endef

.PHONY:test
test:
	$(call FOREACH_BUILD_BPF, test)

.PHONY: build-bpf-program
build-bpf-program: $(LIBBPF_OBJ)
	$(call FOREACH_BUILD_BPF, bpf2go)

$(LIBBPF_OBJ): $(LIBBPF_SRC) $(wildcard $(LIBBPF_SRC)/*.[ch])
	CC="gcc" CFLAGS="-g -O2 -Wall -fpie" LD_FLAGS="" \
       /usr/bin/make -C $(LIBBPF_PATH)/src \
    	BUILD_STATIC_ONLY=1 \
    	OBJDIR=$(LIBBPF_PATH) \
    	DESTDIR=$(LIBBPF_PATH)/lib \
    	INCLUDEDIR= LIBDIR= UAPIDIR= install

$(LIBBPF_SRC):
ifeq ($(wildcard $@), )
	$(GIT) submodule update --init --recursive
endif