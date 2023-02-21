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

import (
	"io/ioutil"
)

type credFromFile struct {
	myCertPath string
	myKeyPath  string
	rootCaPath string
	mTls       MTlsCred
}

func (c *credFromFile) getBlob(url string) ([]byte, error) {
	return ioutil.ReadFile(url)
}

func (c *credFromFile) ReadMyCert() (err error) {
	c.mTls.myCert, err = c.getBlob(c.myCertPath)
	return err
}

func (c *credFromFile) ReadMyKey() (err error) {
	c.mTls.myKey, err = c.getBlob(c.myKeyPath)
	return err
}

func (c *credFromFile) ReadRootCA() (err error) {
	c.mTls.rootCA, err = c.getBlob(c.rootCaPath)
	return err
}

func (c *credFromFile) ReturnAll() *MTlsCred {
	return &c.mTls
}
