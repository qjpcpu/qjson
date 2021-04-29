package qjson

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func BenchmarkMarshalQJSON(b *testing.B) {
	t, err := Decode(textTpl)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := t.MarshalJSON(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalQJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if tree, err := Decode(textTpl); err != nil {
			b.Fatal(err)
		} else {
			tree.Release()
		}
	}
}

func BenchmarkMarshalStd(b *testing.B) {
	m := make(map[string]interface{})
	if err := json.Unmarshal(textTpl, &m); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(m); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalStd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := make(map[string]interface{})
		if err := json.Unmarshal(textTpl, &m); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParallel(b *testing.B) {
	jsonFeed := make(map[string]string)
	files, err := ioutil.ReadDir("./test_feed")
	if err != nil {
		return
	}
	for _, file := range files {
		filename := "./test_feed/" + file.Name()
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			b.Fatalf("read test file %s %v", filename, err)
		}
		jsonFeed[file.Name()] = string(data)
	}
	jsonFeed["buildin"] = string(textTpl)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for _, js := range jsonFeed {
				validJSON(b, []byte(js))
			}
		}
	})
}

func validJSON(b *testing.B, data []byte) {
	compareFn := func(m interface{}, tree *JSONTree) {
		if mv, ok := m.(map[string]interface{}); ok {
			compareTreeWithMap(b, mv, tree.Root.ObjectValues)
		} else if av, ok := m.([]interface{}); ok {
			compareTreeWithArray(b, av, tree.Root.ArrayValues)
		} else {
			vv, err := json.Marshal(m)
			if err != nil {
				b.Fatal(err)
			}
			if string(vv) != tree.Root.Value {
				b.Fatalf("%s != %s", string(vv), tree.Root.Value)
			}
		}
	}
	var tree1, tree2 *JSONTree
	var err error
	tree1, err = Decode(data)
	if err != nil {
		b.Fatal(err)
	}

	var m interface{}
	json.Unmarshal(data, &m)
	compareFn(m, tree1)

	// compare again
	data, err = tree1.MarshalJSON()
	if err != nil {
		b.Fatal(err)
	}
	tree1.Release()

	tree2, err = Decode(data)
	if err != nil {
		b.Fatal(err)
	}
	compareFn(m, tree2)
	tree2.Release()
}

func compareTreeWithArray(b *testing.B, m []interface{}, elems []*Node) {
	if len(m) != len(elems) {
		b.Fatal("length not match")
	}
	for i, item := range m {
		tv := elems[i]
		switch elems[i].Type {
		case Null:
			if item != nil {
				b.Fatal("item should be nil")
			}
		case String:
			var s string
			err := json.Unmarshal([]byte(tv.Value), &s)
			if err != nil {
				b.Fatal(err)
			}
			if item.(string) != s {
				b.Fatalf("%v != %v", item.(string), s)
			}
		case Bool:
			var s bool
			err := json.Unmarshal([]byte(tv.Value), &s)
			if err != nil {
				b.Fatal(err)
			}
			if item.(bool) != s {
				b.Fatalf("%v != %v", item.(bool), s)
			}
		case Integer, Float:
			vs, _ := json.Marshal(item)
			if string(vs) != tv.Value {
				b.Fatalf("%s != %s", string(vs), tv.Value)
			}
		case Object:
			sub, ok := item.(map[string]interface{})
			if !ok {
				b.Fatal("not ok")
			}
			compareTreeWithMap(b, sub, tv.ObjectValues)
		case Array:
			arr, ok := item.([]interface{})
			if !ok {
				b.Fatal("not ok")
			}
			compareTreeWithArray(b, arr, tv.ArrayValues)
		}
	}
}

