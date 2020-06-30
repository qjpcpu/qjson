package qjson

import (
	"bytes"
	"io/ioutil"
	"testing"

	"encoding/json"

	"github.com/stretchr/testify/suite"
)

type JSONTreeTestSuite struct {
	suite.Suite
}

func (suite *JSONTreeTestSuite) BeforeTest(suiteName, testName string) {

}

func (suite *JSONTreeTestSuite) AfterTest(suiteName, testName string) {
}

func TestJSONTree(t *testing.T) {
	suite.Run(t, &JSONTreeTestSuite{})
}

func (suite *JSONTreeTestSuite) ValidJSON(data []byte) {
	tree, err := Decode(data)
	suite.Nil(err)
	m := make(map[string]interface{})
	suite.Nil(json.Unmarshal(data, &m))
	suite.compareTreeWithMap(m, tree.Root.ObjectValues)
}

func (suite *JSONTreeTestSuite) compareTreeWithArray(m []interface{}, elems []*Node) {
	suite.Equal(len(m), len(elems))
	for i, item := range m {
		tv := elems[i]
		switch elems[i].Type {
		case Null:
			suite.Nil(item)
		case String:
			var s string
			suite.Nil(json.Unmarshal([]byte(tv.Value), &s))
			suite.Equal(item.(string), s)
		case Bool:
			var s bool
			suite.Nil(json.Unmarshal([]byte(tv.Value), &s))
			suite.Equal(item.(bool), s)
		case Integer, Float:
			vs, _ := json.Marshal(item)
			suite.Equal(string(vs), tv.Value)
		case Object:
			sub, ok := item.(map[string]interface{})
			suite.True(ok)
			suite.compareTreeWithMap(sub, tv.ObjectValues)
		case Array:
			arr, ok := item.([]interface{})
			suite.True(ok)
			suite.compareTreeWithArray(arr, tv.ArrayValues)
		}
	}
}

func (suite *JSONTreeTestSuite) compareTreeWithMap(m map[string]interface{}, objectValues []*ObjectElem) {
	suite.Equal(len(m), len(objectValues))
	for k, v := range m {
		var tv *Node
		var found bool
		for _, kv := range objectValues {
			var str string
			suite.Nil(json.Unmarshal([]byte(kv.Key.Value), &str))
			if k == str {
				found = true
				tv = kv.Value
				break
			}
		}
		if !found {
			suite.True(found, "should find key %s", k)
			return
		}
		switch tv.Type {
		case Null:
			suite.Nil(v)
		case String:
			var s string
			suite.Nil(json.Unmarshal([]byte(tv.Value), &s))
			suite.Equal(v.(string), s)
		case Bool:
			var s bool
			suite.Nil(json.Unmarshal([]byte(tv.Value), &s))
			suite.Equal(v.(bool), s)
		case Float, Integer:
			vs, _ := json.Marshal(v)
			suite.Equal(string(vs), tv.Value)
		case Object:
			sub, ok := v.(map[string]interface{})
			suite.True(ok)
			suite.compareTreeWithMap(sub, tv.ObjectValues)
		case Array:
			arr, ok := v.([]interface{})
			suite.True(ok)
			suite.compareTreeWithArray(arr, tv.ArrayValues)
		}
	}
}

func (suite *JSONTreeTestSuite) TestDecodeInvalidJSON() {
	invalidCheck := func(s string) {
		_, err := Decode([]byte(s))
		suite.NotNil(err)
	}
	invalidCheck(`1,`)
	invalidCheck(`,1`)
	invalidCheck(`s`)
	invalidCheck(`"s",`)
	invalidCheck(`}`)
	invalidCheck(`[`)
	invalidCheck(`]`)
	invalidCheck(`"`)
}

func (suite *JSONTreeTestSuite) TestDecodeSimpleInt() {
	bytes := []byte(`1`)
	tree, err := Decode(bytes)
	suite.Nil(err)
	suite.Equal(Integer, tree.Root.Type)
	suite.Equal("1", tree.Root.Value)
}

func (suite *JSONTreeTestSuite) TestDecodeSimpleString() {
	bytes := []byte(`""`)
	tree, err := Decode(bytes)
	suite.Nil(err)
	suite.Equal(String, tree.Root.Type)
	suite.Equal(`""`, tree.Root.Value)

	bytes = []byte(`"Hello,世界"`)
	tree, err = Decode(bytes)
	suite.Nil(err)
	suite.Equal(String, tree.Root.Type)
	suite.Equal(`"Hello,世界"`, tree.Root.Value)

	bytes = []byte(`"Hello,\"世界"`)
	tree, err = Decode(bytes)
	suite.Nil(err)
	suite.Equal(String, tree.Root.Type)
	suite.Equal(`"Hello,\"世界"`, tree.Root.Value)
}

