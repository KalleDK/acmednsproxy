package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"gopkg.in/yaml.v3"
)

const valid_yaml = `---
type: jwtauth
jwk:
  - kty: EC
    kid: main
    crv: P-521
    x: 6346465687230269390528684383187047644613514507746409766845877836899676532978276903329258589750862559531134970938049999412427996900487716818247810682237199037
    y: 4093520538738021593174382625861913469571622768095628588302758201755113534369556547294692470492927352406337273388577715272437451619810567181552724517045209355
    d: 4153706707440949913606384900087906325438651336766954420064969985863633823834823212357817295068035995589365595721972506997857879319614542249900284917297714004
keys:
  - name: main
    priv: |
        -----BEGIN PRIVATE KEY-----
        MIHcAgEBBEIBNcwvdlKW5kgF1hFwZTVBcr0crZpT1cgC0xUG/Pcio4A3qiBLUJHs
        mwDFrlUQ0KkpAC7RFkRsnzoPyWDN+OkwE1SgBwYFK4EEACOhgYkDgYYABAHZV04X
        miG6iIeEgRGi/26esWVemQHLD51b6Rgtd+IKxKYBvJTu6L5SHhKMYE0PrmWBN1yV
        PsnQCyPEmtB0ElOivQExTweY/sEMBMiuMbu5lJ5duLFy393QFMc17V84gRl55plh
        nrgoYUDdYNE+eCm1/Ybum2p82QR9Qa+v9r2voTUJCw==
        -----END PRIVATE KEY-----
    pub: |
        -----BEGIN PUBLIC KEY-----
        MIGbMBAGByqGSM49AgEGBSuBBAAjA4GGAAQB2VdOF5ohuoiHhIERov9unrFlXpkB
        yw+dW+kYLXfiCsSmAbyU7ui+Uh4SjGBND65lgTdclT7J0AsjxJrQdBJTor0BMU8H
        mP7BDATIrjG7uZSeXbixct/d0BTHNe1fOIEZeeaZYZ64KGFA3WDRPngptf2G7ptq
        fNkEfUGvr/a9r6E1CQs=
        -----END PUBLIC KEY-----
`

type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
}

type JWKEC struct {
	JWK
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
	D   string `json:"d"`
}

func (j *JWKEC) GetKey() (*ecdsa.PrivateKey, error) {
	x := &big.Int{}
	x.SetString(j.X, 10)
	y := &big.Int{}
	y.SetString(j.Y, 10)
	d := &big.Int{}
	d.SetString(j.D, 10)

	crv, err := func() (elliptic.Curve, error) {
		switch j.Crv {
		case "P-256": // ECDSA with P-256 curve
			return elliptic.P256(), nil
		case "P-384": // ECDSA with P-384 curve
			return elliptic.P384(), nil
		case "P-521": // ECDSA with P-521 curve
			return elliptic.P521(), nil
		default:
			return nil, fmt.Errorf("unsupported curve: %s", j.Crv)
		}
	}()
	if err != nil {
		return nil, err
	}

	return &ecdsa.PrivateKey{
		D: d,
		PublicKey: ecdsa.PublicKey{
			Curve: crv,
			X:     x,
			Y:     y,
		},
	}, nil
}

type Config struct {
	Type string
	Jwk  []JWKEC
	Keys []struct {
		Name string
		Priv string
		Pub  string
	}
}

type MyCustomClaims struct {
	Foo string `json:"foo"`
	jwt.RegisteredClaims
}

type Token[Claims jwt.Claims] struct {
	Token  string
	Claims Claims
}

func fds[Claims jwt.Claims](t string) (*Token[Claims], error) {
	a, b := jwt.ParseWithClaims(t, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return nil, nil
	})
	return &Token[Claims]{Token: t, Claims: a.Claims.(Claims)}, b
}

func main() {

	x := &big.Int{}
	x.SetString("6864321535891942193567929857733336713205648373648304297643801782122250913610318779430781884774843426720090896834371917214112288991018202553005293705434583864", 10)

	yy := &ecdsa.PrivateKey{
		D: nil,
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P521(),
			X:     x,
			Y:     nil,
		},
	}
	fmt.Println(yy)
	p, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		panic(err)
	}
	fmt.Println(p)
	d, err := x509.MarshalPKIXPublicKey(&p.PublicKey)
	if err != nil {
		panic(err)
	}
	pem.Encode(os.Stdout, &pem.Block{Type: "PUBLIC KEY", Bytes: d})

	y, err := x509.MarshalECPrivateKey(p)
	if err != nil {
		panic(err)
	}
	pem.Encode(os.Stdout, &pem.Block{Type: "PRIVATE KEY", Bytes: y})

	var value Config
	if err := yaml.Unmarshal([]byte(valid_yaml), &value); err != nil {
		panic(err)
	}
	fmt.Println(value.Keys[0].Priv)
	fmt.Println(value.Jwk[0].GetKey())

	privateKey, err := jwt.ParseECPrivateKeyFromPEM([]byte(value.Keys[0].Priv))
	if err != nil {
		panic(err)
	}
	fmt.Println(privateKey)

	// Create the Claims
	claims := MyCustomClaims{
		"bar",
		jwt.RegisteredClaims{
			//ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer: "test",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES512, claims)
	ss, err := token.SignedString(privateKey)
	fmt.Printf("%v %v\n", ss, err)

	fmt.Println(privateKey.Params().BitSize)
	fmt.Println(privateKey.Curve.Params().Name)

	token, err = jwt.ParseWithClaims(ss, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(token.Valid)
	tt := token.Claims.(*MyCustomClaims)
	fmt.Println(tt.Foo)

	d, err = x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(err)
	}
	pem.Encode(os.Stdout, &pem.Block{Type: "PUBLIC KEY", Bytes: d})
}
