/*
Copyright 2021 Wim Henderickx.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package parser

import (
	"strings"

	config "github.com/netw-device-driver/ndd-grpc/config/configpb"
	"github.com/openconfig/goyang/pkg/yang"
)

// ProcessLeafRef processes the leafref and returns if a leafref localPath, remotePath and if the leafRef is local or external to the resource
// used for yang parser
func (p *Parser) ProcessLeafRef(e *yang.Entry, resfullPath string, activeResPath *config.Path) (*config.Path, *config.Path, bool) {
	switch p.GetTypeName(e) {
	default:
		switch p.GetTypeKind(e) {
		case "leafref":
			//fmt.Println(e.Node.Statement().String())
			splitData := strings.Split(e.Node.Statement().String(), "\n")
			var path string
			var elem string
			var k string
			for _, s := range splitData {
				if strings.Contains(s, "path ") {
					// strip the junk from the leafref to get a plain xpath
					//fmt.Printf("LeafRef Path: %s\n", s)
					s = strings.ReplaceAll(s, "path ", "")
					s = strings.ReplaceAll(s, ";", "")
					s = strings.ReplaceAll(s, "\"", "")
					s = strings.ReplaceAll(s, " ", "")
					s = strings.ReplaceAll(s, "\t", "")
					//fmt.Printf("LeafRef Path: %s\n", s)

					// split the leafref per "/" and split the element and key from the path
					// last element is the key
					// 2nd last element is the element
					split2data := strings.Split(s, "/")
					//fmt.Printf("leafRef Len Split2 %d\n", len(split2data))

					for i, s2 := range split2data {
						switch i {
						case 0: // the first element in the leafref split is typically "", since the string before the "/" is empty
							if s2 != "" { // if not empty ensure we use the right data and split the string before ":" sign
								path += "/" + strings.Split(s2, ":")[len(strings.Split(s2, ":"))-1]

							}
						case (len(split2data) - 1): // last element is the key
							k = strings.Split(s2, ":")[len(strings.Split(s2, ":"))-1]
						case (len(split2data) - 2): // 2nd last element is the element
							elem = strings.Split(s2, ":")[len(strings.Split(s2, ":"))-1]
						default: // any other element gets added to the list
							path += "/" + strings.Split(s2, ":")[len(strings.Split(s2, ":"))-1]

						}
					}
					// if no path element exits we take the root "/" path
					if path == "" {
						path = "/"
					}
					// if the path contains /.. this is a relative leafref path
					relativeIndex := strings.Count(path, "/..")
					if relativeIndex > 0 {
						//fmt.Printf("leafRef Relative Path: %s, Element: %s, Key: %s, '/..' count %d\n", path, elem, k, relativeIndex)
						// check if the final p contains relative indirection to the resourcePath -> "/.."
						resSplitData := strings.Split(p.RemoveFirstEntry(resfullPath), "/")
						//fmt.Printf("ResPath Split Length: %d data: %v\n", len(resSplitData), resSplitData)
						var addString string
						for i := 1; i <= (len(resSplitData) - 1 - strings.Count(path, "/..")); i++ {
							addString += "/" + resSplitData[i]
						}
						//fmt.Printf("leafRef Absolute Path Add string: %s\n", addString)
						path = addString + strings.ReplaceAll(path, "/..", "")
					}
					//fmt.Printf("leafRef Absolute Path: %s, Element: %v, Key: %s, '/..' count %d\n", path, e, k, relativeIndex)

				}
			}
			//fmt.Printf("Path: %s, Elem: %s, Key: %s\n", path, elem, k)
			remotePath := p.XpathToConfigGnmiPath(path, 0)
			p.AppendElemInPath(remotePath, elem, k)

			// build a gnmi path and remove the first entry since the yang contains a duplicate path
			localPath := p.XpathToConfigGnmiPath(resfullPath, 1)
			// the last element hould be a key in the previous element
			//localPath = TransformPathToLeafRefPath(localPath)

			if strings.Contains(*p.ConfigGnmiPathToXPath(remotePath, false), *p.ConfigGnmiPathToXPath(activeResPath, false)) {
				// if the remotePath and the active Path match exactly we classify this in the external leafref category
				// since we dont allow multiple elments of the same key in the same resource
				// E.g. interface ethernet-1/1 which reference a lag should be resolve to another interface in
				// another resource and hence this should be classified as an external leafref
				if *p.ConfigGnmiPathToXPath(remotePath, false) != *p.ConfigGnmiPathToXPath(activeResPath, false) {
					// this is a local leafref within the resource
					// make the localPath and remotePath relative to the resource
					//fmt.Printf("localPath: %v, remotePath %v, activePath %v\n", localPath, remotePath, activeResPath)
					localPath = p.TransformPathAsRelative2Resource(localPath, activeResPath)
					remotePath = p.TransformPathAsRelative2Resource(remotePath, activeResPath)
					//fmt.Printf("localPath: %v, remotePath %v\n", localPath, remotePath)
					return localPath, remotePath, true
				}

			}
			// leafref is external to the resource
			//fmt.Printf("localPath: %v, remotePath %v, activePath %v\n", localPath, remotePath, activeResPath)
			// make the localPath relative to the resource
			localPath = p.TransformPathAsRelative2Resource(localPath, activeResPath)
			//fmt.Printf("localPath: %v, remotePath %v\n", localPath, remotePath)

			return localPath, remotePath, false
		}
	}
	return nil, nil, false
}
