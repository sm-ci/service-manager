/*
 *    Copyright 2018 The Service Manager Authors
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package main

import (
	"context"
	"github.com/Peripli/service-manager/pkg/env"

	"github.com/Peripli/service-manager/config"

	"github.com/Peripli/service-manager/pkg/sm"
	"github.com/Peripli/service-manager/version"
)

func main() {
	version.Log()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	env, err := env.Default(ctx, config.AddPFlags)
	if err != nil {
		panic(err)
	}

	serviceManager, err := sm.New(ctx, cancel, env)
	if err != nil {
		panic(err)
	}

	serviceManager.Build().Run()
}
