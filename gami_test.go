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
		ami.Close()
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
func Login(client *gami.AMIClient) error {
	user, pass := GetCredentials()
	return client.Login(user, pass)
}

/*
func TestInitiation(t *testing.T) {
	InitRunClose(GetUrl(), t, nil)
}

func TestSeveralCommands(t *testing.T) {
	InitRunClose(GetUrl(), t, func(client *gami.AMIClient) {
		if err := Login(client); err != nil {
			t.Fatal(err)
		}

		if _, err := client.Action("ListCommands", nil); err != nil {
			t.Fatal(err)
		}
		if _, err := client.Action("ListCommands", nil); err != nil {
			t.Fatal(err)
		}
	})
}
*/

func TestReconnection(t *testing.T) {
	fmt.Println("Rec1")
	url := GetUrl()
	ami, err := gami.Dial(url)
	if err != nil {
		t.Fatal(err)
	}
	defer ami.Close()
	fmt.Println("Rec2")
	ami.Reconnect()
	fmt.Println("Rec3")
	Login(ami)
	fmt.Println("Rec4")
	ami.Reconnect()
	fmt.Println("Rec5")
}

func BenchmarkInitRunClose(b *testing.B) {
	url := GetUrl()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		InitRunClose(url, b, nil)
	}
}

func BenchmarkOperations(b *testing.B) {
	url := GetUrl()
	b.ReportAllocs()
	InitRunClose(url, b, func(client *gami.AMIClient) {
		if err := Login(client); err != nil {
			b.Fatal(err)
		}
		for i := 0; i < b.N; i++ {
			client.Action("Ping", nil)
		}
	})
}
