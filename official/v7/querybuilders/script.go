package querybuilders

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Script holds all the parameters necessary to compile or find in cache
// and then execute a script.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/7.0/modules-scripting.html
// for details of scripting.
type Script struct {
	script string
	typ    string
	lang   string
	params map[string]interface{}
}

// NewScript creates and initializes a new Script. By default, it is of
// type "inline". Use NewScriptStored for a stored script (where type is "id").
func NewScript(script string) *Script {
	return &Script{
		script: script,
		typ:    "inline",
		params: make(map[string]interface{}),
	}
}

// Script is either the cache key of the script to be compiled/executed
// or the actual script source code for inline scripts. For indexed
// scripts this is the id used in the request. For file scripts this is
// the file name.
func (s *Script) Script(script string) *Script {
	s.script = script
	return s
}

// Type sets the type of script: "inline" or "id".
func (s *Script) Type(typ string) *Script {
	s.typ = typ
	return s
}

// Lang sets the language of the script. The default scripting language
// is Painless ("painless").
// See https://www.elastic.co/guide/en/elasticsearch/reference/7.0/modules-scripting.html
// for details.
func (s *Script) Lang(lang string) *Script {
	s.lang = lang
	return s
}

// Param adds a key/value pair to the parameters that this script will be executed with.
func (s *Script) Param(name string, value interface{}) *Script {
	if s.params == nil {
		s.params = make(map[string]interface{})
	}
	s.params[name] = value
	return s
}

// Params sets the map of parameters this script will be executed with.
func (s *Script) Params(params map[string]interface{}) *Script {
	s.params = params
	return s
}

// Source returns the JSON serializable data for this Script.
func (s *Script) Source() (interface{}, error) {
	if s.typ == "" && s.lang == "" && len(s.params) == 0 {
		return s.script, nil
	}
	source := make(map[string]interface{})
	// Beginning with 6.0, the type can only be "source" or "id"
	if s.typ == "" || s.typ == "inline" {
		src := s.rawScriptSource(s.script)
		source["source"] = src
	} else {
		source["id"] = s.script
	}
	if s.lang != "" {
		source["lang"] = s.lang
	}
	if len(s.params) > 0 {
		source["params"] = s.params
	}
	return source, nil
}

// rawScriptSource returns an embeddable script. If it uses a short
// script form, e.g. "ctx._source.likes++" (without the quotes), it
// is quoted. Otherwise it returns the raw script that will be directly
// embedded into the JSON data.
func (s *Script) rawScriptSource(script string) interface{} {
	v := strings.TrimSpace(script)
	if !strings.HasPrefix(v, "{") && !strings.HasPrefix(v, `"`) {
		v = fmt.Sprintf("%q", v)
	}
	raw := json.RawMessage(v)
	return &raw
}
