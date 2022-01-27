package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/openconfig/gnmi/cache"
	"github.com/openconfig/gnmi/ctree"
	"github.com/openconfig/gnmi/path"
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/ndd-yang/pkg/parser"
	"github.com/yndd/ndd-yang/pkg/yentry"
	"github.com/yndd/ndd-yang/pkg/yparser"
)

type Cache struct {
	c   *cache.Cache
	p   *parser.Parser
	log logging.Logger
}

// Option can be used to manipulate Options.
type Option func(c *Cache)

func WithLogging(l logging.Logger) Option {
	return func(c *Cache) {
		c.log = l
	}
}

func WithParser(l logging.Logger) Option {
	return func(c *Cache) {
		c.p = parser.NewParser(parser.WithLogger(l))
	}
}

func New(t []string, opts ...Option) *Cache {
	c := &Cache{
		c: cache.New(t),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Cache) GetCache() *cache.Cache {
	return c.c
}

func (c *Cache) GnmiUpdate(t string, n *gnmi.Notification) error {
	return c.GetCache().GetTarget(t).GnmiUpdate(n)
}

// GetNotificationFromJson provides fine granular notifications from a JSON blob
func (c *Cache) GetNotificationFromJSON2(prefix *gnmi.Path, p *gnmi.Path, val interface{}, rs *yentry.Entry) (*gnmi.Notification, error) {
	c.log.Debug("GetNotificationFromJSON2", "Path", yparser.GnmiPath2XPath(p, true), "Value", val)
	updates := make([]*gnmi.Update, 0)
	var err error
	updates, err = c.getNotificationFromJSON2(p, val, updates, rs)
	if err != nil {
		return nil, err
	}
	return &gnmi.Notification{
		Timestamp: time.Now().UnixNano(),
		Prefix:    prefix,
		Update:    updates,
	}, nil
}

func (c *Cache) getNotificationFromJSON2(path *gnmi.Path, val interface{}, u []*gnmi.Update, rs *yentry.Entry) ([]*gnmi.Update, error) {
	var err error
	switch value := val.(type) {
	case nil:
		return u, nil
	case map[string]interface{}:
		// add the keys as data in the last element
		if len(path.GetElem()) != 0 {
			for k, v := range path.GetElem()[len(path.GetElem())-1].GetKey() {
				val, err := json.Marshal(v)
				if err != nil {
					return nil, err
				}
				update := &gnmi.Update{
					Path: &gnmi.Path{Elem: append(path.GetElem(), &gnmi.PathElem{Name: k})},
					Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonVal{JsonVal: val}},
				}
				u = append(u, update)
			}

			// add the values and add further processing
			for k, v := range value {
				switch value := v.(type) {
				case []interface{}:
					for _, v := range value {
						switch value := v.(type) {
						case map[string]interface{}:
							newPath := c.p.DeepCopyGnmiPath(path)
							// k = lastElem
							newPath = c.p.AppendElemInGnmiPath(newPath, k, nil)
							keys := rs.GetKeys(newPath)
							//keys := c.p.GetKeyNamesFromGnmiPaths(newPath, k, refPaths)
							pathKeys := make(map[string]string)
							if len(keys) != 0 {
								for _, key := range keys {
									pathKeys[key] = fmt.Sprintf("%v", value[key])
								}
								newPath = c.p.AppendElemInGnmiPathWithFullKey(path, k, pathKeys)
							} else {
								newPath = c.p.AppendElemInGnmiPath(path, k, nil)
							}

							// TODO expand keys
							u, err = c.getNotificationFromJSON2(newPath, v, u, rs)
							if err != nil {
								return nil, err
							}
						}
					}
				default:
					// this would be map[string]interface{}
					val, err := json.Marshal(v)
					if err != nil {
						return nil, err
					}
					update := &gnmi.Update{
						Path: &gnmi.Path{Elem: append(path.GetElem(), &gnmi.PathElem{Name: k})},
						Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonVal{JsonVal: val}},
					}
					u = append(u, update)
				}
			}
		}
	}
	return u, nil
}

