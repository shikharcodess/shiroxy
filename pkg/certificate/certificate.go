package certificate

// Copyright 2020 Matthew Holt
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"strings"

	"github.com/mholt/acmez/acme"
	"go.uber.org/zap"
)

func GenerateCertificate(domain string, email string) error {
	// put your domains here (IDNs must be in ASCII form)
	domains := []string{domain}

	// a context allows us to cancel long-running ops
	ctx := context.Background()

	// Logging is important - replace with your own zap logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}

	// first you need a private key for your certificate
	certPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generating certificate key: %v", err)
	}

	// then you need a certificate request; here's a simple one - we need
	// to fill out the template, then create the actual CSR, then parse it
	csrTemplate := &x509.CertificateRequest{DNSNames: domains}
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, csrTemplate, certPrivateKey)
	if err != nil {
		return fmt.Errorf("generating CSR: %v", err)
	}
	csr, err := x509.ParseCertificateRequest(csrDER)
	if err != nil {
		return fmt.Errorf("parsing generated CSR: %v", err)
	}

	// before you can get a cert, you'll need an account registered with
	// the ACME CA - it also needs a private key and should obviously be
	// different from any key used for certificates!
	accountPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generating account key: %v", err)
	}

	fmt.Println(strings.TrimSpace(email))
	var emails []string = []string{}

	account := acme.Account{
		Contact:              emails,
		TermsOfServiceAgreed: true,
		PrivateKey:           accountPrivateKey,
	}

	contactEmails := account.Contact
	fmt.Println("contactEmails: ", contactEmails)

	// now we can make our low-level ACME client
	client := &acme.Client{
		Directory: "https://127.0.0.1:14000/dir",
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
		Logger: logger,
	}

	// if the account is new, we need to create it; only do this once!
	// then be sure to securely store the account key and metadata so
	// you can reuse it later!
	account, err = client.NewAccount(ctx, account)
	if err != nil {
		return fmt.Errorf("new account: %v", err)
	}

	// now we can actually get a cert; first step is to create a new order
	var ids []acme.Identifier
	for _, domain := range domains {
		ids = append(ids, acme.Identifier{Type: "dns", Value: domain})
	}

	fmt.Println("2")
	order := acme.Order{Identifiers: ids}
	order, err = client.NewOrder(ctx, account, order)
	if err != nil {
		return fmt.Errorf("creating new order: %v", err)
	}

	fmt.Println("3")

	// each identifier on the order should now be associated with an
	// authorization object; we must make the authorization "valid"
	// by solving any of the challenges offered for it
	for _, authzURL := range order.Authorizations {
		authz, err := client.GetAuthorization(ctx, account, authzURL)
		if err != nil {
			return fmt.Errorf("getting authorization %q: %v", authzURL, err)
		}

		fmt.Println("=========================== authzURL: ", authzURL)

		challenge := authz.Challenges[0]

		challenge, err = client.InitiateChallenge(ctx, account, challenge)
		if err != nil {
			return fmt.Errorf("initiating challenge %q: %v", challenge.URL, err)
		}
		authz, err = client.PollAuthorization(ctx, account, authz)
		if err != nil {
			return fmt.Errorf("solving challenge: %v", err)
		}
	}

	order, err = client.FinalizeOrder(ctx, account, order, csr.Raw)
	if err != nil {
		return fmt.Errorf("finalizing order: %v", err)
	}

	certChains, err := client.GetCertificateChain(ctx, account, order.Certificate)
	if err != nil {
		return fmt.Errorf("downloading certs: %v", err)
	}

	for _, cert := range certChains {
		fmt.Printf("Certificate %q:\n%s\n\n", cert.URL, cert.ChainPEM)
	}

	return nil
}