func (suite *JSONTreeTestSuite) TestDecodeObject() {
	bytes := []byte(`{
"hello":"world","num":2,"em":{100:true,"lang":"golang", "other": null}}`)
	tree, err := Decode(bytes)
	suite.Nil(err)
	suite.Equal(Object, tree.Root.Type)
	suite.Len(tree.Root.ObjectValues, 3)
	suite.Equal(String, tree.Root.ObjectValues[0].Key.Type)
	suite.Equal(`"hello"`, tree.Root.ObjectValues[0].Key.Value)
	suite.Equal(String, tree.Root.ObjectValues[0].Value.Type)
	suite.Equal(`"world"`, tree.Root.ObjectValues[0].Value.Value)

	suite.Equal(String, tree.Root.ObjectValues[1].Key.Type)
	suite.Equal(`"num"`, tree.Root.ObjectValues[1].Key.Value)
	suite.Equal(Integer, tree.Root.ObjectValues[1].Value.Type)
	suite.Equal(`2`, tree.Root.ObjectValues[1].Value.Value)

	suite.Equal(Object, tree.Root.ObjectValues[2].Value.Type)
	em := tree.Root.ObjectValues[2].Value
	suite.Len(em.ObjectValues, 3)
	suite.Equal(Integer, em.ObjectValues[0].Key.Type)
	suite.Equal(Bool, em.ObjectValues[0].Value.Type)
	suite.Equal(`100`, em.ObjectValues[0].Key.Value)
	suite.Equal(`true`, em.ObjectValues[0].Value.Value)

	suite.Equal(String, em.ObjectValues[1].Key.Type)
	suite.Equal(String, em.ObjectValues[1].Value.Type)
	suite.Equal(`"lang"`, em.ObjectValues[1].Key.Value)
	suite.Equal(`"golang"`, em.ObjectValues[1].Value.Value)

	suite.Equal(String, em.ObjectValues[2].Key.Type)
	suite.Equal(Null, em.ObjectValues[2].Value.Type)
	suite.Equal(`"other"`, em.ObjectValues[2].Key.Value)
	suite.Equal(`null`, em.ObjectValues[2].Value.Value)
}

func (suite *JSONTreeTestSuite) TestValidateObject() {
	bytes := []byte(`{"hello":"world","num":2,"em":{"100":true,"lang":"golang", "other": null}}`)
	suite.ValidJSON(bytes)
}

func (suite *JSONTreeTestSuite) TestObjectToMap() {
	bytes := []byte(`{"hello":"world","num":2,"em":{"100":true,"lang":"golang", "other": null}}`)
	tree, err := Decode(bytes)
	suite.Nil(err)
	m := tree.Root.AsMap()
	suite.Equal(`"world"`, m[`hello`].Value)
	suite.Equal(`2`, m[`num`].Value)
}

func (suite *JSONTreeTestSuite) TestDecodeArray() {
	bytes := []byte(`[]`)
	tree, err := Decode(bytes)
	suite.Nil(err)
	suite.Equal(Array, tree.Root.Type)

	bytes = []byte(`[
1]`)
	tree, err = Decode(bytes)
	suite.Nil(err)
	suite.Equal(Array, tree.Root.Type)
	suite.Len(tree.Root.ArrayValues, 1)
	suite.Equal(Integer, tree.Root.ArrayValues[0].Type)
	suite.Equal(`1`, tree.Root.ArrayValues[0].Value)
}

func (suite *JSONTreeTestSuite) TestDecodeComplexJSON() {
	bytes := []byte(text2)

	tree, err := Decode(bytes)
	suite.Nil(err)
	suite.Equal(Object, tree.Root.Type)

	suite.Len(tree.Root.ObjectValues, 2)

	// accounting
	accouting := tree.Root.ObjectValues[0]
	suite.Equal(`"accounting"`, accouting.Key.Value)
	suite.Equal("32", accouting.Value.ArrayValues[1].ObjectValues[2].Value.Value)

	// sales
	sales := tree.Root.ObjectValues[1]
	suite.Equal(`"sales"`, sales.Key.Value)
	suite.Equal(Float, sales.Value.ArrayValues[1].ObjectValues[2].Value.Type)
	suite.Equal("-1.2", sales.Value.ArrayValues[1].ObjectValues[2].Value.Value)
	suite.Equal(`"Galley"`, sales.Value.ArrayValues[1].ObjectValues[1].Value.Value)
}

