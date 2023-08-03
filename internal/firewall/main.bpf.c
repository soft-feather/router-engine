//+build ignore
#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_endian.h>
#ifdef asm_inline
#undef asm_inline
#define asm_inline asm
#endif

#define MAX_CPU 128
#define ETH_P_IP        0x0800

// 存储对应网口对应的sock fd
struct {
    __uint(type, BPF_MAP_TYPE_XSKMAP);
    __uint(max_entries, MAX_CPU);
    __uint(key_size, sizeof(u32));
    __uint(value_size, sizeof(u32));
} xsk_maps SEC(".maps");

// 存储对应网口是否需要使用af_xdp
struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, MAX_CPU);
    __uint(key_size, sizeof(u32));
    __uint(value_size, sizeof(u32));
} index_stat SEC(".maps");

static __always_inline __u16 csum_fold_helper(__u64 csum)
{
int i;
#pragma unroll
for (i = 0; i < 4; i++)
{
if (csum >> 16)
csum = (csum & 0xffff) + (csum >> 16);
}
return ~csum;
}

static __always_inline __u16 ipv4_csum(struct iphdr *iph)
{
    iph->check = 0;
    unsigned long long csum = bpf_csum_diff(0, 0, (unsigned int *)iph, sizeof(struct iphdr), 0);
    return csum_fold_helper(csum);
}

struct pkt_meta {
    union {
        __be32 src;
        __be32 srcv6[4];
    };
    union {
        __be32 dst;
        __be32 dstv6[4];
    };
    __u16 port16[2];
    __u16 l3_proto;
    __u16 l4_proto;
    __u16 data_len;
    __u16 pkt_len;
    __u32 seq;
};

static __always_inline bool parse_tcp(struct tcphdr *tcp,
                                      struct pkt_meta *pkt)
{
    pkt->port16[0] = tcp->source;
    pkt->port16[1] = tcp->dest;
    pkt->seq = tcp->seq;
    return true;
}

static __always_inline bool parse_ip4(struct iphdr *iph,
                                      struct pkt_meta *pkt)
{
    if (iph->ihl != 5)
        return false;
    pkt->src = iph->saddr;
    pkt->dst = iph->daddr;
    pkt->l4_proto = iph->protocol;
    return true;
}

SEC("xdp")
int xdp_prog(struct xdp_md *ctx)
{
    int index = ctx->rx_queue_index;
    if (!bpf_map_lookup_elem(&xsk_maps, &index)) {
        const int ret_val = bpf_redirect_map(&xsk_maps, index, 0);
        bpf_printk("RET-VAL: %d\n", ret_val);
        return ret_val;
    }
    void *data_end = (void *)(long)ctx->data_end;
    void *data = (void *)(long)ctx->data;
    struct ethhdr *eth = data;
    struct tcphdr *tcp;
    struct iphdr *iph;
    struct pkt_meta pkt = {};
    __u32 off;

    /* parse packet for IP Addresses and Ports */
    off = sizeof(struct ethhdr);
    if (data + off > data_end)
        return XDP_ABORTED;

    pkt.l3_proto = bpf_htons(eth->h_proto);

    if (pkt.l3_proto == ETH_P_IP) {
        iph = data + off;
        if ((void *)(iph + 1) > data_end)
            return XDP_ABORTED;
        if (!parse_ip4(iph, &pkt))
            return XDP_PASS;
        off += sizeof(struct iphdr);
    }
    if (data + off > data_end)
        return XDP_ABORTED;

    if (pkt.l4_proto == IPPROTO_ICMP) {
        const int ret_val = bpf_redirect_map(&xsk_maps, index, 0);
        return ret_val;
    }
    return XDP_PASS;
}
char LICENSE[] SEC("license") = "GPL";