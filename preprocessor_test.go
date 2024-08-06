package socks

import (
	"bytes"
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/runtime"
	"io"
	"testing"
)

func TestPreprocessor(t *testing.T) {
	layout := `<html><head><title>Abc</title></head><body><v-slot name="content"><v-component name="clock.html"></v-component></v-slot></body></html>`
	index := `<v-component name="layout.html">
	<div :slot="content">
		<h1>Welcome on our page!</h1>
		<v-component name="clock.html"></v-component>
	</div>
</v-component>`
	clock := `<v-component name="flex.html"><button :slot="content" class="py-2 px-4 font-medium">Fetch time</button></v-component>`
	flex := `<div style="display: flex; justify-content: center; align-items: center;"><v-slot name="content"></v-slot></div>`

	preprocessed, err := Preprocess(map[string]io.Reader{"layout.html": bytes.NewBufferString(layout), "index.html": bytes.NewBufferString(index), "clock.html": bytes.NewBufferString(clock), "flex.html": bytes.NewBufferString(flex)}, nil, nil)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	output := bytes.NewBufferString("")
	fmt.Println(runtime.NewEvaluator(helpers.File{Name: "index.html"}, preprocessed["index.html"], nil).Evaluate(output, nil))

	fmt.Println(output.String())
}
