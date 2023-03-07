package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestField(t *testing.T) {
	type customStruct struct {
		MyValue       bool
		myHiddenValue string
		MyMap         map[string]interface{}
		MyStringMap   map[string]string
	}
	var nilStructPtr *customStruct

	testCases := []struct {
		Name     string
		Obj      interface{}
		Path     string
		Expected interface{}
		NotFound bool
	}{
		{
			Name: "Int Value",
			Obj: map[string]interface{}{
				"hello": 1,
			},
			Path:     "hello",
			Expected: 1,
		},
		{
			Name: "String Value",
			Obj: map[string]interface{}{
				"hello": "world",
			},
			Path:     "hello",
			Expected: "world",
		},
		{
			Name: "Slice Value",
			Obj: map[string]interface{}{
				"hello": []interface{}{
					"world", "rancher",
				},
			},
			Path: "hello",
			Expected: []interface{}{
				"world", "rancher",
			},
		},
		{
			Name: "Slice At Index",
			Obj: []interface{}{
				"hello", "world",
			},
			Path:     "[0]",
			Expected: "hello",
		},
		{
			Name: "Slice At Second Index",
			Obj: []interface{}{
				"hello", "world",
			},
			Path:     "[1]",
			Expected: "world",
		},
		{
			Name: "Slice At Nonexistent Index",
			Obj: []interface{}{
				"hello", "world",
			},
			Path:     "[2]",
			NotFound: true,
		},
		{
			Name: "Slice At Bad Index",
			Obj: []interface{}{
				"hello", "world",
			},
			Path:     "[]",
			NotFound: true,
		},
		{
			Name:     "Nil Key Access",
			Obj:      nil,
			Path:     "hello",
			NotFound: true,
		},
		{
			Name:     "Nil Int Index",
			Obj:      nil,
			Path:     "[5]",
			NotFound: true,
		},
		{
			Name:     "Nil String Index",
			Obj:      nil,
			Path:     `["hello"]`,
			NotFound: true,
		},
		{
			Name:     "Nil Struct",
			Obj:      nilStructPtr,
			Path:     `MyValue`,
			NotFound: true,
		},
		{
			Name:     "Empty Bool From Struct",
			Obj:      customStruct{},
			Path:     "MyValue",
			Expected: false,
		},
		{
			Name:     "Empty Map From Struct",
			Obj:      customStruct{},
			Path:     "MyMap",
			Expected: (map[string]interface{})(nil),
		},
		{
			Name: "Dot",
			Obj: []interface{}{
				"hello", "world",
			},
			Path: ".",
			Expected: []interface{}{
				"hello", "world",
			},
		},
		{
			Name: "Double Dot",
			Obj: []interface{}{
				"hello", "world",
			},
			Path: "..",
			Expected: []interface{}{
				"hello", "world",
			},
		},
		{
			Name: "Map",
			Obj: map[string]interface{}{
				"hello": map[string]interface{}{
					"world": "rancher",
				},
			},
			Path: "hello",
			Expected: map[string]interface{}{
				"world": "rancher",
			},
		},
		{
			Name: "Map At Key",
			Obj: map[string]interface{}{
				"hello": map[string]interface{}{
					"world": "rancher",
				},
			},
			Path:     "hello.world",
			Expected: "rancher",
		},
		{
			Name: "Map At String Index",
			Obj: map[string]interface{}{
				"hello": map[string]interface{}{
					"world": "rancher",
				},
			},
			Path:     `hello["world"]`,
			Expected: "rancher",
		},
		{
			Name: "Map At String Index With Single Quotes",
			Obj: map[string]interface{}{
				"hello": map[string]interface{}{
					"world": "rancher",
				},
			},
			Path:     `hello['world']`,
			Expected: "rancher",
		},
		{
			Name: "Map At String Index with Double Quotes",
			Obj: map[string]interface{}{
				"hello": map[string]interface{}{
					"world": "rancher",
				},
			},
			Path:     `hello["world"]`,
			Expected: "rancher",
		},
		{
			Name: "Complex Combination",
			Obj: map[string]interface{}{
				"hello": map[string]interface{}{
					"world": []interface{}{
						"not me",
						"not me either",
						map[string]interface{}{
							"rancher": map[string]interface{}{
								"world": []interface{}{
									"not me again",
									"not me again either",
									map[string]interface{}{
										"cattle": []map[string]interface{}{
											{
												"hull": []int{
													1,
													9001,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Path:     `hello.world[2].rancher["world"][2]["cattle"][0].hull[1]`,
			Expected: 9001,
		},
		{
			Name: "Accessing Embedded Structs",
			Obj: map[string]interface{}{
				"Chart": customStruct{
					MyValue: true,
				},
			},
			Path:     "Chart.MyValue",
			Expected: true,
		},
		{
			Name: "Accessing Hidden Fields In Embedded Structs",
			Obj: map[string]interface{}{
				"Chart": customStruct{
					myHiddenValue: "secret",
				},
			},
			Path:     "Chart.myHiddenValue",
			NotFound: true,
		},
		{
			Name: "Accessing Nested Fields In Embedded Structs",
			Obj: map[string]interface{}{
				"Chart": customStruct{
					MyMap: map[string]interface{}{
						"hello": customStruct{
							MyStringMap: map[string]string{
								"world": "rancher",
							},
						},
					},
				},
			},
			Path:     "Chart.MyMap.hello.MyStringMap.world",
			Expected: "rancher",
		},
		{
			Name: "Invalid Field In Struct",
			Obj: map[string]interface{}{
				"Chart": customStruct{
					MyMap: map[string]interface{}{
						"hello": customStruct{
							MyStringMap: map[string]string{
								"world": "rancher",
							},
						},
					},
				},
			},
			Path:     "Chart.MyNonExistentKey",
			NotFound: true,
		},
		{
			Name: "Accessing Field In Pointer Struct",
			Obj: map[string]interface{}{
				"Chart": &customStruct{
					MyMap: map[string]interface{}{
						"hello": &customStruct{
							MyStringMap: map[string]string{
								"world": "rancher",
							},
						},
					},
				},
			},
			Path:     "Chart.MyNonExistentKey",
			NotFound: true,
		},
		{
			Name: "Invalid Int Index Access On Map",
			Obj: map[string]interface{}{
				"hello": "world",
			},
			Path:     "hello[0]",
			NotFound: true,
		},
		{
			Name: "Invalid String Index Access On Slice",
			Obj: map[string]interface{}{
				"hello": []interface{}{
					"world", "rancher",
				},
			},
			Path:     `hello["world"]`,
			NotFound: true,
		},
		{
			Name: "Invalid String Index Without Quotes",
			Obj: map[string]interface{}{
				"hello": []interface{}{
					"world", "rancher",
				},
			},
			Path:     `hello[world]`,
			NotFound: true,
		},
		{
			Name: "Empty String Index With Double Quotes",
			Obj: map[string]interface{}{
				"hello": map[string]interface{}{
					"": "rancher",
				},
			},
			Path:     `hello[""]`,
			Expected: "rancher",
		},
		{
			Name: "Empty String Index With Single Quotes",
			Obj: map[string]interface{}{
				"hello": map[string]interface{}{
					"": "rancher",
				},
			},
			Path:     `hello[""]`,
			Expected: "rancher",
		},
		{
			Name: "Invalid Key Access On Slice",
			Obj: map[string]interface{}{
				"hello": []interface{}{
					"world", "rancher",
				},
			},
			Path:     `hello.world`,
			NotFound: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			value, found := Field[interface{}](tc.Obj, tc.Path)
			assert.Equal(t, tc.Expected, value)
			if tc.NotFound {
				assert.False(t, found, "expected not to find any value in object, found %s", value)
			} else {
				assert.True(t, found, "expected to find value %s in object at path %s, found nil", tc.Expected, tc.Path)
			}
		})
	}

	obj := map[string]interface{}{
		"string": "string",
		"int":    999,
		"slice":  []string{"hello", "world"},
		"map":    map[string]string{"hello": "world"},
	}

	t.Run("Get String Expect String", func(t *testing.T) {
		value, found := Field[string](obj, "string")
		assert.Equal(t, "string", value)
		assert.True(t, found, "could not find string")
	})
	t.Run("Get Int Expect Int", func(t *testing.T) {
		value, found := Field[int](obj, "int")
		assert.Equal(t, 999, value)
		assert.True(t, found, "could not find int")
	})
	t.Run("Get Slice Expect Slice", func(t *testing.T) {
		value, found := Field[[]string](obj, "slice")
		assert.Equal(t, []string{"hello", "world"}, value)
		assert.True(t, found, "could not find slice")
	})
	t.Run("Get Map Expect Map", func(t *testing.T) {
		value, found := Field[map[string]string](obj, "map")
		assert.Equal(t, map[string]string{"hello": "world"}, value)
		assert.True(t, found, "could not find map")
	})
	t.Run("Get String Expect Map", func(t *testing.T) {
		value, found := Field[map[string]string](obj, "string")
		assert.Nil(t, value, "expected no value to be found since value at 'string' is string, found this value")
		assert.False(t, found, "expected no value to be found since value at 'string' is string")
	})
}
