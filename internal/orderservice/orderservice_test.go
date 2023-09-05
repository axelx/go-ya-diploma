package orderservice

import (
	"fmt"
	"github.com/axelx/go-ya-diploma/internal/logger"
	"testing"
)

func TestOrderLunaCheck(t *testing.T) {
	fmt.Println("hello test order")

	lg := logger.Initialize("fatal")
	os := Order{LG: lg}

	var tests = []struct {
		name  string
		input string
		want  bool
	}{
		{name: "Correct number", input: "91", want: true},
		{name: "Non correct number", input: "92", want: false},
		{name: "zero", input: "0", want: false},
		{name: "alfabit", input: "asdf", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans := os.LunaCheck(tt.input)
			if ans != tt.want {
				t.Errorf("got #(ans), want #{tt.want]")
			}
		})
	}

}
