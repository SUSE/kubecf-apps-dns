/*
Copyright 2020 SUSE

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package svcdiscovery

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SDCClient is the Service Discovery Controller Client used to make calls to
// the service-discovery-controller job to discover internal routes to apps.
type SDCClient struct {
	httpClient *http.Client
	sdcURLBase string
}

// Discover discovers internal app routes from the Service Discovery Controller
// and returns the list of IPs from these discovered routes.
func (sdcc *SDCClient) Discover(ctx context.Context, domainName string) ([]string, error) {
	url := sdcc.sdcURLBase + domainName
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to discover service: %w", err)
	}
	res, err := sdcc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to discover service: %w", err)
	}
	defer res.Body.Close()
	decoder := json.NewDecoder(res.Body)
	var sdcClientResponse SDCClientResponse
	if err := decoder.Decode(&sdcClientResponse); err != nil {
		return nil, fmt.Errorf("failed to discover service: %w", err)
	}
	ips := make([]string, len(sdcClientResponse.Hosts))
	for i, host := range sdcClientResponse.Hosts {
		ips[i] = host.IPAddress
	}
	return ips, nil
}

// SDCClientResponse represents a response from the Service Discovery
// Controller. An example response in JSON:
//
// {
//   "hosts": [
//     {
//       "ip_address": "10.255.141.235",
//       "last_check_in": "",
//       "port": 0,
//       "revision": "",
//       "service": "",
//       "service_repo_name": "",
//       "tags": {}
//     }
//   ],
//   "env": "",
//   "service": ""
// }
//
// It's important to notice that we only need the `ip_address` of each host in
// the response to be able to construct the DNS answer, hence the reason why we
// ignore the other JSON fields.
type SDCClientResponse struct {
	Hosts []struct {
		IPAddress string `json:"ip_address"`
	} `json:"hosts"`
}
