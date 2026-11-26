package log_test

import (
	"errors"
	"fmt"
	"strings"

	"github.com/moov-io/base/log"
)

func ExampleLogger() {
	buf, logger := log.NewBufferLogger()

	logger.Info().Log("Application started")
	logger.Info().Logf("Processing request %s", "12345")
	logger.LogError(errors.New("boom"))
	logger.Error().LogError(errors.New("boom"))

	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		fmt.Println(fields(line))
	}
	// Output:
	// msg="Application started" level=info
	// msg="Processing request 12345" level=info
	// msg=boom errored=true level=info
	// msg=boom errored=true level=error
}

func ExampleStructContext() {
	type Address struct {
		Street  string `log:"street"`
		City    string `log:"city"`
		Country string `log:"country"`
	}
	type User struct {
		ID      int     `log:"id"`
		Address Address `log:"address"`
	}

	buf, logger := log.NewBufferLogger()
	user := User{
		ID: 1,
		Address: Address{
			Street:  "123 Main St",
			City:    "New York",
			Country: "USA",
		},
	}
	logger.With(log.StructContext(user)).Info().Log("User logged in")
	fmt.Println(fields(buf.String()))
	// Output:
	// msg="User logged in" address="{123 Main St New York USA}" address.city="New York" address.country=USA address.street="123 Main St" id=1 level=info
}

// fields drops the timestamp and keeps logfmt tokens in order.
func fields(line string) string {
	var out []string
	s := strings.TrimSpace(line)
	for len(s) > 0 {
		eq := strings.IndexByte(s, '=')
		if eq < 0 {
			out = append(out, s)
			break
		}
		end := eq + 1
		if end < len(s) && s[end] == '"' {
			end++
			for end < len(s) && s[end] != '"' {
				end++
			}
			if end < len(s) {
				end++
			}
		} else {
			for end < len(s) && s[end] != ' ' {
				end++
			}
		}
		tok := s[:end]
		if !strings.HasPrefix(tok, "ts=") {
			out = append(out, tok)
		}
		s = strings.TrimSpace(s[end:])
	}
	return strings.Join(out, " ")
}