func (suite *JSONTreeTestSuite) TestDecodeComplexJSONWithIndent() {
	m := make(map[string]string)
	m["Text"] = text2
	data, _ := json.MarshalIndent(m, "\t", "    ")
	tree, err := Decode(data)
	suite.Nil(err)
	suite.Equal(Object, tree.Root.Type)
	var str string
	json.Unmarshal([]byte(tree.Root.ObjectValues[0].Value.Value), &str)
	suite.Equal(text2, str)

	tree, err = Decode([]byte(text2))
	suite.Nil(err)
	suite.Equal(Object, tree.Root.Type)
	m2 := make(map[string]interface{})
	x, _ := json.Marshal(tree)
	err = json.Unmarshal(x, &m2)
	suite.Nil(err)
	_, err = json.Marshal(m2)
	suite.Nil(err)
}

func (suite *JSONTreeTestSuite) TestColorMarshal() {
	tree, err := Decode([]byte(text1))
	suite.Nil(err)
	/* should have color */
	data := tree.ColorfulMarshal()
	suite.T().Logf("%s", data)
	_, err = json.Marshal(tree)
	suite.Nil(err)
}

func (suite *JSONTreeTestSuite) TestDecodeComplexJSONWithFloat() {
	tree, err := Decode([]byte(text3))
	suite.Nil(err)

	suite.Equal(`"content"`, tree.Root.ObjectValues[0].Key.Value)
	suite.Equal(`"edd05471-7823-5d38-8a7c-e60d763c3001"`, tree.Root.ObjectValues[2].Value.Value)
	suite.Equal(`"1234567"`, tree.Root.ObjectValues[1].Value.ArrayValues[0].Value)
	suite.Equal(`1`, tree.Root.ObjectValues[1].Value.ArrayValues[1].Value)
	suite.Equal(`1.2`, tree.Root.ObjectValues[1].Value.ArrayValues[2].Value)
	suite.Equal(`-100.2`, tree.Root.ObjectValues[1].Value.ArrayValues[3].Value)
	suite.Equal(`false`, tree.Root.ObjectValues[1].Value.ArrayValues[4].Value)
	suite.Equal(`null`, tree.Root.ObjectValues[1].Value.ArrayValues[5].Value)

	m := make(map[string]interface{})
	json.Unmarshal([]byte(text3), &m)
	var s string
	json.Unmarshal([]byte(tree.Root.ObjectValues[0].Value.Value), &s)
	suite.Equal(m["content"].(string), s)
}

func (suite *JSONTreeTestSuite) TestDecodeUseStdDecoder() {
	tree := &JSONTree{}
	err := json.NewDecoder(bytes.NewBuffer([]byte(text3))).Decode(tree)
	suite.Nil(err)

	suite.Equal(`"content"`, tree.Root.ObjectValues[0].Key.Value)
	suite.Equal(`"edd05471-7823-5d38-8a7c-e60d763c3001"`, tree.Root.ObjectValues[2].Value.Value)
	suite.Equal(`"1234567"`, tree.Root.ObjectValues[1].Value.ArrayValues[0].Value)
	suite.Equal(`1`, tree.Root.ObjectValues[1].Value.ArrayValues[1].Value)
	suite.Equal(`1.2`, tree.Root.ObjectValues[1].Value.ArrayValues[2].Value)
	suite.Equal(`-100.2`, tree.Root.ObjectValues[1].Value.ArrayValues[3].Value)
	suite.Equal(`false`, tree.Root.ObjectValues[1].Value.ArrayValues[4].Value)
	suite.Equal(`null`, tree.Root.ObjectValues[1].Value.ArrayValues[5].Value)

	m := make(map[string]interface{})
	json.Unmarshal([]byte(text3), &m)
	var s string
	json.Unmarshal([]byte(tree.Root.ObjectValues[0].Value.Value), &s)
	suite.Equal(m["content"].(string), s)
}

