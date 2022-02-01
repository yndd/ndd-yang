package cache

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/openconfig/gnmi/ctree"
	"github.com/openconfig/gnmi/path"
	"github.com/openconfig/gnmi/proto/gnmi"
)

//{"level":"debug","ts":1633674399.5347052,"logger":"ipam","msg":"Create Fine Grane Updates","resource":"ipam-default-ipprefix-isl-ipv4","Resource":"ipam-default-ipprefix-isl-ipv4","Path":"/ipam/tenant[name=default]/network-instance[name=default]/ip-prefix[prefix=100.64.0.0/16]","Value":"json_ietf_val:\"{\\\"address-allocation-strategy\\\":\\\"first-address\\\",\\\"admin-state\\\":\\\"enable\\\"}\""}
//{"level":"debug","ts":1633674399.5347154,"logger":"ipam","msg":"Create Fine Grane Updates","resource":"ipam-default-ipprefix-isl-ipv4","Resource":"ipam-default-ipprefix-isl-ipv4","Path":"/ipam/tenant[name=default]/network-instance[name=default]/ip-prefix[prefix=100.64.0.0/16]/tag[key=purpose]","Value":"json_ietf_val:\"{\\\"value\\\":\\\"isl\\\"}\""}

func Callback(n *ctree.Leaf) {
	switch v := n.Value().(type) {
	case *gnmi.Notification:
		fmt.Printf("Cache change notification, Alias: %v, Prefix: %v, Path: %v, Value: %v\n", v.GetAlias(), path.ToStrings(v.GetPrefix(), true), path.ToStrings(v.GetUpdate()[0].GetPath(), true), v.GetUpdate()[0].GetVal())
	default:
		fmt.Printf("State CacheUpdates unexpected type: %v\n", reflect.TypeOf(n.Value()))
	}
}

func TestDynamicUpdates(t *testing.T) {
	target := "dev1"
	origin := "test"
	prefix := &gnmi.Path{
		Target: target,
		Origin: origin,
	}
	tests := []struct {
		inp *gnmi.Notification
		exp interface{}
	}{
		{
			inp: &gnmi.Notification{
				Prefix: prefix,
				Alias:  "config",
				Update: []*gnmi.Update{
					{
						Path: &gnmi.Path{
							Elem: []*gnmi.PathElem{
								{Name: "ipam"},
								{Name: "rir", Key: map[string]string{"name": "rfc1918"}},
							},
						},
						Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "rfc1918"}},
					},
				},
			},
		},
		{
			inp: &gnmi.Notification{
				Prefix: prefix,
				Alias:  "Config",
				Update: []*gnmi.Update{
					{
						Path: &gnmi.Path{
							Elem: []*gnmi.PathElem{
								{Name: "ipam"},
								{Name: "rir", Key: map[string]string{"name": "rfc1918"}},
							},
						},
						Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "rfc1918"}},
					},
				},
			},
		},
		{
			inp: &gnmi.Notification{
				Prefix: prefix,
				Alias:  "Config",
				Update: []*gnmi.Update{
					{
						Path: &gnmi.Path{
							Elem: []*gnmi.PathElem{
								{Name: "ipam"},
								{Name: "rir", Key: map[string]string{"name": "rfc1918"}},
							},
						},
						Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "rfc1918"}},
					},
				},
			},
		},
		{
			inp: &gnmi.Notification{
				Prefix: prefix,
				Alias:  "Config",
				Update: []*gnmi.Update{
					{
						Path: &gnmi.Path{
							Elem: []*gnmi.PathElem{
								{Name: "ipam"},
								{Name: "rir", Key: map[string]string{"name": "rfc1918"}},
							},
						},
						Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "rfc1918"}},
					},
				},
			},
		},
	}

	c := New([]string{target})
	c.GetCache().SetClient(Callback)
	for i, tt := range tests {
		fmt.Printf("##### TestGetNotificationFromUpdate Nbr: %d ###### \n", i)
		tt.inp.Timestamp = time.Now().UnixNano()
		if err := c.GnmiUpdate(target, tt.inp); err != nil {
			t.Errorf("GetNotificationFromUpdate: %v\n", err)
		}
	}
}

