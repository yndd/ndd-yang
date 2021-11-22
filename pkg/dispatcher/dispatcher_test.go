package dispatcher

import (
	"fmt"
	"testing"

	"github.com/openconfig/gnmi/path"
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/yndd/ndd-yang/pkg/yparser"
)

func TestAdd(t *testing.T) {
	fmt.Println("TestDispatcher")
	
	resourcePaths := []*gnmi.Path{
		
		{
			Elem: []*gnmi.PathElem{
				{Name: "ipam"},
				{Name: "tenant", Key: map[string]string{"name": "*"}},
			},
		},
		{
			Elem: []*gnmi.PathElem{
				{Name: "ipam"},
				{Name: "tenant", Key: map[string]string{"name": "*"}},
				{Name: "network-instance", Key: map[string]string{"name": "*"}},
			},
		},
		{
			Elem: []*gnmi.PathElem{
				{Name: "ipam"},
				{Name: "tenant", Key: map[string]string{"name": "*"}},
				{Name: "network-instance", Key: map[string]string{"name": "*"}},
				{Name: "ip-address", Key: map[string]string{"address": "*"}},
			},
		},
		{
			Elem: []*gnmi.PathElem{
				{Name: "ipam"},
				{Name: "tenant", Key: map[string]string{"name": "*"}},
				{Name: "network-instance", Key: map[string]string{"name": "*"}},
				{Name: "ip-prefix", Key: map[string]string{"prefix": "*"}},
			},
		},
		{
			Elem: []*gnmi.PathElem{
				{Name: "ipam"},
				{Name: "tenant", Key: map[string]string{"name": "*"}},
				{Name: "network-instance", Key: map[string]string{"name": "*"}},
				{Name: "ip-range", Key: map[string]string{"end": "*", "start": "*"}},
			},
		},
		{
			Elem: []*gnmi.PathElem{
				{Name: "ipam"},
			},
		},
		
	}

	testPaths := []*gnmi.Path{
		{
			Elem: []*gnmi.PathElem{
				{Name: "ipam"},
				{Name: "tenant", Key: map[string]string{"name": "default"}},
			},
		},
		{
			Elem: []*gnmi.PathElem{
				{Name: "ipam"},
				{Name: "tenant", Key: map[string]string{"name": "wim"}},
				{Name: "network-instance", Key: map[string]string{"name": "ni-wim2"}},
			},
		},
		{
			Elem: []*gnmi.PathElem{
				{Name: "ipam"},
				{Name: "tenant", Key: map[string]string{"name": "default"}},
				{Name: "network-instance", Key: map[string]string{"name": "default"}},
				{Name: "ip-prefix", Key: map[string]string{"prefix": "100.64.0.0/16"}},
			},
		},
		{
			Elem: []*gnmi.PathElem{
				{Name: "ipam"},
				{Name: "tenant", Key: map[string]string{"name": "default"}},
				{Name: "network-instance", Key: map[string]string{"name": "default"}},
				{Name: "ip-prefix", Key: map[string]string{"prefix": "100.64.0.0/16"}},
				{Name: "tag", Key: map[string]string{"key": "purpose"}},
			},
		},
		{
			Elem: []*gnmi.PathElem{
				{Name: "ipam"},
			},
		},
	}
	d := New()
	d.Init(resourcePaths)
	d.ShowTree()
	for _, p := range testPaths {
		pe := d.GetPathElem(p)
		fmt.Printf("PathElem: %v\n", pe)
		key, path := getPath2Process(p, pe)
		fmt.Printf("Key: %s, Path: %s\n", key, yparser.GnmiPath2XPath(path, true))
	}

}


// getPath2Process resolves the keys in the pathElem
// returns the resolved path based on the pathElem returned from lpm cache lookup
// returns the key which is using path.Strings where each element in the path.Strings
// is delineated by a .
func getPath2Process(p *gnmi.Path, pe []*gnmi.PathElem) (string, *gnmi.Path) {
	newPathElem := make([]*gnmi.PathElem, 0)
	var key string
	for i, elem := range pe {
		e := &gnmi.PathElem{
			Name: elem.GetName(),
		}
		if len(p.GetElem()[i].GetKey()) != 0 {
			e.Key = make(map[string]string)
			for keyName, keyValue := range p.GetElem()[i].GetKey() {
				e.Key[keyName] = keyValue
			}
		}
		newPathElem = append(newPathElem, e)
	}
	newPath := &gnmi.Path{Elem: newPathElem}
	stringlist := path.ToStrings(p, false)[:len(path.ToStrings(newPath, false))]
	for _, s := range stringlist {
		key = s + "."
	}
	return key, newPath
}
