// Copyright 2015 Google Inc. All Rights Reserved.
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

package influxdb

import (
	"testing"
	"time"

	"net/http/httptest"
	"net/url"

	influxdb_common "github.com/sq325/kube-eventer/common/influxdb"
	"github.com/sq325/kube-eventer/core"
	"github.com/stretchr/testify/assert"
	kube_api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	util "k8s.io/client-go/util/testing"
)

type fakeInfluxDBEventSink struct {
	core.EventSink
	fakeDbClient *influxdb_common.FakeInfluxDBClient
}

func NewFakeSink() fakeInfluxDBEventSink {
	return fakeInfluxDBEventSink{
		&influxdbSink{
			client: influxdb_common.Client,
			c:      influxdb_common.Config,
		},
		influxdb_common.Client,
	}
}

func TestStoreDataEmptyInput(t *testing.T) {
	fakeSink := NewFakeSink()
	eventBatch := core.EventBatch{}
	fakeSink.ExportEvents(&eventBatch)
	assert.Equal(t, 0, len(fakeSink.fakeDbClient.Pnts))
}

func TestStoreMultipleDataInput(t *testing.T) {
	fakeSink := NewFakeSink()
	timestamp := time.Now()

	now := time.Now()
	event1 := kube_api.Event{
		Message:        "event1",
		Count:          100,
		LastTimestamp:  metav1.NewTime(now),
		FirstTimestamp: metav1.NewTime(now),
	}

	event2 := kube_api.Event{
		Message:        "event2",
		Count:          101,
		LastTimestamp:  metav1.NewTime(now),
		FirstTimestamp: metav1.NewTime(now),
	}

	data := core.EventBatch{
		Timestamp: timestamp,
		Events: []*kube_api.Event{
			&event1,
			&event2,
		},
	}

	fakeSink.ExportEvents(&data)
	assert.Equal(t, 2, len(fakeSink.fakeDbClient.Pnts))
}

func TestCreateInfluxdbSink(t *testing.T) {
	handler := util.FakeHandler{
		StatusCode:   200,
		RequestBody:  "",
		ResponseBody: "",
		T:            t,
	}
	server := httptest.NewServer(&handler)
	defer server.Close()

	stubInfluxDBUrl, err := url.Parse(server.URL)
	assert.NoError(t, err)

	//create influxdb sink
	sink, err := CreateInfluxdbSink(stubInfluxDBUrl)
	assert.NoError(t, err)

	//check sink name
	assert.Equal(t, sink.Name(), "InfluxDB Sink")
}
