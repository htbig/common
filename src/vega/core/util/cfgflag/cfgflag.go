// Copyright (c) 2016, Virtual Gateway Labs. All rights reserved.

// Package cfgflag provide APIs for compare configs
package cfgflag

import (
	"fmt"
	"reflect"
)

const (
	OP_ADD = iota
	OP_UPDATE
	OP_DEL
	OP_NORMAL
)

type (
	ChangeMap map[string] int
	ChangeListMap map[string] ChangeMap
)

type ConfigFlag struct {
	flag ChangeMap
	listFlag ChangeListMap
}

var is_initial bool = false

func IsInitial() bool {
	return is_initial
}

func SetInitialFlag( val bool ) {
	is_initial = val
}

func (cfgflag *ConfigFlag) Init() {
	cfgflag.flag = make(ChangeMap)
	cfgflag.listFlag = make(ChangeListMap)
}

func (cfgflag ConfigFlag) GetFlag( key string ) int {
	return cfgflag.flag[key]
}

func (cfgflag ConfigFlag) GetListFlag( key string ) ChangeMap {
	return cfgflag.listFlag[key]
}

func (cfgflag ConfigFlag) Print() {
	for key, val := range cfgflag.flag {
		fmt.Println(key, val)
	}

	for key, val := range cfgflag.listFlag {
		fmt.Println(key, val)
	}
}

func (cfgflag *ConfigFlag) UpdateFlag(old_data, new_data interface{} ) {
	if ( reflect.TypeOf(old_data) != reflect.TypeOf(new_data) ) {
		fmt.Println(reflect.TypeOf(old_data))
		fmt.Println(reflect.TypeOf(new_data))
		fmt.Println("type error")
		return
	}

	v1_elem := reflect.ValueOf(old_data).Elem()
	v2_elem := reflect.ValueOf(new_data).Elem()

	for i := 0; i < v1_elem.NumField(); i++ {
		v1_val := v1_elem.Field(i)
		v2_val := v2_elem.Field(i)
		v1_type := v1_elem.Type().Field(i)

		switch v1_val.Kind() {
		case reflect.Array:
			fmt.Println("Array")
		case reflect.Slice:
			fmt.Println("Slice")
        	v1_tag := v1_type.Tag
			for i := 0; i < v1_val.Len(); i++ {
				key_field := v1_tag.Get("key")
				key := v1_val.Index(i).FieldByName(key_field)
				fmt.Println(key_field, key)
			}

		case reflect.Map:
			fmt.Println("Map")
			cfgflag.listFlag[v1_type.Name] = make(ChangeMap)

			keys := v1_val.MapKeys()
			for _, key := range keys {
				if v2_val.MapIndex(key).IsValid() {
					// for update flag
					if reflect.DeepEqual(v1_val.MapIndex(key).Interface(), v2_val.MapIndex(key).Interface()) {
						cfgflag.listFlag[v1_type.Name][key.Interface().(string)] = OP_NORMAL
					} else {
						cfgflag.listFlag[v1_type.Name][key.Interface().(string)] = OP_UPDATE
					}
				} else {
					// for delete flag
					cfgflag.listFlag[v1_type.Name][key.Interface().(string)] = OP_DEL
				}
			}

			// for add flag
			keys = v2_val.MapKeys()
			for _, key := range keys {
				if !v1_val.MapIndex(key).IsValid() {
					cfgflag.listFlag[v1_type.Name][key.Interface().(string)] = OP_ADD
				}
			}

		default:
			// for scalar flag
			if reflect.DeepEqual(v1_val.Interface(), v2_val.Interface()) {
				cfgflag.flag[v1_type.Name] = OP_NORMAL
			} else {
				cfgflag.flag[v1_type.Name] = OP_UPDATE
			}
		}
	}

	return
}