// GetNotificationFromJson provides fine granular notifications from a JSON blob
func (c *Cache) GetNotificationFromJSON(prefix *gnmi.Path, p *gnmi.Path, val interface{}, refPaths []*gnmi.Path) (*gnmi.Notification, error) {
	updates := make([]*gnmi.Update, 0)
	var err error
	updates, err = c.getNotificationFromJSON(p, val, updates, refPaths)
	if err != nil {
		return nil, err
	}
	return &gnmi.Notification{
		Timestamp: time.Now().UnixNano(),
		Prefix:    prefix,
		Update:    updates,
	}, nil
}

func (c *Cache) getNotificationFromJSON(p *gnmi.Path, val interface{}, u []*gnmi.Update, refPaths []*gnmi.Path) ([]*gnmi.Update, error) {
	var err error
	switch value := val.(type) {
	case nil:
		return u, nil
	case map[string]interface{}:
		// add the keys as data in the last element
		for k, v := range p.GetElem()[len(p.GetElem())-1].GetKey() {
			val, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			p := c.p.DeepCopyGnmiPath(p)
			update := &gnmi.Update{
				Path: &gnmi.Path{Elem: append(p.GetElem(), &gnmi.PathElem{Name: k})},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonVal{JsonVal: val}},
			}
			u = append(u, update)
		}

		// add the values and add further processing
		for k, v := range value {
			switch value := v.(type) {
			case []interface{}:
				for _, v := range value {
					switch value := v.(type) {
					case map[string]interface{}:
						newPath := c.p.DeepCopyGnmiPath(p)
						keys := c.p.GetKeyNamesFromGnmiPaths(newPath, k, refPaths)
						pathKeys := make(map[string]string)
						if len(keys) != 0 {
							for _, key := range keys {
								pathKeys[key] = fmt.Sprintf("%v", value[key])
							}
							newPath = c.p.AppendElemInGnmiPathWithFullKey(newPath, k, pathKeys)
						} else {
							newPath = c.p.AppendElemInGnmiPath(newPath, k, nil)
						}

						// TODO expand keys
						u, err = c.getNotificationFromJSON(newPath, v, u, refPaths)
						if err != nil {
							return nil, err
						}
					}
				}
			default:
				// this would be map[string]interface{}
				val, err := json.Marshal(v)
				if err != nil {
					return nil, err
				}
				p := c.p.DeepCopyGnmiPath(p)
				update := &gnmi.Update{
					Path: &gnmi.Path{Elem: append(p.GetElem(), &gnmi.PathElem{Name: k})},
					Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonVal{JsonVal: val}},
				}
				u = append(u, update)
			}
		}
	}
	return u, nil
}

// GetNotificationFromUpdate provides fine granular notifications from the gnmi update by expanding the json blob value into
// inividual notifications.
func (c *Cache) GetNotificationFromUpdate(prefix *gnmi.Path, u *gnmi.Update) (*gnmi.Notification, error) {
	val, err := c.p.GetValue(u.GetVal())
	if err != nil {
		return nil, err
	}
	updates := []*gnmi.Update{}
	switch value := val.(type) {
	case nil:
		return nil, nil
	case map[string]interface{}:
		for k, v := range value {
			val, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			p := c.p.DeepCopyGnmiPath(u.GetPath())
			update := &gnmi.Update{
				Path: &gnmi.Path{Elem: append(p.GetElem(), &gnmi.PathElem{Name: k})},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonVal{JsonVal: val}},
			}
			updates = append(updates, update)
		}
		// add the keys as data in the last element
		for k, v := range u.Path.GetElem()[len(u.Path.GetElem())-1].GetKey() {
			val, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			p := c.p.DeepCopyGnmiPath(u.GetPath())
			update := &gnmi.Update{
				Path: &gnmi.Path{Elem: append(p.GetElem(), &gnmi.PathElem{Name: k})},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonVal{JsonVal: val}},
			}
			updates = append(updates, update)
		}

	default:
		updates = append(updates, u)
		fmt.Printf("Default Type: %v\n", reflect.TypeOf(val))
		for k, v := range u.Path.GetElem()[len(u.Path.GetElem())-1].GetKey() {
			val, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			update := &gnmi.Update{
				Path: &gnmi.Path{Elem: append(u.GetPath().GetElem(), &gnmi.PathElem{Name: k})},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_JsonVal{JsonVal: val}},
			}
			updates = append(updates, update)
		}
	}
	return &gnmi.Notification{
		Timestamp: time.Now().UnixNano(),
		Prefix:    prefix,
		Update:    updates,
	}, nil
}

