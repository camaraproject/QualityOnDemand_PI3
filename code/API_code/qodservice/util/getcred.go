// Copyright 2023 Spry Fox Networks
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

package util

import "fmt"

type MTlsCred struct {
	myCert []byte
	myKey  []byte
	rootCA []byte
}

const (
	ROOTCA_OBJ_NAME = "/rootCA.pem"
	CERT_OBJ_NAME   = "/cert.pem"
	KEY_OBJ_NAME    = "/key.pem"
	ENV_LOCAL       = "local"
	ENV_AZURE       = "azure"
)

type CredentialReader interface {
	ReadMyCert() error
	ReadMyKey() error
	ReadRootCA() error
	ReturnAll() *MTlsCred
}

func (m *MTlsCred) GetMyCert() []byte {
	return m.myCert
}

func (m *MTlsCred) GetMyKey() []byte {
	return m.myKey
}

func (m *MTlsCred) GetRootCA() []byte {
	return m.rootCA
}

func getCertAndKey(c CredentialReader) error {
	err := c.ReadMyCert()
	if err != nil {
		return err
	}
	return c.ReadMyKey()

}

func getCredentials(c CredentialReader) (m *MTlsCred, err error) {
	err = getCertAndKey(c)
	if err != nil {
		return nil, err
	}
	err = c.ReadRootCA()
	if err != nil {
		return nil, err
	}
	return c.ReturnAll(), nil
}

func GetTlsCredentials(env, domainName, rootCaPath string) (m *MTlsCred, err error) {
	if env == ENV_LOCAL {
		f := credFromFile{
			myCertPath: domainName + CERT_OBJ_NAME,
			myKeyPath:  domainName + KEY_OBJ_NAME,
			rootCaPath: rootCaPath + ROOTCA_OBJ_NAME,
		}
		return getCredentials(&f)
	} else {
		return nil, fmt.Errorf("unknown env %v", env)
	}
}

func GetTlsCredentialsWithoutRootCA(env, domainName string) (m *MTlsCred, err error) {
	if env == ENV_LOCAL {
		f := credFromFile{
			myCertPath: domainName + CERT_OBJ_NAME,
			myKeyPath:  domainName + KEY_OBJ_NAME,
		}
		getCertAndKey(&f)
		return f.ReturnAll(), nil
	} else {
		return nil, fmt.Errorf("unknown env %v", env)
	}
}
