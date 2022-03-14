#include <linux/bpf.h>
#include <linux/pkt_cls.h>
#include <linux/ip.h>
#include <linux/tcp.h>
#include <arpa/inet.h>
#include <linux/if_ether.h>
#include "include/bpf_helpers.h"
#include "include/bpf_endian.h"

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

struct event {
    __u8 seq;
    __u16 sport;
    __u16 dport;
    __u32 saddr;
    __u32 daddr;
    __u32 my_pkt_len;
    __u32 payload_offset;
    __u32 class_id;
};

struct bpf_elf_map SEC("maps") my_pkt = {
	.type           = BPF_MAP_TYPE_HASH,
	.size_key       = sizeof(__u8),
	.size_value     = sizeof(struct event),
	.max_entries    = 1,
	// pin path default is /sys/fs/bpf/tc/globals/my_pkt
	// waterline can write qos rule to my_pkt by pin path
	.pinning        = 2,
};

SEC("classifier")
int tc_egress(struct __sk_buff *skb) {
    if (skb->protocol != bpf_htons(ETH_P_IP)) {
        return TC_ACT_OK;
    }

    struct event *evt = bpf_map_lookup_elem(&my_pkt, &evt_key);
    if (!evt) return TC_ACT_OK;

    if (iph.saddr != evt.daddr || iph.daddr != evt.saddr
        || tcph.source != evt.dest || tcph.dest != evt.source) {

    	return TC_ACT_OK;
    }

    __u32 payload_offset = evt.payload_offset;
    __u8 header[4];
    bpf_skb_load_bytes(skb, payload_offset, &header, sizeof(header));
    if (header[3] <= 1) return TC_ACT_OK;

    skb->tc_classid = evt.class_id;
    return TC_ACT_OK;
}

char LICENSE[] SEC("license") = "GPL";