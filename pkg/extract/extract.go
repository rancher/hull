package extract

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/rancher/wrangler/v3/pkg/data/convert"
)

var renderValuesCmdRe = regexp.MustCompile(`(?P<field>[^\[\]]*)+(?P<indices>\[.*\])*`)

func Field[O interface{}](obj interface{}, path string) (O, bool) {
	var ok bool
	for _, cmd := range strings.Split(path, ".") {
		// process 'Values' in something like 'Values[0][1][2]'
		matches := renderValuesCmdRe.FindStringSubmatch(cmd)
		fieldIdx := renderValuesCmdRe.SubexpIndex("field")
		field := matches[fieldIdx]
		if len(field) > 0 {
			obj, ok = getValueFromObject(obj, field)
			if !ok {
				return *new(O), false
			}
		}
		// process all the other indices
		indicesIdx := renderValuesCmdRe.SubexpIndex("indices")
		multipleMatches := renderValuesCmdRe.FindAllStringSubmatch(cmd, -1)
		for _, matches := range multipleMatches {
			indicesString := matches[indicesIdx]
			if len(indicesString) < 2 {
				// no match found for index
				continue
			}
			indices := strings.Split(indicesString[1:len(indicesString)-1], "][")
			for _, index := range indices {
				if index == "" {
					return *new(O), false
				}
				hasDoubleQuotes := strings.HasPrefix(index, `"`) && strings.HasSuffix(index, `"`)
				hasSingleQuotes := strings.HasPrefix(index, `'`) && strings.HasSuffix(index, `'`)
				if hasDoubleQuotes || hasSingleQuotes {
					// found string index for map
					field = index[1 : len(index)-1]
					obj, ok = getValueFromObject(obj, field)
					if !ok {
						return *new(O), false
					}
				} else {
					// should be integer index
					index, err := strconv.Atoi(index)
					if err != nil {
						return *new(O), false
					}
					obj, ok = getValueFromSlice(obj, index)
					if !ok {
						return *new(O), false
					}
				}
			}
		}
	}
	typedObj, ok := obj.(O)
	if !ok {
		// try marshalling and unmarshalling
		err := convert.ToObj(obj, &typedObj)
		if err != nil {
			return *new(O), false
		}
	}
	return typedObj, true
}

func getValueFromObject(renderValues interface{}, key string) (interface{}, bool) {
	if renderValues == nil {
		return nil, false
	}
	var val reflect.Value
	switch reflect.TypeOf(renderValues).Kind() {
	case reflect.Map:
		val = reflect.ValueOf(renderValues).MapIndex(reflect.ValueOf(key))
	case reflect.Struct:
		val = reflect.ValueOf(renderValues).FieldByName(key)
	case reflect.Pointer:
		underlyingRenderVal := reflect.Indirect(reflect.ValueOf(renderValues))
		if underlyingRenderVal.Kind() == reflect.Invalid {
			return nil, false
		}
		return getValueFromObject(underlyingRenderVal.Interface(), key)
	default:
		return nil, false
	}
	if val.Kind() == reflect.Invalid {
		return nil, false
	}
	if !val.CanInterface() {
		// might be a hidden field in a struct
		return nil, false
	}
	return val.Interface(), true
}

func getValueFromSlice(renderValues interface{}, index int) (interface{}, bool) {
	if renderValues == nil {
		return nil, false
	}
	var val reflect.Value
	switch reflect.TypeOf(renderValues).Kind() {
	case reflect.Slice, reflect.Array:
		renderValuesVal := reflect.ValueOf(renderValues)
		if index >= renderValuesVal.Len() {
			return nil, false
		}
		val = renderValuesVal.Index(index)
	default:
		return nil, false
	}
	return val.Interface(), true
}
