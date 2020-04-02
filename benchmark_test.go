package qjson

import (
	"encoding/json"
	"testing"
)

func BenchmarkUnmarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := Decode(textTpl); err != nil {
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

func BenchmarkMarshal(b *testing.B) {
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
