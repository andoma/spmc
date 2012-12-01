
package main

import "os"
import "io"

import "encoding/json"
import "encoding/base64"
import "crypto/sha1"
import "crypto/hmac"
import "crypto/rand"
import "fmt"
import "strings"

var (
	cookie_key[32] byte;
)

func init() {
	cklen := 32;
	cookiekeyfile := "cookiekey.txt";
	file, err := os.Open(cookiekeyfile);
	if err != nil {
		n, err := io.ReadFull(rand.Reader, cookie_key[:]);
		if err != nil || n != cklen {
			fmt.Printf("Unable to generate cookie key\n");
			os.Exit(1);
		}
		file, err = os.Create(cookiekeyfile);
		if err != nil {
			fmt.Printf("Unable to create cookie key file\n");
			os.Exit(1);
		}
		file.Write(cookie_key[:]);
		file.Chmod(0400);
		file.Close();
	} else {
		n, err := io.ReadFull(file, cookie_key[:]);
		if err != nil || n != cklen {
			fmt.Printf("Unable to read cookie key\n");
			os.Exit(1);
		}
	}
}


// validate user credentials and return a cookie
func authenticateUser(u *User) (*string, error) {
	js, err := json.Marshal(u);
	if err != nil {
		return nil, err;
	}
	b64 := base64.StdEncoding.EncodeToString([]byte(js));
	h := hmac.New(sha1.New, cookie_key[:]);
	h.Write([]byte(b64));
	sig := base64.StdEncoding.EncodeToString(h.Sum(nil));
	cookie := fmt.Sprintf("%s_%s\n", b64, sig);
	return &cookie, nil;
}


// validate user credentials and return a cookie
func validateCookie(cookie string) (*User) {
	c := strings.Split(cookie, "_");
	if len(c) != 2 {
		return nil;
	}
	h := hmac.New(sha1.New, cookie_key[:]);
	h.Write([]byte(c[0]));
	sig := base64.StdEncoding.EncodeToString(h.Sum(nil));

	if sig != c[1] {
		return nil;
	}

	js, err := base64.StdEncoding.DecodeString(c[0]);
	if err != nil {
		return nil;
	}

	u := new(User);
	err = json.Unmarshal(js, u);
	if err != nil {
		return nil;
	}

	return u;
}