package rollbar

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"
)

type CustomError struct {
	s string
}

func (e *CustomError) Error() string {
	return e.s
}

func testErrorStack(s string) {
	testErrorStack2(s)
}

func testErrorStack2(s string) {
	Error("error", errors.New(s))
}

func testErrorStackWithSkip(s string) {
	testErrorStackWithSkip2(s)
}

func testErrorStackWithSkip2(s string) {
	ErrorWithStackSkip("error", errors.New(s), 2)
}

func TestErrorClass(t *testing.T) {
	errors := map[string]error{
		"{508e076d}":          fmt.Errorf("Something is broken!"),
		"rollbar.CustomError": &CustomError{"Terrible mistakes were made."},
	}

	for expected, err := range errors {
		if errorClass(err) != expected {
			t.Error("Got:", errorClass(err), "Expected:", expected)
		}
	}
}

func TestEverything(t *testing.T) {
	Token = os.Getenv("TOKEN")
	Environment = "test"

	Error("critical", errors.New("Normal critical error"))
	Error("error", &CustomError{"This is a custom error"})

	testErrorStack("This error should have a nice stacktrace")
	testErrorStackWithSkip("This error should have a skipped stacktrace")

	done := make(chan bool)
	go func() {
		testErrorStack("I'm in a goroutine")
		done <- true
	}()
	<-done

	Message("error", "This is an error message")
	Message("info", "And this is an info message")

	// If you don't see the message sent on line 65 in Rollbar, that means this
	// is broken:
	Wait()
}

func TestErrorRequest(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://foo.com/somethere?param1=true", nil)
	r.RemoteAddr = "1.1.1.1:123"

	object := errorRequest(r)

	if object["url"] != "http://foo.com/somethere?param1=true" {
		t.Errorf("wrong url, got %v", object["url"])
	}

	if object["method"] != "GET" {
		t.Errorf("wrong method, got %v", object["method"])
	}

	if object["query_string"] != "param1=true" {
		t.Errorf("wrong id, got %v", object["query_string"])
	}
}

func TestScrubValues(t *testing.T) {
	values := map[string][]string{
		"password":     []string{"one"},
		"toallyfine":   []string{"one"},
		"sswo":         []string{"one"},
		"access_token": []string{"one"},
	}

	scrubbed := scrubValues(values)
	if scrubbed["password"][0] != "------" {
		t.Error("should scrub password parameter")
	}

	if scrubbed["toallyfine"][0] == "------" {
		t.Error("should keep toallyfine parameter")
	}

	// substring of scrubField
	if scrubbed["sswo"][0] == "------" {
		t.Error("should keep sswo parameter")
	}

	if scrubbed["access_token"][0] != "------" {
		t.Error("should keep access_token parameter")
	}
}

func TestFlattenValues(t *testing.T) {
	values := map[string][]string{
		"a": []string{"one"},
		"b": []string{"one", "two"},
	}

	flattened := flattenValues(values)
	if flattened["a"].(string) != "one" {
		t.Error("should flatten single parameter to string")
	}

	if len(flattened["b"].([]string)) != 2 {
		t.Error("should leave multiple parametres as []string")
	}
}
