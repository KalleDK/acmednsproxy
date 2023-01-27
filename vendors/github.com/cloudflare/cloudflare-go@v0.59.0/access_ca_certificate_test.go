package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAcessCACertificate(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "result": {
    "id": "4f74df465b2a53271d4219ac2ce2598e24b5e2c60c7924f4",
    "aud": "7d1996154eb606c19e31dd777fe6981f57a5ab66735c5c00fefd01b1200ba9d0",
    "public_key": "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTI...3urg/XpGMdgaSs5ZdptUPw= open-ssh-ca@cloudflareaccess.org"
  },
  "success": true,
  "errors": [],
  "messages": []
}
		`)
	}

	want := AccessCACertificate{
		ID:        "4f74df465b2a53271d4219ac2ce2598e24b5e2c60c7924f4",
		Aud:       "7d1996154eb606c19e31dd777fe6981f57a5ab66735c5c00fefd01b1200ba9d0",
		PublicKey: "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTI...3urg/XpGMdgaSs5ZdptUPw= open-ssh-ca@cloudflareaccess.org",
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps/f174e90a-fafe-4643-bbbc-4a0ed4fc8415/ca", handler)

	actual, err := client.AccessCACertificate(context.Background(), testAccountID, "f174e90a-fafe-4643-bbbc-4a0ed4fc8415")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/apps/f174e90a-fafe-4643-bbbc-4a0ed4fc8415/ca", handler)

	actual, err = client.ZoneLevelAccessCACertificate(context.Background(), testZoneID, "f174e90a-fafe-4643-bbbc-4a0ed4fc8415")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestAcessCACertificates(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "result": [{
    "id": "4f74df465b2a53271d4219ac2ce2598e24b5e2c60c7924f4",
    "aud": "7d1996154eb606c19e31dd777fe6981f57a5ab66735c5c00fefd01b1200ba9d0",
    "public_key": "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTI...3urg/XpGMdgaSs5ZdptUPw= open-ssh-ca@cloudflareaccess.org"
  }],
  "success": true,
  "errors": [],
  "messages": []
}
		`)
	}

	want := []AccessCACertificate{{
		ID:        "4f74df465b2a53271d4219ac2ce2598e24b5e2c60c7924f4",
		Aud:       "7d1996154eb606c19e31dd777fe6981f57a5ab66735c5c00fefd01b1200ba9d0",
		PublicKey: "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTI...3urg/XpGMdgaSs5ZdptUPw= open-ssh-ca@cloudflareaccess.org",
	}}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps/ca", handler)

	actual, err := client.AccessCACertificates(context.Background(), testAccountID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/apps/ca", handler)

	actual, err = client.ZoneLevelAccessCACertificates(context.Background(), testZoneID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateAcessCACertificates(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "result": {
    "id": "4f74df465b2a53271d4219ac2ce2598e24b5e2c60c7924f4",
    "aud": "7d1996154eb606c19e31dd777fe6981f57a5ab66735c5c00fefd01b1200ba9d0",
    "public_key": "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTI...3urg/XpGMdgaSs5ZdptUPw= open-ssh-ca@cloudflareaccess.org"
  },
  "success": true,
  "errors": [],
  "messages": []
}
		`)
	}

	want := AccessCACertificate{
		ID:        "4f74df465b2a53271d4219ac2ce2598e24b5e2c60c7924f4",
		Aud:       "7d1996154eb606c19e31dd777fe6981f57a5ab66735c5c00fefd01b1200ba9d0",
		PublicKey: "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTI...3urg/XpGMdgaSs5ZdptUPw= open-ssh-ca@cloudflareaccess.org",
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps/f174e90a-fafe-4643-bbbc-4a0ed4fc8415/ca", handler)

	actual, err := client.CreateAccessCACertificate(context.Background(), testAccountID, "f174e90a-fafe-4643-bbbc-4a0ed4fc8415")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/apps/f174e90a-fafe-4643-bbbc-4a0ed4fc8415/ca", handler)

	actual, err = client.CreateZoneLevelAccessCACertificate(context.Background(), testZoneID, "f174e90a-fafe-4643-bbbc-4a0ed4fc8415")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteAcessCACertificates(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "result": {
    "id": "4f74df465b2a53271d4219ac2ce2598e24b5e2c60c7924f4"
  },
  "success": true,
  "errors": [],
  "messages": []
}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps/f174e90a-fafe-4643-bbbc-4a0ed4fc8415/ca", handler)

	err := client.DeleteAccessCACertificate(context.Background(), testAccountID, "f174e90a-fafe-4643-bbbc-4a0ed4fc8415")

	assert.NoError(t, err)

	mux.HandleFunc("/zones/"+testZoneID+"/access/apps/f174e90a-fafe-4643-bbbc-4a0ed4fc8415/ca", handler)

	err = client.DeleteZoneLevelAccessCACertificate(context.Background(), testZoneID, "f174e90a-fafe-4643-bbbc-4a0ed4fc8415")

	assert.NoError(t, err)
}
