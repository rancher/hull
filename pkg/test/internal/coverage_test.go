package internal

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateCoverage(t *testing.T) {
	type ExampleStruct struct {
		First struct {
			Name int
		}
		Second struct {
			Third struct {
				NamePtr *float64
			}
		}
		Third struct {
			Fourth struct {
				Fifth struct {
					hidden string
				}
			}
		}
		Fourth struct {
			NameSlice []rune
			Fifth     struct {
				NameMap map[string]string
				Sixth   struct {
					NameSliceSlice [][]float64
					NameMapMap     map[string]map[string]int64
					Seventh        struct {
						NameSliceMap []map[string]byte
						NameMapSlice map[string][]interface{}
					}
				}
			}
		}
		Fifth []struct {
			NameHello string
			Sixth     map[string][]map[string]struct {
				Seventh map[string][]struct {
					NameWorld interface{}
				}
			}
		}
	}

	expectedKeys := getAllKeysFromStructType(reflect.TypeOf(ExampleStruct{}))
	numExpectedKeys := float64(len(expectedKeys))

	testCases := []struct {
		Name     string
		Values   map[string]interface{}
		Struct   interface{}
		Coverage float64
	}{
		{
			Name:     "No Coverage",
			Values:   nil,
			Struct:   ExampleStruct{},
			Coverage: 0,
		},
		{
			Name: "One Unknown",
			Values: map[string]interface{}{
				"first": map[string]interface{}{
					"hello": 5, // unknown
				},
			},
			Struct:   ExampleStruct{},
			Coverage: 0,
		},
		{
			Name: "One Known",
			Values: map[string]interface{}{
				"first": map[string]interface{}{
					"name": 5, // known
				},
			},
			Struct:   ExampleStruct{},
			Coverage: 1 / numExpectedKeys,
		},
		{
			Name: "One Known One Unknown",
			Values: map[string]interface{}{
				"first": map[string]interface{}{
					"name":  5, // known
					"hello": 5, // unknown
				},
			},
			Struct:   ExampleStruct{},
			Coverage: 1 / numExpectedKeys,
		},
		{
			Name: "Two Known One Unknown",
			Values: map[string]interface{}{
				"first": map[string]interface{}{
					"name":  5, // known
					"hello": 5, // unknown
				},
				"second": map[string]interface{}{
					"third": map[string]interface{}{
						"namePtr": "hello", // knwon
					},
				},
			},
			Struct:   ExampleStruct{},
			Coverage: 2 / numExpectedKeys,
		},
		{
			Name: "Six Known",
			Values: map[string]interface{}{
				"first": map[string]interface{}{
					"name": 5, // known
				},
				"second": map[string]interface{}{
					"third": map[string]interface{}{
						"namePtr": "hello", // known
					},
				},
				"fourth": map[string]interface{}{
					"nameSlice": true, // known
					"fifth": map[string]interface{}{
						"nameMap": []rune{'/'}, // known
						"sixth": map[string]interface{}{
							"nameSliceSlice": true,  // known
							"nameMapMap":     100.0, // known
						},
					},
				},
			},
			Struct:   ExampleStruct{},
			Coverage: 6 / numExpectedKeys,
		},
		{
			Name: "Nested Two Known",
			Values: map[string]interface{}{
				"fourth": map[string]interface{}{
					"fifth": map[string]interface{}{
						"sixth": map[string]interface{}{
							"seventh": map[string]interface{}{
								"nameSliceMap": []map[string]interface{}{
									{
										"hi": map[string]interface{}{
											"hi": 5,
										},
									},
								}, // known
								"nameMapSlice": map[string][]interface{}{
									"hi": {5},
								}, // known
							},
						},
					},
				},
			},
			Struct:   ExampleStruct{},
			Coverage: 2 / numExpectedKeys,
		},
		{
			Name: "Super Nested Two Known",
			Values: map[string]interface{}{
				"fifth": []interface{}{
					map[string]interface{}{
						"nameHello": "hello", // known
						"sixth": map[string]interface{}{
							"world": []interface{}{
								map[string]interface{}{
									"rancher": map[string]interface{}{
										"seventh": map[string]interface{}{
											"hull": []interface{}{
												map[string]interface{}{
													"nameWorld": "recursion", // known
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
			Struct:   ExampleStruct{},
			Coverage: 2 / numExpectedKeys,
		},
		{
			Name: "Six Known, Two Nil",
			Values: map[string]interface{}{
				"first": map[string]interface{}{
					"name": 5, // known
				},
				"second": map[string]interface{}{
					"third": map[string]interface{}{
						"namePtr": "hello", // known
					},
				},
				"fourth": map[string]interface{}{
					"nameSlice": true, // known
					"fifth": map[string]interface{}{
						"nameMap": []rune{'/'}, // known
						"sixth": map[string]interface{}{
							"nameSliceSlice": nil,   // known, but nil
							"nameMapMap":     100.0, // known
							"seventh": map[string]interface{}{
								"nameSliceMap": nil,                        // known, but nil
								"nameMapSlice": map[string][]interface{}{}, // known
							},
						},
					},
				},
			},
			Struct:   ExampleStruct{},
			Coverage: 6 / numExpectedKeys,
		},
		{
			Name: "Eight Known",
			Values: map[string]interface{}{
				"first": map[string]interface{}{
					"name": 5, // known
				},
				"second": map[string]interface{}{
					"third": map[string]interface{}{
						"namePtr": "hello", // known
					},
				},
				"fourth": map[string]interface{}{
					"nameSlice": true, // known
					"fifth": map[string]interface{}{
						"nameMap": []rune{'/'}, // known
						"sixth": map[string]interface{}{
							"nameSliceSlice": true,  // known
							"nameMapMap":     100.0, // known
							"seventh": map[string]interface{}{
								"nameSliceMap": []map[string]interface{}{
									{
										"hi": map[string]interface{}{
											"hi": 5,
										},
									},
								}, // known
								"nameMapSlice": map[string][]interface{}{
									"hi": {5},
								}, // known
							},
						},
					},
				},
			},
			Struct:   ExampleStruct{},
			Coverage: 8 / numExpectedKeys,
		},
		{
			Name: "Full Coverage",
			Values: map[string]interface{}{
				"first": map[string]interface{}{
					"name": 5, // known
				},
				"second": map[string]interface{}{
					"third": map[string]interface{}{
						"namePtr": "hello", // known
					},
				},
				"fourth": map[string]interface{}{
					"nameSlice": true, // known
					"fifth": map[string]interface{}{
						"nameMap": []rune{'/'}, // known
						"sixth": map[string]interface{}{
							"nameSliceSlice": true,  // known
							"nameMapMap":     100.0, // known
							"seventh": map[string]interface{}{
								"nameSliceMap": []map[string]interface{}{
									{
										"hi": map[string]interface{}{
											"hi": 5,
										},
									},
								}, // known
								"nameMapSlice": map[string][]interface{}{
									"hi": {5},
								}, // known
							},
						},
					},
				},
				"fifth": []interface{}{
					map[string]interface{}{
						"nameHello": "hello", // known
						"sixth": map[string]interface{}{
							"world": []interface{}{
								map[string]interface{}{
									"rancher": map[string]interface{}{
										"seventh": map[string]interface{}{
											"hull": []interface{}{
												map[string]interface{}{
													"nameWorld": "recursion", // known
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
			Struct:   ExampleStruct{},
			Coverage: 1,
		},
		{
			Name: "All Nil Coverage",
			Values: map[string]interface{}{
				"first": map[string]interface{}{
					"name": nil, // known, but nil
				},
				"second": map[string]interface{}{
					"third": map[string]interface{}{
						"namePtr": nil, // known, but nil
					},
				},
				"fourth": map[string]interface{}{
					"nameSlice": nil, // known, but nil
					"fifth": map[string]interface{}{
						"nameMap": nil, // known, but nil
						"sixth": map[string]interface{}{
							"nameSliceSlice": nil, // known, but nil
							"nameMapMap":     nil, // known, but nil
							"seventh": map[string]interface{}{
								"nameSliceMap": nil, // known, but nil
								"nameMapSlice": nil, // known, but nil
							},
						},
					},
				},
				"fifth": []interface{}{
					map[string]interface{}{
						"nameHello": nil, // known, but nil
						"sixth": map[string]interface{}{
							"world": []interface{}{
								map[string]interface{}{
									"rancher": map[string]interface{}{
										"seventh": map[string]interface{}{
											"hull": []interface{}{
												map[string]interface{}{
													"nameWorld": nil, // known, but nil
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
			Struct:   ExampleStruct{},
			Coverage: 0,
		},
		{
			Name: "Full Coverage With Unknowns",
			Values: map[string]interface{}{
				"first": map[string]interface{}{
					"name":    5,         // known
					"unknown": "unknown", // unknown
				},
				"second": map[string]interface{}{
					"third": map[string]interface{}{
						"namePtr": "hello",   // known
						"unknown": "unknown", // unknown
					},
				},
				"fourth": map[string]interface{}{
					"nameSlice": true, // known
					"fifth": map[string]interface{}{
						"nameMap": []rune{'/'}, // known
						"sixth": map[string]interface{}{
							"nameSliceSlice": true,      // known
							"unknown":        "unknown", // unknown
							"nameMapMap":     100.0,     // known
							"seventh": map[string]interface{}{
								"nameSliceMap": []map[string]interface{}{
									{
										"hi": map[string]interface{}{
											"hi": 5,
										},
									},
								}, // known
								"unknown": "unknown", // unknown
								"nameMapSlice": map[string][]interface{}{
									"hi": {5},
								}, // known
							},
						},
						"unknown": "unknown", // unknown
					},
				},
				"fifth": []interface{}{
					map[string]interface{}{
						"nameHello": "hello",   // known
						"unknown":   "unknown", // unknown
						"sixth": map[string]interface{}{
							"unknown": "unknown", // unknown
							"world": []interface{}{
								nil,
								map[string]interface{}{
									"unknown": "unknown",
								}, // unknown
								[]interface{}{
									"unknown",
								}, // unknown
								map[string]interface{}{
									"unknown": "unknown", // unknown
									"rancher": map[string]interface{}{
										"unknown": "unknown", // unknown
										"seventh": map[string]interface{}{
											"hull": []interface{}{
												[]interface{}{
													"unknown",
												}, // unknown
												map[string]interface{}{
													"unknown": "unknown",
												}, // unknown
												map[string]interface{}{
													"nameWorld": "recursion", // known
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
			Struct:   ExampleStruct{},
			Coverage: 1,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			coverage, report := CalculateCoverage(tc.Values, reflect.TypeOf(tc.Struct))
			assert.Equal(t, tc.Coverage, coverage, report)
		})
	}
}

func TestGetSetKeysFromMapInterface(t *testing.T) {
	testCases := []struct {
		Name    string
		Values  map[string]interface{}
		SetKeys []string
	}{
		{
			Name:    "Nil",
			Values:  nil,
			SetKeys: nil,
		},
		{
			Name:    "No Values",
			Values:  map[string]interface{}{},
			SetKeys: []string{},
		},
		{
			Name: "Single Field",
			Values: map[string]interface{}{
				"hello": "world",
			},
			SetKeys: []string{".hello"},
		},
		{
			Name: "List",
			Values: map[string]interface{}{
				"hello": []string{"world"},
			},
			SetKeys: []string{".hello"},
		},
		{
			Name: "List List",
			Values: map[string]interface{}{
				"hello": [][]string{
					{"world"},
				},
			},
			SetKeys: []string{".hello"},
		},
		{
			Name: "List List List",
			Values: map[string]interface{}{
				"hello": [][][]string{
					{{"world"}},
				},
			},
			SetKeys: []string{".hello"},
		},
		{
			Name: "Map",
			Values: map[string]interface{}{
				"hello": map[string]interface{}{
					"world": "rancher",
				},
			},
			SetKeys: []string{".hello", ".hello.world"},
		},
		{
			Name: "Map Map",
			Values: map[string]interface{}{
				"hello": map[string]interface{}{
					"world": map[string]interface{}{
						"rancher": "hull",
					},
				},
			},
			SetKeys: []string{
				".hello",               // if struct { Hello interface{} }
				".hello.world",         // if struct { Hello struct{ World interface{} } }
				".hello[].rancher",     // if struct { Hello map[string]struct{ Rancher interface{} } } or { Hello []struct{ Rancher interface{} } }
				".hello.world.rancher", // if struct { Hello struct{ World struct { Rancher interface{} } } }
			},
		},
		{
			Name: "Map Slice",
			Values: map[string]interface{}{
				"hello": map[string]interface{}{
					"world": []interface{}{
						"rancher",
					},
				},
			},
			SetKeys: []string{
				".hello",       // if struct { Hello interface{} }
				".hello.world", // if struct { Hello struct{ World []interface{} } }
			},
		},
		{
			Name: "Slice Map",
			Values: map[string]interface{}{
				"hello": []interface{}{
					map[string]interface{}{
						"world": "rancher",
					},
				},
			},
			SetKeys: []string{
				".hello",         // if struct { Hello []interface{} }
				".hello[].world", // if struct { Hello []struct{ World interface{} } }
			},
		},
		{
			Name: "Complex",
			Values: map[string]interface{}{
				"hello":  "world",
				"hello2": []string{"world2"},
				"hello3": [][]string{
					{"world3"},
				},
				"hello4": [][][]string{
					{{"world4"}},
				},
				"hello5": map[string]interface{}{
					"world5": "rancher5",
				},
				"hello6": map[string]interface{}{
					"world6": map[string]interface{}{
						"rancher6": "hull6",
					},
				},
				"hello7": map[string]interface{}{
					"world7": []interface{}{
						"rancher7",
					},
				},
				"hello8": []interface{}{
					map[string]interface{}{
						"world8": "rancher8",
					},
				},
			},
			SetKeys: []string{
				".hello",
				".hello2",
				".hello3",
				".hello4",
				".hello5", ".hello5.world5",
				".hello6", ".hello6.world6", ".hello6[].rancher6", ".hello6.world6.rancher6",
				".hello7", ".hello7.world7",
				".hello8", ".hello8[].world8",
			},
		},
		{
			Name: "Recursively Complex",
			Values: map[string]interface{}{
				"hello": []interface{}{
					[]interface{}{
						map[string]interface{}{
							"world": []interface{}{
								map[string]interface{}{
									"rancher": map[string]interface{}{
										"hull": []string{
											"recursion",
										},
									},
								},
							},
						},
					},
				},
			},
			SetKeys: []string{
				".hello",
				".hello[][].world",
				".hello[][].world[].rancher",
				".hello[][][][].rancher",
				".hello[][].world[].rancher.hull",
				".hello[][][][].rancher.hull",
				".hello[][].world[][].hull",
				".hello[][][][][].hull",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			setKeyMap := getSetKeysFromMapInterface(tc.Values)
			expectedSetKeyMap := make(map[string]bool, len(tc.SetKeys))
			for _, k := range tc.SetKeys {
				expectedSetKeyMap[k] = true
			}
			assert.Equal(t, expectedSetKeyMap, setKeyMap)
		})
	}
}

func TestGetAllKeysFromStructType(t *testing.T) {
	type ComplexWithStructs struct {
		First struct {
			Name int
		}
		Second struct {
			Third struct {
				NamePtr *float64
			}
		}
		Third struct {
			Fourth struct {
				Fifth struct {
					hidden string
				}
			}
		}
		Fourth struct {
			NameSlice []rune
			Fifth     struct {
				NameMap map[string]string
				Sixth   struct {
					NameSliceSlice [][]float64
					NameMapMap     map[string]map[string]int64
					Seventh        struct {
						NameSliceMap []map[string]byte
						NameMapSlice map[string][]interface{}
					}
				}
			}
		}
		Fifth []struct {
			NameHello string
			Sixth     map[string][]map[string]struct {
				Seventh map[string][]struct {
					NameWorld interface{}
				}
			}
		}
	}

	type complexWithStructs struct {
		First struct {
			Name int
		}
		Second struct {
			Third struct {
				NamePtr *float64
			}
		}
		Third struct {
			Fourth struct {
				Fifth struct {
					hidden string
				}
			}
		}
		Fourth struct {
			NameSlice []rune
			Fifth     struct {
				NameMap map[string]string
				Sixth   struct {
					NameSliceSlice [][]float64
					NameMapMap     map[string]map[string]int64
					Seventh        struct {
						NameSliceMap []map[string]byte
						NameMapSlice map[string][]interface{}
					}
				}
			}
		}
		Fifth []struct {
			NameHello string
			Sixth     map[string][]map[string]struct {
				Seventh map[string][]struct {
					NameWorld interface{}
				}
			}
		}
	}

	type privateStruct struct {
		Name string
	}

	testCases := []struct {
		Name   string
		Struct interface{}
		Keys   []string
	}{
		{
			Name:   "Empty Struct",
			Struct: struct{}{},
			Keys:   nil,
		},
		{
			Name: "Single Field",
			Struct: struct {
				Name string
			}{},
			Keys: []string{".name"},
		},
		{
			Name:   "Private Struct",
			Struct: privateStruct{},
			Keys:   nil,
		},
		{
			Name:   "Embedded Private Struct",
			Struct: struct{ privateStruct }{},
			Keys:   nil,
		},
		{
			Name: "Single Ptr Field",
			Struct: struct {
				Name *string
			}{},
			Keys: []string{".name"},
		},
		{
			Name: "Hidden Field",
			Struct: struct {
				name string
			}{},
			Keys: nil,
		},
		{
			Name: "JSON Tag Override",
			Struct: struct {
				Hello struct {
					Name string `json:"world"`
				}
			}{},
			Keys: []string{".hello.world"},
		},
		{
			Name: "JSON Tag Override With OmitEmpty",
			Struct: struct {
				Hello struct {
					Name string `json:"world,omitempty"`
				}
			}{},
			Keys: []string{".hello.world"},
		},
		{
			Name: "JSON Tag With Only OmitEmpty",
			Struct: struct {
				Hello struct {
					Name string `json:",omitempty"`
				}
			}{},
			Keys: []string{".hello.name"},
		},
		{
			Name: "Single, Ptr, and Hidden Field",
			Struct: struct {
				Hello string
				my    string
				World *string
			}{},
			Keys: []string{".hello", ".world"},
		},
		{
			Name: "Slice",
			Struct: struct {
				Hello []string
			}{},
			Keys: []string{".hello"},
		},
		{
			Name: "Map",
			Struct: struct {
				Hello map[string]string
			}{},
			Keys: []string{".hello"},
		},
		{
			Name: "Slice Slice",
			Struct: struct {
				Hello [][]string
			}{},
			Keys: []string{".hello"},
		},
		{
			Name: "Map Map",
			Struct: struct {
				Hello map[string]map[string]string
			}{},
			Keys: []string{".hello"},
		},
		{
			Name: "Slice Map",
			Struct: struct {
				Hello []map[string]string
			}{},
			Keys: []string{".hello"},
		},
		{
			Name: "Map Slice",
			Struct: struct {
				Hello map[string][]string
			}{},
			Keys: []string{".hello"},
		},
		{
			Name: "Complex Without Structs",
			Struct: struct {
				Name           int
				NamePtr        *float64
				hidden         string
				NameSlice      []rune
				NameMap        map[string]string
				NameSliceSlice [][]float64
				NameMapMap     map[string]map[string]int64
				NameSliceMap   []map[string]byte
				NameMapSlice   map[string][]interface{}
			}{},
			Keys: []string{".name", ".namePtr", ".nameSlice", ".nameMap", ".nameSliceSlice", ".nameMapMap", ".nameSliceMap", ".nameMapSlice"},
		},
		{
			Name: "Struct",
			Struct: struct {
				Hello struct {
					World string
				}
			}{},
			Keys: []string{".hello.world"},
		},
		{
			Name: "Struct Slice",
			Struct: struct {
				Hello struct {
					World []string
				}
			}{},
			Keys: []string{".hello.world"},
		},
		{
			Name: "Struct Map",
			Struct: struct {
				Hello struct {
					World map[string]string
				}
			}{},
			Keys: []string{".hello.world"},
		},
		{
			Name: "Struct Struct",
			Struct: struct {
				Hello struct {
					World struct {
						Rancher string
					}
				}
			}{},
			Keys: []string{".hello.world.rancher"},
		},
		{
			Name: "Struct List Struct",
			Struct: struct {
				Hello struct {
					World []struct {
						Rancher string
					}
				}
			}{},
			Keys: []string{".hello.world[].rancher"},
		},
		{
			Name: "Struct Map Struct",
			Struct: struct {
				Hello struct {
					World map[string]struct {
						Rancher string
					}
				}
			}{},
			Keys: []string{".hello.world[].rancher"},
		},
		{
			Name: "Struct List List Struct",
			Struct: struct {
				Hello struct {
					World [][]struct {
						Rancher string
					}
				}
			}{},
			Keys: []string{".hello.world[][].rancher"},
		},
		{
			Name: "Struct Map Map Struct",
			Struct: struct {
				Hello struct {
					World map[string]map[string]struct {
						Rancher string
					}
				}
			}{},
			Keys: []string{".hello.world[][].rancher"},
		},
		{
			Name: "Struct List Map Struct",
			Struct: struct {
				Hello struct {
					World []map[string]struct {
						Rancher string
					}
				}
			}{},
			Keys: []string{".hello.world[][].rancher"},
		},
		{
			Name: "Struct Map List Struct",
			Struct: struct {
				Hello struct {
					World map[string][]struct {
						Rancher string
					}
				}
			}{},
			Keys: []string{".hello.world[][].rancher"},
		},
		{
			Name:   "Exported Complex With Structs",
			Struct: ComplexWithStructs{},
			Keys: []string{
				".first.name",
				".second.third.namePtr",
				".fourth.nameSlice",
				".fourth.fifth.nameMap",
				".fourth.fifth.sixth.nameSliceSlice", ".fourth.fifth.sixth.nameMapMap",
				".fourth.fifth.sixth.seventh.nameSliceMap", ".fourth.fifth.sixth.seventh.nameMapSlice",
				".fifth[].nameHello", ".fifth[].sixth[][][].seventh[][].nameWorld",
			},
		},
		{
			Name:   "Exported Complex With Structs Anonymously Embedded Without JSON Tag",
			Struct: struct{ ComplexWithStructs }{},
			Keys: []string{
				".first.name",
				".second.third.namePtr",
				".fourth.nameSlice",
				".fourth.fifth.nameMap",
				".fourth.fifth.sixth.nameSliceSlice", ".fourth.fifth.sixth.nameMapMap",
				".fourth.fifth.sixth.seventh.nameSliceMap", ".fourth.fifth.sixth.seventh.nameMapSlice",
				".fifth[].nameHello", ".fifth[].sixth[][][].seventh[][].nameWorld",
			},
		},
		{
			Name: "Exported Complex With Structs Anonymously Embedded With JSON Tag",
			Struct: struct {
				ComplexWithStructs `json:"complex"`
			}{},
			Keys: []string{
				".complex.first.name",
				".complex.second.third.namePtr",
				".complex.fourth.nameSlice",
				".complex.fourth.fifth.nameMap",
				".complex.fourth.fifth.sixth.nameSliceSlice", ".complex.fourth.fifth.sixth.nameMapMap",
				".complex.fourth.fifth.sixth.seventh.nameSliceMap", ".complex.fourth.fifth.sixth.seventh.nameMapSlice",
				".complex.fifth[].nameHello", ".complex.fifth[].sixth[][][].seventh[][].nameWorld",
			},
		},
		{
			Name:   "Private Complex With Structs",
			Struct: complexWithStructs{},
			Keys:   nil,
		},
		{
			Name:   "Private Anonymous Complex With Structs Anonymously Embedded Without JSON Tag",
			Struct: struct{ complexWithStructs }{},
			Keys:   nil,
		},
		{
			Name: "Private Anonymous Complex With Structs Anonymously Embedded With JSON Tag",
			Struct: struct {
				complexWithStructs `json:"complex"`
			}{},
			Keys: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			keyMap := getAllKeysFromStructType(reflect.TypeOf(tc.Struct))
			expectedKeyMap := make(map[string]bool, len(tc.Keys))
			for _, k := range tc.Keys {
				expectedKeyMap[k] = true
			}
			assert.Equal(t, expectedKeyMap, keyMap)
		})
	}
}