/*
func TestGetNotificationFromUpdate(t *testing.T) {

	target := "dev1"
	origin := "test"
	prefix := &gnmi.Path{
		Target: target,
		Origin: origin,
		//Elem:   []*gnmi.PathElem{{Name: "a"}, {Name: "b", Key: map[string]string{"key": "value"}}},
	}
	x1 := map[string]interface{}{"admin-state": "enable", "rir-name": "rfc1918"}
	b1, _ := json.Marshal(x1)
	x2 := map[string]interface{}{"address-allocation-strategy": "first-address", "admin-state": "enable"}
	b2, _ := json.Marshal(x2)
	x3 := map[string]interface{}{"value": "isl"}
	b3, _ := json.Marshal(x3)
	tests := []struct {
		inp *gnmi.Update
		exp interface{}
	}{
		{
			inp: &gnmi.Update{
				Path: &gnmi.Path{
					Elem: []*gnmi.PathElem{
						{Name: "ipam"},
						{Name: "aggregate", Key: map[string]string{"tenant": "default", "prefix": "10.0.0.0/8", "network-instance": "default"}},
					},
				},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonVal{JsonVal: b1}},
			},

			//exp: "ipam",
		},
		{
			inp: &gnmi.Update{
				Path: &gnmi.Path{
					Elem: []*gnmi.PathElem{
						{Name: "ipam"},
						{Name: "tenant", Key: map[string]string{"name": "default"}},
						{Name: "network-instance", Key: map[string]string{"name": "default"}},
						{Name: "ip-prefix", Key: map[string]string{"prefix": "1.1.1.1/24"}},
					},
				},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonVal{JsonVal: b2}},
			},

			//exp: "ipam",
		},
		{
			inp: &gnmi.Update{
				Path: &gnmi.Path{
					Elem: []*gnmi.PathElem{
						{Name: "ipam"},
						{Name: "tenant", Key: map[string]string{"name": "default"}},
						{Name: "network-instance", Key: map[string]string{"name": "default"}},
						{Name: "ip-prefix", Key: map[string]string{"prefix": "1.1.1.1/24"}},
						{Name: "tag", Key: map[string]string{"key": "purpose"}},
					},
				},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonVal{JsonVal: b3}},
			},

			//exp: "ipam",
		},
	}
	c := New([]string{target})

	for i, tt := range tests {
		n, err := c.GetNotificationFromUpdate(prefix, tt.inp)
		if err != nil {
			t.Errorf("GetNotificationFromUpdate: %v\n", err)
		}
		fmt.Printf("##### TestGetNotificationFromUpdate Nbr: %d ###### \n", i)
		for _, u := range n.GetUpdate() {
			fmt.Printf("Update: %v\n", u)
		}
	}

}
*/