func (c *Cache) GetNotificationFromDelete(prefix *gnmi.Path, p *gnmi.Path) (*gnmi.Notification, error) {
	return &gnmi.Notification{
		Timestamp: time.Now().UnixNano(),
		Prefix:    prefix,
		Delete:    []*gnmi.Path{{Elem: p.GetElem()}},
	}, nil

}

func (c *Cache) GetGnmiUpdateAsJsonBlob(t, o string, u *gnmi.Update) error {

	// to retrun the data we want to create a function that return the updates as a json blob iso individual path
	return nil
}

func (c *Cache) QueryAll(t string, prefix *gnmi.Path, p *gnmi.Path) ([]*gnmi.Notification, error) {
	notifications := []*gnmi.Notification{}
	fp, err := path.CompletePath(prefix, p)
	if err != nil {
		return nil, err
	}
	//pp := path.ToStrings(fp, true)
	if err := c.c.Query(t, fp,
		func(_ []string, _ *ctree.Leaf, n interface{}) error {
			if n, ok := n.(*gnmi.Notification); ok {
				notifications = append(notifications, n)
			}
			return nil
		}); err != nil {
		return nil, err
	}
	return notifications, nil
}

func (c *Cache) Query(t string, prefix *gnmi.Path, p *gnmi.Path) (*gnmi.Notification, error) {
	var notification *gnmi.Notification
	fp, err := path.CompletePath(prefix, p)
	if err != nil {
		return nil, err
	}
	//pp := path.ToStrings(fp, true)
	if err := c.c.Query(t, fp,
		func(_ []string, _ *ctree.Leaf, n interface{}) error {
			if n, ok := n.(*gnmi.Notification); ok {
				notification = n
			}
			return nil
		}); err != nil {
		return nil, err
	}
	return notification, nil
}

func (c *Cache) GetJson(t string, prefix *gnmi.Path, p *gnmi.Path) (interface{}, error) {
	var err error
	fp, err := path.CompletePath(prefix, p)
	if err != nil {
		return nil, err
	}
	var data interface{}
	//pp := path.ToStrings(p, true)
	if err := c.c.Query(t, fp,
		func(_ []string, _ *ctree.Leaf, n interface{}) error {
			if n, ok := n.(*gnmi.Notification); ok {
				for _, u := range n.GetUpdate() {
					//fmt.Printf("Notif: %v\n", u)
					// fp[2:]
					fmt.Printf("fp: %v\n", fp)
					if len(fp) < 2 {
						if data, err = c.addData(data, u.GetPath().GetElem(), fp, u.GetVal()); err != nil {
							return err
						}
					} else {
						if data, err = c.addData(data, u.GetPath().GetElem(), fp[1:], u.GetVal()); err != nil {
							return err
						}
					}

				}
			}
			return nil
		}); err != nil {
		return nil, err
	}
	return data, nil
}

func (c *Cache) addContainerValue(d interface{}, e string, val *gnmi.TypedValue) (interface{}, error) {
	var err error
	// check if the data was initialized
	if reflect.TypeOf((d)) == nil {
		d = make(map[string]interface{})
	}
	switch dd := d.(type) {
	case map[string]interface{}:
		// add the value to the element
		dd[e], err = c.p.GetValue(val)
		return d, err
	default:
		// we should never end up here
		return nil, errors.New("container last value JSON unexpected data structure")
	}
}

func (c *Cache) addListValue(d interface{}, e string, k map[string]string, val *gnmi.TypedValue) (interface{}, error) {
	var err error
	// initialize the data
	if reflect.TypeOf((d)) == nil {
		d = make(map[string]interface{})
	}
	switch dd := d.(type) {
	case map[string]interface{}:
		// initialize the data
		if _, ok := dd[e]; !ok {
			dd[e] = make([]interface{}, 0)
		}
		// create a container and initialize with keyNames/keyValues and value
		de := make(map[string]interface{})
		// add value
		de[e], err = c.p.GetValue(val)
		if err != nil {
			return nil, err
		}
		// add keyNames/keyValues
		for keyName, keyValue := range k {
			de[keyName] = keyValue
		}
		// add the data to the list
		switch l := dd[e].(type) {
		case []interface{}:
			dd[e] = append(l, de)
		default:
			return nil, errors.New("list last value JSON unexpected data structure")
		}
	}
	return d, nil
}

