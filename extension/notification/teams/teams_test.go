package teams

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/quilla-hq/quilla/constants"
	"github.com/quilla-hq/quilla/types"
	"github.com/quilla-hq/quilla/version"
)

func TestTrimLeftChar(t *testing.T) {
	fmt.Printf("%q\n", "Hello, 世界")
	fmt.Printf("%q\n", TrimFirstChar(""))
	fmt.Printf("%q\n", TrimFirstChar("H"))
	fmt.Printf("%q\n", TrimFirstChar("世"))
	fmt.Printf("%q\n", TrimFirstChar("Hello"))
	fmt.Printf("%q\n", TrimFirstChar("世界"))
}

func TestTeamsRequest(t *testing.T) {
	handler := func(resp http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			t.Errorf("failed to parse body: %s", err)
		}

		bodyStr := string(body)

		if !strings.Contains(bodyStr, "MessageCard") {
			t.Errorf("missing MessageCard indicator")
		}

		if !strings.Contains(bodyStr, "themeColor") {
			t.Errorf("missing themeColor")
		}

		if !strings.Contains(bodyStr, constants.QuillaLogoURL) {
			t.Errorf("missing logo url")
		}

		if !strings.Contains(bodyStr, "**"+types.NotificationPreDeploymentUpdate.String()+"**") {
			t.Errorf("missing deployment type")
		}

		if !strings.Contains(bodyStr, version.GetquillaVersion().Version) {
			t.Errorf("missing version")
		}

		if !strings.Contains(bodyStr, "update deployment") {
			t.Errorf("missing name")
		}
		if !strings.Contains(bodyStr, "message here") {
			t.Errorf("missing message")
		}

		t.Log(bodyStr)

	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	s := &sender{
		endpoint: ts.URL,
		client:   &http.Client{},
	}

	s.Send(types.EventNotification{
		Name:    "update deployment",
		Message: "message here",
		Type:    types.NotificationPreDeploymentUpdate,
	})
}
