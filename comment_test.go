package comment

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"
)

func TestCollectStructComment(t *testing.T) {
	for _, cas := range []interface{}{
		&User{},
	} {
		s, err := ResolveStruct(cas)
		if err != nil {
			t.Fatal(err)
		}
		jsonPrint(os.Stdout, s)
	}
}

func jsonPrint(w io.Writer, in interface{}) {
	var data []byte
	if v, ok := in.([]byte); ok {
		data = v
	} else {
		var err error
		data, err = json.Marshal(in)
		if err != nil {
			panic(err)
		}
	}
	var buf = new(bytes.Buffer)
	if err := json.Indent(buf, data, "", "\t"); err != nil {
		panic(err)
	}
	if _, err := buf.WriteTo(w); err != nil {
		panic(err)
	}
}
