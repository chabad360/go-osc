package main

import (
	"fmt"
	"github.com/chabad360/go-osc/osc"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	files, err := ioutil.ReadDir(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	i := 20
	fmt.Printf("// Code generated by testcasegen.go DO NOT EDIT\n\n")
	fmt.Printf("type testCase struct {\n\tname string\n\tobj Packet\n\traw []byte\n\twantErr bool\n}\n\nvar (\n\t%sTestCases = []testCase{\n", strings.ToLower(os.Args[2]))
	for _, file := range files {
		if strings.HasPrefix(file.Name(), os.Args[2]) {
			b, err := ioutil.ReadFile(os.Args[1] + file.Name())
			if err != nil {
				panic(err)
			}

			m, err := osc.ParsePacket(b)
			i++
			name := file.Name()

			a, s := argGen([]interface{}{m})
			s = strings.TrimPrefix(s, " (")
			s = strings.TrimSuffix(s, ")")
			fmt.Printf("\t\t{\"%s\"%s,\n\t\t\t%#v, %t}, // %s\n", name, a, b, err != nil, s)
		}
	}
	fmt.Printf("\t}\n)")
}

func argGen(i []interface{}) (args string, str string) {
	for _, v := range i {
		args += ", "
		str += " "
		switch t := v.(type) {
		case bool:
			if t {
				args += "true"
			} else {
				args += "false"
			}
		case nil:
			args += "nil"
		case int32:
			args += fmt.Sprintf("int32(%d)", t)
			str += fmt.Sprintf("%d", t)
		case int64:
			args += fmt.Sprintf("int64(%d)", t)
			str += fmt.Sprintf("%d", t)
		case float32:
			args += fmt.Sprintf("float32(%f)", t)
			str += fmt.Sprintf("%f", t)
		case float64:
			args += fmt.Sprintf("float64(%f)", t)
			str += fmt.Sprintf("%f", t)
		case string:
			args += fmt.Sprintf("\"%s\"", t)
			str += fmt.Sprintf("\"%s\"", t)
		case []byte:
			args += fmt.Sprintf("%#v", t)
			str += fmt.Sprintf("%v", t)
		case osc.Timetag:
			args += fmt.Sprintf("\"%d\"", t)
			str += fmt.Sprintf("T:%d.%d", int64(t>>32), int64(t&0xffffffff))
		case *osc.Message:
			a, s := argGen(t.Arguments)
			tt, _ := osc.GetTypeTag(t.Arguments)
			args += fmt.Sprintf("NewMessage(\"%s\"%s)", t.Address, a)
			str += fmt.Sprintf("(%s %s%s)", t.Address, tt, s)
		case *osc.Bundle:
			ii := make([]interface{}, len(t.Elements))
			for v := range t.Elements {
				ii[v] = t.Elements[v]
			}
			a, s := argGen(ii)
			_, ss := argGen([]interface{}{t.Timetag})
			args += fmt.Sprintf("NewBundleWithTime(Timetag(%d).Time()%s)", t.Timetag, a)
			str += fmt.Sprintf("(#bundle%s%s)", ss, s)
		}
	}
	return
}
