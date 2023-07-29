export BASEDIR := $(abspath ./)

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

build-bpf-program:
	$(call FOREACH_BUILD_BPF, bpf2go)