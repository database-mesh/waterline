#include <linux/bpf.h>
#include <linux/pkt_cls.h>
#include <linux/ip.h>
#include <linux/tcp.h>
#include <arpa/inet.h>
#include <linux/if_ether.h>
#include "headers/bpf_helpers.h"
#include "headers/bpf_endian.h"

// bpf_elf map definition from https://github.com/shemminger/iproute2/blob/main/include/bpf_elf.h
struct bpf_elf_map {
    unsigned int type;
    unsigned int size_key;
    unsigned int size_value;
    unsigned int max_elem;
    unsigned int flags;
    unsigned int id;
    unsigned int pinning;
    unsigned int inner_id;
    unsigned int inner_idx;
};

struct event_key {
    __u8 seq;
    __u16 sport;
    __u16 dport;
    __u32 saddr;
    __u32 daddr;
};

struct event_value {
	__u32 payload_offset;
	__u32 my_pkt_len;
	__u32 class_id;
};

struct bpf_elf_map SEC("maps") my_pkt = {
	.type           = BPF_MAP_TYPE_HASH,
	.size_key       = sizeof(struct event_key),
	.size_value     = sizeof(struct event_value),
	// pin path default is /sys/fs/bpf/tc/globals/my_pkt
	// waterline can write qos rule to my_pkt by pin path
	.pinning        = 2,
};

// attach to eth0 || cni0 || docker0
SEC("classifier")
int tc_egress(struct __sk_buff *skb) {
    if (skb->protocol != bpf_htons(ETH_P_IP)) {
        return TC_ACT_OK;
    }

    struct iphdr iph;
    bpf_skb_load_bytes(skb, ETH_HLEN, &iph, sizeof(iph));

    if (iph.protocol != IPPROTO_TCP) {
    	return TC_ACT_OK;
    }

    //__u32 ip_hlen = iph.ihl << 2;

    struct tcphdr tcph;

    bpf_skb_load_bytes(skb, ETH_HLEN + sizeof(iph), &tcph, sizeof(tcph));

    //__u32 tcp_hlen = tcph.doff << 2;

    struct event_key evt_key;
	__builtin_memset(&evt_key, 0, sizeof(struct event_key));
	evt_key.saddr = iph.daddr;
	evt_key.sport = tcph.dest;
	evt_key.daddr = iph.saddr;
	evt_key.dport = tcph.source;
	evt_key.seq = 0;

    struct event_value *evt_value = bpf_map_lookup_elem(&my_pkt, &evt_key);
    if (!evt_value) return TC_ACT_OK;

    __u32 payload_offset = evt_value->payload_offset;
    __u8 header[4];
    bpf_skb_load_bytes(skb, payload_offset, &header, sizeof(header));
    if (header[3] < 1) return TC_ACT_OK;

    skb->tc_classid = evt_value->class_id;
    return TC_ACT_OK;
}

char LICENSE[] SEC("license") = "GPL";