func compareTreeWithMap(b *testing.B, m map[string]interface{}, objectValues []*ObjectElem) {
	if len(m) != len(objectValues) {
		b.Fatal("length not match")
	}
	for k, v := range m {
		/* find key */
		var tv *Node
		var found bool
		for _, kv := range objectValues {
			var str string
			err := json.Unmarshal([]byte(kv.Key.Value), &str)
			if err != nil {
				b.Fatal(err)
			}
			if k == str {
				found = true
				tv = kv.Value
				break
			}
		}
		if !found {
			b.Fatalf("should find key %s", k)
			return
		}
		/* match value */
		switch tv.Type {
		case Null:
			if v != nil {
				b.Fatal("should be nil")
			}
		case String:
			var s string
			err := json.Unmarshal([]byte(tv.Value), &s)
			if err != nil {
				b.Fatal(err)
			}
			if v.(string) != s {
				b.Fatalf("%s != %s", v.(string), s)
			}
		case Bool:
			var s bool
			err := json.Unmarshal([]byte(tv.Value), &s)
			if err != nil {
				b.Fatal(err)
			}
			if v.(bool) != s {
				b.Fatalf("%v != %v", v.(bool), s)
			}
		case Float, Integer:
			vs, _ := json.Marshal(v)
			if string(vs) != tv.Value {
				b.Fatalf("%s != %s", string(vs), tv.Value)
			}
		case Object:
			sub, ok := v.(map[string]interface{})
			if !ok {
				b.Fatal("should be ok")
			}
			compareTreeWithMap(b, sub, tv.ObjectValues)
		case Array:
			arr, ok := v.([]interface{})
			if !ok {
				b.Fatal("should be ok")
			}
			compareTreeWithArray(b, arr, tv.ArrayValues)
		}
	}
}

/* test JSON snippets */
var (
	textTpl = []byte(`
{
  "destination_addresses": [
    "Washington, DC, USA",
    "Philadelphia, PA, USA",
    "Santa Barbara, CA, USA",
    "Miami, FL, USA",
    "Austin, TX, USA",
    "Napa County, CA, USA"
  ],
  "origin_addresses": [
    "New York, NY, USA"
  ],
  "rows": [{
    "elements": [{
        "distance": {
          "text": "227 mi",
          "value": 365468
        },
        "duration": {
          "text": "3 hours 54 mins",
          "value": 14064
        },
        "status": "OK"
      },
      {
        "distance": {
          "text": "94.6 mi",
          "value": 152193
        },
        "duration": {
          "text": "1 hour 44 mins",
          "value": 6227
        },
        "status": "OK"
      },
      {
        "distance": {
          "text": "2,878 mi",
          "value": 4632197
        },
        "duration": {
          "text": "1 day 18 hours",
          "value": 151772
        },
        "status": "OK"
      },
      {
        "distance": {
          "text": "1,286 mi",
          "value": 2069031
        },
        "duration": {
          "text": "18 hours 43 mins",
          "value": 67405
        },
        "status": "OK"
      },
      {
        "distance": {
          "text": "1,742 mi",
          "value": 2802972
        },
        "duration": {
          "text": "1 day 2 hours",
          "value": 93070
        },
        "status": "OK"
      },
      {
        "distance": {
          "text": "2,871 mi",
          "value": 4620514
        },
        "duration": {
          "text": "1 day 18 hours",
          "value": 152913
        },
        "status": "OK"
      }
    ]
  }],
"names":null,
  "status": "OK"
}
`)
	text2Tpl = []byte(`{ 
  "accounting" : [   
                     { "firstName" : "John",  
                       "lastName"  : "Doe",
                       "age"       : 23 },

                     { "firstName" : "Mary",  
                       "lastName"  : "Smith",
                        "age"      : 32 }
                 ],                            
  "sales"      : [ 
                     { "firstName" : "Sally", 
                       "lastName"  : "Green",
                        "age"      : 27 },

                     { "firstName" : "Jim",   
                       "lastName"  : "Galley",
                       "age"       : -1.2 }
                 ] 
} `)
	text3Tpl = []byte(`{
  "content": "{\"blocks\":[{\"key\":\"70eu2\",\"text\":\"show tex 2@ttx \",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[{\"offset\":14,\"length\":4,\"key\":0}],\"data\":{}}],\"entityMap\":{\"0\":{\"type\":\"mention\",\"mutability\":\"IMMUTABLE\",\"data\":{\"mention\":{\"id\":\"117\",\"name\":\"RTx\",\"additionalName\":\"eer\",\"emailPrefix\":\"674829897\",\"desc\":\"world123\",\"query\":\"\",\"localeNames\":{\"zh\":\"erte\",\"en\":\".7x5\"},\"avatar\":\"https://ggogle..png\"}}}},\"atEmployeeIds\":[\"697\"]}",
  "notify": [
    "1234567",
     1,
     1.2,
     -100.2,
     false,
     null
  ],
  "token": "edd05471-7823-5d38-8a7c-e60d763c3001",
  "version": 100,
  "t_uuid": "Av9DGOKlUgM",
  "take_t": true,
  "changesets": "[]"
}`)
)
