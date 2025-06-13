package jsonmerge

import (
	"encoding/json"
	"testing"

	"github.com/necroskillz/config-service/util/test"
	"gotest.tools/v3/assert"
)

func TestJsonMerge(t *testing.T) {
	t.Run("Merge", func(t *testing.T) {
		type testCase struct {
			a      string
			b      string
			expect string
		}

		run := func(t *testing.T, tc testCase) {
			var a any
			var b any
			err := json.Unmarshal([]byte(tc.a), &a)
			if err != nil {
				t.Fatal(err)
			}

			err = json.Unmarshal([]byte(tc.b), &b)
			if err != nil {
				t.Fatal(err)
			}
			merged := Merge(a, b)

			var expect any
			if err := json.Unmarshal([]byte(tc.expect), &expect); err != nil {
				t.Fatal(err)
			}

			assert.DeepEqual(t, expect, merged)
		}

		testCases := map[string]testCase{
			"object":               {a: `{"a": "a"}`, b: `{"b": "b"}`, expect: `{"a":"a","b":"b"}`},
			"object type mismatch": {a: `{"a": 1}`, b: `{"a": "2"}`, expect: `{"a":"2"}`},
			"remove key":           {a: `{"a": "a", "b": "b"}`, b: `{"a": null}`, expect: `{"a":null,"b":"b"}`},
			"array":                {a: `["a", "b"]`, b: `["c", "d"]`, expect: `["c","d"]`},
			"array nil":            {a: `["a", "b"]`, b: `null`, expect: `null`},
			"array type mismatch":  {a: `["a", 1]`, b: `1`, expect: `1`},
			"string":               {a: `"a"`, b: `"b"`, expect: `"b"`},
			"string nil":           {a: `"a"`, b: `null`, expect: `null`},
			"string type mismatch": {a: `"a"`, b: `1`, expect: `1`},
			"number":               {a: `1`, b: `2`, expect: `2`},
			"number nil":           {a: `1`, b: `null`, expect: `null`},
			"number type mismatch": {a: `1`, b: `"2"`, expect: `"2"`},
			"bool":                 {a: `true`, b: `false`, expect: `false`},
			"bool nil":             {a: `true`, b: `null`, expect: `null`},
			"bool type mismatch":   {a: `true`, b: `1`, expect: `1`},
			"nested object":        {a: `{"a": {"a": "a", "b": "b"}}`, b: `{"a": {"a": "d", "c": "c"}}`, expect: `{"a":{"a":"d","b":"b","c":"c"}}`},
		}

		test.RunCases(t, run, testCases)
	})
}
