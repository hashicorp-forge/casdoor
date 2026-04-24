// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package object

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/beevik/etree"
)

func TestNewSamlSigningContextCanonicalizer(t *testing.T) {
	keyStore := newSamlSigningTestKeyStore(t)

	tests := []struct {
		name              string
		enableSamlC14n10  bool
		wantAlgorithm     string
		unwantedAlgorithm string
		wantPrefixListXs  bool
	}{
		{
			name:              "default branch keeps c14n11",
			enableSamlC14n10:  false,
			wantAlgorithm:     "http://www.w3.org/2006/12/xml-c14n11",
			unwantedAlgorithm: "http://www.w3.org/2001/10/xml-exc-c14n#",
			wantPrefixListXs:  false,
		},
		{
			name:              "c14n10 branch uses exclusive canonicalization without prefix list",
			enableSamlC14n10:  true,
			wantAlgorithm:     "http://www.w3.org/2001/10/xml-exc-c14n#",
			unwantedAlgorithm: "http://www.w3.org/2006/12/xml-c14n11",
			wantPrefixListXs:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newSamlSigningContext(&Application{EnableSamlC14n10: tt.enableSamlC14n10}, keyStore)
			signatureXML := buildSamlSignatureXML(t, ctx)

			if !strings.Contains(signatureXML, tt.wantAlgorithm) {
				t.Fatalf("signature = %q, want algorithm %q", signatureXML, tt.wantAlgorithm)
			}

			if tt.unwantedAlgorithm != "" && strings.Contains(signatureXML, tt.unwantedAlgorithm) {
				t.Fatalf("signature = %q, found unexpected algorithm %q", signatureXML, tt.unwantedAlgorithm)
			}

			hasPrefixListXs := strings.Contains(signatureXML, `PrefixList="xs"`) || strings.Contains(signatureXML, "InclusiveNamespaces")
			if hasPrefixListXs != tt.wantPrefixListXs {
				t.Fatalf("signature = %q, PrefixList xs presence = %v, want %v", signatureXML, hasPrefixListXs, tt.wantPrefixListXs)
			}
		})
	}
}

func buildSamlSignatureXML(t *testing.T, ctx interface {
	ConstructSignature(el *etree.Element, enveloped bool) (*etree.Element, error)
}) string {
	t.Helper()

	assertion := etree.NewElement("saml:Assertion")
	assertion.CreateAttr("xmlns:saml", "urn:oasis:names:tc:SAML:2.0:assertion")
	assertion.CreateAttr("xmlns:xs", "http://www.w3.org/2001/XMLSchema")
	assertion.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
	assertion.CreateAttr("ID", "_test-assertion")
	assertion.CreateElement("saml:Issuer").SetText("https://example.com")
	attributeValue := assertion.CreateElement("saml:AttributeStatement").CreateElement("saml:Attribute").CreateElement("saml:AttributeValue")
	attributeValue.CreateAttr("xsi:type", "xs:string")
	attributeValue.SetText("user@example.com")

	signature, err := ctx.ConstructSignature(assertion, true)
	if err != nil {
		t.Fatalf("ConstructSignature() error = %v", err)
	}

	doc := etree.NewDocument()
	doc.SetRoot(signature)

	bytes, err := doc.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes() error = %v", err)
	}

	return string(bytes)
}

func newSamlSigningTestKeyStore(t *testing.T) *X509Key {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "casdoor.test",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
	}

	certificateDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("CreateCertificate() error = %v", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	return &X509Key{
		PrivateKey:      string(privateKeyPEM),
		X509Certificate: base64.StdEncoding.EncodeToString(certificateDER),
	}
}
