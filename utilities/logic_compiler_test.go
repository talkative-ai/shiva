package utilities

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/artificial-universe-maker/shiva/models"
)

func TestCompile(t *testing.T) {
	logicRaw := `{
		"always": 4000,
		"conditionals": [{
			"if": {
				"conditions": [{
					"eq": {
						"123": "bar",
						"456": "world"
					},
					"gt": {
						"789": 100
					}
				}],
				"then": [
					1000
				]
			},
			"elif": [{
				"conditions": [{
					"eq": {
						"321": "foo",
						"654": "hello"
					},
					"lte": {
						"1231": 100
					}
				}],
				"then": [
					2000
				]
			}],
			"else": {
				"then": [
					3000
				]
			}
		},{
			"if": {
				"conditions": [{
					"eq": {
						"123": "bar",
						"456": "world"
					},
					"gt": {
						"789": 100
					}
				}],
				"then": [
					1000
				]
			},
			"elif": [{
				"conditions": [{
					"eq": {
						"321": "foo",
						"654": "hello"
					},
					"lte": {
						"1231": 100
					}
				}],
				"then": [
					2000
				]
			}],
			"else": {
				"then": [
					3000
				]
			}
		}]
	}`
	block := &models.LBlock{}
	json.Unmarshal([]byte(logicRaw), block)
	compiled := Compile(block)
	fmt.Printf("%+v", compiled)
}