func (suite *JSONTreeTestSuite) TestDecodeComplexJSONWithStd() {
	tree := &JSONTree{}
	err := json.Unmarshal([]byte(text3), tree)
	suite.Nil(err)

	suite.Equal(`"content"`, tree.Root.ObjectValues[0].Key.Value)
	suite.Equal(`"edd05471-7823-5d38-8a7c-e60d763c3001"`, tree.Root.ObjectValues[2].Value.Value)
	suite.Equal(`"1234567"`, tree.Root.ObjectValues[1].Value.ArrayValues[0].Value)
	suite.Equal(`1`, tree.Root.ObjectValues[1].Value.ArrayValues[1].Value)
	suite.Equal(`1.2`, tree.Root.ObjectValues[1].Value.ArrayValues[2].Value)
	suite.Equal(`-100.2`, tree.Root.ObjectValues[1].Value.ArrayValues[3].Value)
	suite.Equal(`false`, tree.Root.ObjectValues[1].Value.ArrayValues[4].Value)
	suite.Equal(`null`, tree.Root.ObjectValues[1].Value.ArrayValues[5].Value)

	m := make(map[string]interface{})
	json.Unmarshal([]byte(text3), &m)
	var s string
	json.Unmarshal([]byte(tree.Root.ObjectValues[0].Value.Value), &s)
	suite.Equal(m["content"].(string), s)
}

func (suite *JSONTreeTestSuite) TestWithStdLib() {
	files, _ := ioutil.ReadDir("./test_feed")
	if len(files) > 0 {
		suite.T().Logf("there are %d json files for testing", len(files))
	}
	for _, file := range files {
		filename := "./test_feed/" + file.Name()
		data, err := ioutil.ReadFile(filename)
		suite.Nil(err, "read test file %s", filename)
		suite.ValidJSON(data)
	}
}

func (suite *JSONTreeTestSuite) TestStringBytes() {
	s := "hello"
	suite.Equal("hello", bytesToString(stringToBytes(s)))
}

func (suite *JSONTreeTestSuite) TestStringEncoderDecoder() {
	s := `"hello' " \n = \r`
	m1 := MarshalString([]byte(s))
	m2, err := json.Marshal(s)
	suite.Nil(err)
	suite.Equal(string(m1), string(m2))

	m3, err := UnmarshalString(m1)
	suite.Nil(err)
	var s1 string
	json.Unmarshal(m1, &s1)
	suite.Equal(s1, string(m3))
	suite.Equal(s1, s)
}

func (suite *JSONTreeTestSuite) TestRelease() {
	tree := &JSONTree{}
	err := json.Unmarshal([]byte(text3), tree)
	suite.Nil(err)
	suite.NotNil(tree.Root)
	tree.Release()
	suite.Nil(tree.Root)
	tree.Release()
	suite.Nil(tree.Root)
}

/* test JSON snippets */
var (
	text1 = ` 
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
`
	text2 = `{ 
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
} `
	text3 = `{
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
}`
)

func (suite *JSONTreeTestSuite) TestCreateObjectNode() {
	tree := makeNewTree()
	tree.Root = CreateObjectNode()
	tree.Root.SetObjectStringElem("KEY", "VAL")
	tree.Root.SetObjectBoolElem("KEY1", true)
	tree.Root.SetObjectIntElem("KEY2", 12)
	ret := string(JSONMarshalWithPanic(tree))
	suite.Equal(`{"KEY":"VAL","KEY1":true,"KEY2":12}`, ret)
	suite.Equal("VAL", tree.Root.FindNodeByKey("KEY").AsString())

	tree.Root.SetObjectBoolElem("KEY2", false)
	ret = string(JSONMarshalWithPanic(tree))
	suite.Equal(`{"KEY":"VAL","KEY1":true,"KEY2":false}`, ret)
}

func (suite *JSONTreeTestSuite) TestIsXXX() {
	suite.True(CreateNode().IsNull())
	suite.True(CreateStringNode().IsString())
	suite.True(CreateIntegerNode().IsInteger())
	suite.True(CreateFloatNode().IsFloat())
	suite.True(CreateBoolNode().IsBool())
	suite.True(CreateIntegerNode().IsNumber())
	suite.True(CreateFloatNode().IsNumber())
}

func (suite *JSONTreeTestSuite) TestConvertSimpleObject() {
	str := "Text"
	obj := &struct {
		Name string
		Text *string `json:"text"`
		Int  int64   `json:"-"`
		Ptr  *int64  `json:"ptr,omitempty"`
	}{Name: "Jack", Int: 100, Text: &str}
	tree, err := ConvertToJSONTree(obj)
	suite.Nil(err)
	std, _ := json.Marshal(obj)
	q, _ := json.Marshal(tree)
	suite.Equal(string(std), string(q))
}

func (suite *JSONTreeTestSuite) TestConvertEmbedObject() {
	type Inner struct {
		X string `json:"xt"`
	}
	type Inner2 struct {
		T string
	}
	str := "Text"
	obj := &struct {
		Name string
		Text *string `json:"text"`
		Int  int64   `json:"-"`
		Ptr  *int64  `json:"ptr,omitempty"`
		Inner
		*Inner2
	}{Name: "Jack", Int: 100, Text: &str}
	obj.Inner.X = "33"
	obj.Inner2 = &Inner2{T: "x"}
	tree, err := ConvertToJSONTree(obj)
	suite.Nil(err)
	std, _ := json.Marshal(obj)
	q, _ := json.Marshal(tree)
	suite.T().Log(string(std), string(q))
	suite.Equal(string(std), string(q))
}

