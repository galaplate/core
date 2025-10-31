package supports

import (
	"encoding/json"
	"fmt"
	"os"
)

func Dump(arg ...any) {
	for _, a := range arg {
		if jsonBytes, err := json.MarshalIndent(a, "", "  "); err == nil {
			fmt.Printf("%s\n", jsonBytes)
		} else {
			fmt.Printf("%+v\n", a)
		}
	}
}

func DD(arg ...any) {
	for _, a := range arg {
		if jsonBytes, err := json.MarshalIndent(a, "", "  "); err == nil {
			fmt.Printf("%s\n", jsonBytes)
		} else {
			fmt.Printf("%+v\n", a)
		}
	}
	os.Exit(1)
}