func TestGetJson(t *testing.T) {
	target := "dev1"
	origin := "test"
	prefix := &gnmi.Path{
		Target: target,
		Origin: origin,
		//Elem:   []*gnmi.PathElem{{Name: "a"}, {Name: "b", Key: map[string]string{"key": "value"}}},
	}

	tests := []struct {
		inp *gnmi.Path
		exp interface{}
	}{
		{
			inp: &gnmi.Path{
				//Origin: origin,
				Elem: []*gnmi.PathElem{},
			},
			//exp: "ipam",
		},
		{
			inp: &gnmi.Path{
				//Origin: origin,
				Elem: []*gnmi.PathElem{
					{Name: "ipam"},
				},
			},
			//exp: "aggregate",
		},

		{
			inp: &gnmi.Path{
				//Origin: origin,
				Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "tenant", Key: map[string]string{"name": "default"}},
				},
			},
			//exp: map[string]interface {}{"Prefix":"10.0.0.0/8", "admin-state":"enable", "network-instance":"default", "rir-name":"rfc1918", "tenant":"default"},
		},
		{
			inp: &gnmi.Path{
				//Origin: origin,
				Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "tenant", Key: map[string]string{"name": "default"}},
					{Name: "network-instance", Key: map[string]string{"name": "default"}},
				},
			},
			//exp: map[string]interface {}{"Prefix":"10.0.0.0/8", "admin-state":"enable", "network-instance":"default", "rir-name":"rfc1918", "tenant":"default"},
		},
		{
			inp: &gnmi.Path{
				//Origin: origin,
				Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "tenant", Key: map[string]string{"name": "default"}},
					{Name: "network-instance", Key: map[string]string{"name": "default"}},
					{Name: "ip-prefix", Key: map[string]string{"prefix": "100.64.0.0/16"}},
				},
			},
			//exp: map[string]interface {}{"Prefix":"10.0.0.0/8", "admin-state":"enable", "network-instance":"default", "rir-name":"rfc1918", "tenant":"default"},
		},
	}

	n := &gnmi.Notification{
		Timestamp: time.Now().UnixNano(),
		Prefix:    prefix,
		Update: []*gnmi.Update{
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "rir", Key: map[string]string{"name": "rfc1918"}},
					{Name: "name"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "rfc1918"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "rir", Key: map[string]string{"name": "rfc6598"}},
					{Name: "name"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "rfc6598"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "rir", Key: map[string]string{"name": "ula"}},
					{Name: "name"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "ula"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "rir", Key: map[string]string{"name": "ripe"}},
					{Name: "name"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "ripe"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "tenant", Key: map[string]string{"name": "default"}},
					{Name: "name"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "default"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "tenant", Key: map[string]string{"name": "default"}},
					{Name: "admin-state"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "enable"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "tenant", Key: map[string]string{"name": "default"}},
					{Name: "network-instance", Key: map[string]string{"name": "default"}},
					{Name: "aname"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "default"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "tenant", Key: map[string]string{"name": "default"}},
					{Name: "network-instance", Key: map[string]string{"name": "default"}},
					{Name: "admin-state"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "enable"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "tenant", Key: map[string]string{"name": "default"}},
					{Name: "network-instance", Key: map[string]string{"name": "default"}},
					{Name: "address-allocation-strategy"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "first-address"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "tenant", Key: map[string]string{"name": "default"}},
					{Name: "network-instance", Key: map[string]string{"name": "default"}},
					{Name: "ip-prefix", Key: map[string]string{"prefix": "100.64.0.0/16"}},
					{Name: "prefix"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "100.64.0.0/16"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "tenant", Key: map[string]string{"name": "default"}},
					{Name: "network-instance", Key: map[string]string{"name": "default"}},
					{Name: "ip-prefix", Key: map[string]string{"prefix": "100.64.0.0/16"}},
					{Name: "admin-state"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "enable"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "tenant", Key: map[string]string{"name": "default"}},
					{Name: "network-instance", Key: map[string]string{"name": "default"}},
					{Name: "ip-prefix", Key: map[string]string{"prefix": "100.64.0.0/16"}},
					{Name: "address-allocation-strategy"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "first-address"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "tenant", Key: map[string]string{"name": "default"}},
					{Name: "network-instance", Key: map[string]string{"name": "default"}},
					{Name: "ip-prefix", Key: map[string]string{"prefix": "100.64.0.0/16"}},
					{Name: "tag", Key: map[string]string{"key": "purpose"}},
					{Name: "key"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "purpose"}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{
					{Name: "ipam"},
					{Name: "tenant", Key: map[string]string{"name": "default"}},
					{Name: "network-instance", Key: map[string]string{"name": "default"}},
					{Name: "ip-prefix", Key: map[string]string{"prefix": "100.64.0.0/16"}},
					{Name: "tag", Key: map[string]string{"key": "purpose"}},
					{Name: "value"},
				}},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "isl"}},
			},
		},
	}

	c := New([]string{target})
	if err := c.GnmiUpdate(target, n); err != nil {
		t.Errorf("GnmiUpdate: %v\n", err)
	}

	for i, tt := range tests {
		fmt.Printf("##### TestGetJson Nbr: %d ###### \n", i)
		/*
			pp := path.ToStrings(tt.inp, true)
			fmt.Printf("pp %v\n", pp)
			c.GetCache().Query(target, pp, func(_ []string, _ *ctree.Leaf, val interface{}) error {
				fmt.Printf("val %v\n", val)
				return nil
			})
		*/

		d, err := c.GetJson(target, prefix, tt.inp, nil)
		if err != nil {
			t.Errorf("GetConfig: %v\n", err)
		}
		//fmt.Println(d)
		jsonString, _ := json.MarshalIndent(d, "", "\t")
		fmt.Println(string(jsonString))

		fmt.Printf("out1: %#v\n", d)
		fmt.Printf("out2: %#v\n", tt.exp)

		//if !reflect.DeepEqual(d, tt.exp) {
		//	t.Errorf("sortedVals(%v):\n got  %v\n want %v\n", tt.inp, d, tt.exp)
		//}
	}
}