func (suite *JSONTreeTestSuite) TestConvertMap() {
	type Inner struct {
		X string `json:"xt"`
	}
	obj := map[string]interface{}{
		"aa": 100,
		"bb": &Inner{X: "34"},
	}
	tree, err := ConvertToJSONTree(obj)
	suite.Nil(err)
	std, _ := json.Marshal(obj)
	q, _ := json.Marshal(tree)
	suite.T().Log(string(std), string(q))
	suite.Equal(string(std), string(q))
}

func (suite *JSONTreeTestSuite) TestConvertSlice() {
	type Inner struct {
		X string `json:"xt"`
	}
	obj := []*Inner{
		{X: "1"},
		{X: "2"},
	}
	tree, err := ConvertToJSONTree(obj)
	suite.Nil(err)
	std, _ := json.Marshal(obj)
	q, _ := json.Marshal(tree)
	suite.T().Log(string(std), string(q))
	suite.Equal(string(std), string(q))
}

func (suite *JSONTreeTestSuite) TestConvertWithInterface() {
	type Response struct {
		Code int         `json:"code"`
		Data interface{} `json:"data"`
	}
	type Inner struct {
		X string `json:"xt"`
	}
	obj := &Response{
		Code: 100,
		Data: &Inner{X: "jkljlk"},
	}
	tree, err := ConvertToJSONTree(obj)
	suite.Nil(err)
	std, _ := json.Marshal(obj)
	q, _ := json.Marshal(tree)
	suite.T().Log(string(std), string(q))
	suite.Equal(string(std), string(q))
}

func (suite *JSONTreeTestSuite) TestRemoveProperty() {
	str := []byte(`{"a":1,"b":2}`)
	tree, err := Decode(str)
	suite.Nil(err)
	ok := tree.Root.RemoveObjectElemByKey("a")
	suite.True(ok)

	ok = tree.Root.RemoveObjectElemByKey("a")
	suite.False(ok)

	s := JSONMarshalWithPanic(tree)
	suite.Equal(`{"b":2}`, string(s))
}

func (suite *JSONTreeTestSuite) TestFindObjectElemByKeyRecursive() {
	str := []byte(`{"a":1,"b":2,"c":{"d":{"e":1},"t":2}}`)
	tree, err := Decode(str)
	suite.Nil(err)
	v := tree.Root.FindObjectElemByKeyRecursive("c.d.e")
	suite.Equal("1", v.Value.AsString())
	v = tree.Root.FindObjectElemByKeyRecursive("c.t")
	suite.Equal("2", v.Value.AsString())
	v = tree.Root.FindObjectElemByKeyRecursive("c.x.e")
	suite.Nil(v)
	v = tree.Root.FindObjectElemByKeyRecursive("c.x.e.f")
	suite.Nil(v)
	v = tree.Root.FindObjectElemByKeyRecursive("m.x.e.f")
	suite.Nil(v)
}

func (suite *JSONTreeTestSuite) TestSetString() {
	str := []byte(`{"a":1,"b":2,"c":{"d":{"e":1},"t":2}}`)
	tree, err := Decode(str)
	suite.Nil(err)
	v := tree.Root.FindObjectElemByKey("c")
	s, err := json.Marshal(v.Value)
	suite.Nil(err)
	v.Value.SetString(string(s))
	v.Value.Type = String
	json, err := tree.MarshalJSON()
	suite.Nil(err)
	suite.Equal(`{"a":1,"b":2,"c":"{\"d\":{\"e\":1},\"t\":2}"}`, string(json))
}

func (suite *JSONTreeTestSuite) TestColorMarshalWithIndent() {
	tree, err := Decode([]byte(text1))
	suite.Nil(err)

	/* should have color */
	data := tree.ColorfulMarshalWithIndent()
	suite.T().Logf("%s", data)

	/* should have color */
	data = tree.ColorfulMarshal()
	suite.T().Logf("%s", data)

	suite.T().Logf("%s %s", makeNewTree().ColorfulMarshal(), makeNewTree().ColorfulMarshalWithIndent())
	suite.T().Logf("%s %s", new(JSONTree).ColorfulMarshal(), new(JSONTree).ColorfulMarshalWithIndent())
}
