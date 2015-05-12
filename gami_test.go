package gami

import (
	"../gami"
	"fmt"
	"os"
	"testing"
)

func InitRunClose(url string, tb testing.TB, fn func(*gami.AMIClient)) {
	ami, err := gami.Dial(url)
	if err != nil {
		tb.Fatal(err)
	}
	defer func() {
		fmt.Println("Closing")
		ami.Close()
		fmt.Println("Closed")
	}()
	if fn != nil {
		fn(ami)
	}
}

func GetUrl() (url string) {
	url = os.Getenv("TEST_ASTERISK_URL")
	if url == "" {
		url = "127.0.0.1:5038"
	}
	return url
}

func GetCredentials() (user string, pass string) {
	user = os.Getenv("TEST_ASTERISK_USER")
	pass = os.Getenv("TEST_ASTERISK_PASS")
	if user == "" {
		user = "manager"
	}
	if pass == "" {
		pass = "1234"
	}
	return user, pass
}

func Login(tb testing.TB, client *gami.AMIClient) {
	user, pass := GetCredentials()
	if err := client.Login(user, pass); err != nil {
		fmt.Println("fatal")
		tb.Fatal(err)
	}
}

/*
func TestInitiation(t *testing.T) {
	InitRunClose(GetUrl(), t, nil)
}
*/

func TestSeveralCommands(t *testing.T) {
	InitRunClose(GetUrl(), t, func(client *gami.AMIClient) {
		Login(t, client)

		if rs, err := client.Action("ListCommands", nil); err != nil {
			t.Fatal(err)
		} else {
			fmt.Println("response", rs)
		}
	})
}

func BenchmarkInitRunClose(b *testing.B) {
	url := GetUrl()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		InitRunClose(url, b, nil)
	}
}
