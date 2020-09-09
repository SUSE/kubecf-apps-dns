# Copyright 2020 SUSE
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ARG GO_VERSION=1.15
FROM golang:${GO_VERSION} AS staging
ARG COREDNS_VERSION=1.7.0
ENV COREDNS_VERSION=${COREDNS_VERSION}
WORKDIR /build
RUN curl -LO "https://github.com/coredns/coredns/archive/v${COREDNS_VERSION}.tar.gz"
RUN mkdir coredns
RUN tar zxf "v${COREDNS_VERSION}.tar.gz" --directory=coredns --strip-components=1
WORKDIR /build/coredns
COPY svcdiscovery/plugin/svcdiscovery/ ./plugin/svcdiscovery/
COPY svcdiscovery/plugin.cfg plugin.cfg
RUN go generate
RUN make
WORKDIR /build
COPY svcdiscovery/resolvwriter/ ./resolvwriter
WORKDIR /build/resolvwriter
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o resolvwriter main.go

FROM scratch
COPY --from=staging /build/coredns/coredns /coredns
COPY --from=staging /build/resolvwriter/resolvwriter /resolvwriter
ENTRYPOINT ["/coredns"]