func (c *Cache) addContainer(d interface{}, e string, elems []*gnmi.PathElem, qelems []string, val *gnmi.TypedValue) (interface{}, error) {
	var err error
	// initialize the data
	//fmt.Printf("addContainer QueryPathElems: %v pathElem: %s\n", qelems, e)
	if len(qelems) > 0 && qelems[0] == e {
		// ignore the data
		d, err = c.addData(d, elems[1:], qelems[1:], val)
		return d, err
	} else {
		if reflect.TypeOf((d)) == nil {
			d = make(map[string]interface{})
		}
		switch dd := d.(type) {
		case map[string]interface{}:
			// add the container
			dd[e], err = c.addData(dd[e], elems[1:], qelems, val)
			return d, err
		default:
			return nil, errors.New("addListLastValue JSON unexpected data structure")
		}
	}

}

func (c *Cache) addList(d interface{}, e string, k map[string]string, elems []*gnmi.PathElem, qelems []string, val *gnmi.TypedValue) (interface{}, error) {
	var err error
	fmt.Printf("addList QueryPathElems: %v pathElem: %s, key: %v d: %v\n", qelems, e, k, d)
	// lean approach -> since we know the query should return paths that match the original query we can assume we match the path
	if len(qelems) > 1 {
		d, err = c.addData(d, elems[1:], qelems[1+len(k):], val)
		return d, err
	}
	// conservative approach
	/*
		if len(qelems) > 0 && qelems[0] == e {
			keys := make([]string, 0, len(k))
			for key := range k {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			found := true
			for i, key := range keys {
				if k[key] != qelems[1+i] {
					found = false
				}
			}
			if found {
				d, err = c.addData(d, elems[1:], qelems[1:])
				return d, err
			}
		}
	*/
	// initialize the data
	if reflect.TypeOf((d)) == nil {
		d = make(map[string]interface{})
	}
	switch dd := d.(type) {
	case map[string]interface{}:
		// initialize the data
		if _, ok := dd[e]; !ok {
			dd[e] = make([]interface{}, 0)
		}
		switch l := dd[e].(type) {
		case []interface{}:
			// check if the list entry exists
			for i, le := range l {
				// initialize the data
				if reflect.TypeOf((le)) == nil {
					le = make(map[string]interface{})
				}
				found := true
				switch dd := le.(type) {
				case map[string]interface{}:
					for keyName, keyValue := range k {
						if dd[keyName] != keyValue {
							found = false
						}
					}
					if found {
						// augment the list
						l[i], err = c.addData(dd, elems[1:], qelems, val)
						if err != nil {
							return nil, err
						}
						return d, err
					}
				}
			}
			// list entry not found, add a list entry
			de := make(map[string]interface{})
			for keyName, keyValue := range k {
				de[keyName] = keyValue
			}
			// augment the list
			x, err := c.addData(de, elems[1:], qelems, val)
			if err != nil {
				return nil, err
			}
			// add the list entry to the list
			dd[e] = append(l, x)
			return d, nil
		default:
			return nil, errors.New("list last value JSON unexpected data structure")
		}

	default:
		return nil, errors.New("list last value JSON unexpected data structure")
	}
}

func (c *Cache) addData(d interface{}, elems []*gnmi.PathElem, qelems []string, val *gnmi.TypedValue) (interface{}, error) {
	var err error
	e := elems[0].GetName()
	k := elems[0].GetKey()
	//fmt.Printf("addData, Len: %d, Elem: %s, Key: %v, QElems: %v, Data: %v\n", len(elems), e, k, qelems, d)
	if len(elems)-1 == 0 {
		// last element
		if len(k) == 0 {
			// last element with container
			d, err = c.addContainerValue(d, e, val)
			return d, err
		} else {
			// last element with list
			// not sure if this will ever exist
			d, err = c.addListValue(d, e, k, val)
			return d, err
		}
	} else {
		if len(k) == 0 {
			// not last element -> container
			d, err = c.addContainer(d, e, elems, qelems, val)
			return d, err
		} else {
			// not last element -> list + keys
			d, err = c.addList(d, e, k, elems, qelems, val)
			return d, err
		}
	}
}
