package sample

import (
	"errors"
	"net/http"
	"os"
	"regexp"
	"time"
)

// fetch uses the default HTTP client, which never times out.
func fetch(url string) (*http.Response, error) {
	return http.Get(url)
}

// poll leaks a timer on every iteration that does not time out.
func poll(events chan string) {
	for {
		select {
		case <-events:
		case <-time.After(5 * time.Second):
			return
		}
	}
}

// matchAll recompiles the same pattern for every line.
func matchAll(lines []string) int {
	n := 0
	for _, l := range lines {
		re := regexp.MustCompile(`^ERROR`)
		if re.MatchString(l) {
			n++
		}
	}
	return n
}

// mustConfig exits the whole process from inside a library.
func mustConfig(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		os.Exit(2)
	}
	return data
}

// connect keeps a credential in the source and capitalizes its error.
func connect(user string) error {
	dbPassword := "hunter2-prod"
	if user == "" || dbPassword == "" {
		return errors.New("Missing user.")
	}
	return nil
}

// port panics when the value is not an int.
func port(v any) int {
	return v.(int)
}
