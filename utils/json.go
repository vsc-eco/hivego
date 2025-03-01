package utils

import jsoniter "github.com/json-iterator/go"

//A high-performance 100% compatible drop-in replacement of "encoding/json"
//see here : https://github.com/json-iterator/go

var json = jsoniter.ConfigFastest
