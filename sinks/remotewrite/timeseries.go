package remotewrite

import (
	"time"

	"github.com/sq325/kube-eventer/common/remotewrite/prompb"
	v1 "k8s.io/api/core/v1"
)

type MetricFactory interface {
	EventToMetric(event *v1.Event, cluster string) *prompb.TimeSeries
}

type factory struct {
	WarningM  string   // event_warning
	NormalM   string   // event_normal
	Labelkeys []string // [namespace, type, name, message, reason]
}

func NewMetricFactory() MetricFactory {
	return &factory{
		WarningM:  "event_warning_total",
		NormalM:   "event_normal_total",
		Labelkeys: []string{"namespace", "type", "name", "reason", "message", "cluster"},
	}
}

func (f *factory) EventToMetric(event *v1.Event, cluster string) *prompb.TimeSeries {
	var (
		labels    []*prompb.Label
		count     float64 = float64(event.Count)
		timestamp int64   = time.Now().UnixMilli()
	)
	if event.Type == "Warning" {
		labels = append(labels, &prompb.Label{Name: "__name__", Value: f.WarningM})
	} else {
		labels = append(labels, &prompb.Label{Name: "__name__", Value: f.NormalM})
	}

	for _, key := range f.Labelkeys {
		var value string
		switch key {
		case "namespace":
			value = event.InvolvedObject.Namespace
		case "type":
			value = event.Type
		case "name":
			value = event.InvolvedObject.Name
		case "reason":
			value = event.Reason
		case "message":
			value = event.Message
		case "cluster":
			value = cluster
		}
		labels = append(labels, &prompb.Label{Name: key, Value: value})
	}

	timeseries := &prompb.TimeSeries{
		Labels: labels,
		Samples: []*prompb.Sample{
			{
				Value:     count,
				Timestamp: timestamp,
			},
		},
	}

	return timeseries
}
