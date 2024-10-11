// Copyright 2024 CloudWeGo Authors
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

package client

import (
	"github.com/kitex-contrib/config-consul/consul"
	"github.com/kitex-contrib/config-consul/utils"

	"github.com/cloudwego/kitex/client"

	configclient "github.com/cloudwego-contrib/cwgo-pkg/config/consul/client"
)

// WithCircuitBreaker sets the circuit breaker policy from consul configuration center.
func WithCircuitBreaker(dest, src string, consulClient consul.Client, uniqueID int64, opts utils.Options) []client.Option {
	return configclient.WithCircuitBreaker(dest, src, consulClient, uniqueID, opts)
}
