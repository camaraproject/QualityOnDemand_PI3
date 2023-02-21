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

package main

import (
	"fmt"
	"os"

	"github.com/sfnuser/qodservice/service"
	"github.com/urfave/cli/v2"
)

var QoD *service.QoD

func main() {
	app := cli.NewApp()
	QoD = service.NewQoD()

	app.Name = "qodservice"
	app.Usage = "-h for help"
	app.Action = action
	app.Flags = QoD.GetCliCmd()

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s Run Error: %v", app.Name, err)
	}
}

func action(c *cli.Context) error {
	if err := QoD.Initialize(c); err != nil {
		return fmt.Errorf("failed to initialize. err %v", err)
	}
	QoD.Start(c)
	return nil
}
