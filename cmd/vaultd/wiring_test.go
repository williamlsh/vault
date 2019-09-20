package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/discard"
	"github.com/williamlsh/vault/pkg/mock"
	"github.com/williamlsh/vault/pkg/vaultendpoint"
	"github.com/williamlsh/vault/pkg/vaultransport"
	"github.com/williamlsh/vault/pkg/vaultservice"
)

type testcase struct {
	method, url, body, want string
}

func TestHTTP(t *testing.T) {
	svc := vaultservice.New(log.NewNopLogger(), discard.NewCounter(), mock.NewNopStore())
	eps := vaultendpoint.New(svc, log.NewNopLogger(), discard.NewHistogram())
	mux := vaultransport.NewHTTPHandler(eps, log.NewNopLogger())
	srv := httptest.NewServer(mux)
	defer srv.Close()

	t.Run("hash password", func(t *testing.T) {
		caseHash := testcase{
			method: http.MethodPost,
			url:    srv.URL + "/hash",
			body:   `{"password":"znm9832nmrfz4egwy43rn8"}`,
		}
		req, err := http.NewRequest(caseHash.method, caseHash.url, strings.NewReader(caseHash.body))
		if err != nil {
			t.Fatal(err)
		}
		setHeader(req)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if body == nil {
			t.Fail()
		}
	})

	t.Run("validate hash", func(t *testing.T) {
		caseValidate := testcase{
			method: http.MethodPost,
			url:    srv.URL + "/validate",
			body:   `{"password":"znm9832nmrfz4egwy43rn8","hash":"$2a$10$8e4JwCH9mCppJpTQ3Ax1PevFIt79her0oOg7AFy3eA4BNoeOMX1w."}`,
			want:   `{"valid":true}`,
		}
		req, err := http.NewRequest(caseValidate.method, caseValidate.url, strings.NewReader(caseValidate.body))
		if err != nil {
			t.Fatal(err)
		}
		setHeader(req)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if want, have := caseValidate.want, strings.TrimSpace(string(body)); want != have {
			t.Errorf("%s %s %s: want %s, have %s", caseValidate.method, caseValidate.url, caseValidate.body, want, have)
		}
	})
}

func signTok() string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(1 * time.Second).Unix(),
	})
	ss, err := token.SignedString([]byte("zmh298onj30"))
	if err != nil {
		panic(err)
	}
	return ss
}

func setHeader(r *http.Request) {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", signTok()))
}
