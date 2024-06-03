package remotewrite

import (
	"log"
	"net/url"

	rwClient "github.com/sq325/kube-eventer/common/remotewrite"
	"github.com/sq325/kube-eventer/common/remotewrite/prompb"
	"github.com/sq325/kube-eventer/core"
	v1 "k8s.io/api/core/v1"
)

// remotewriteSink is a sink that writes events to a remote write endpoint.
// It implements the Sink interface.
type remotewriteSink struct {
	client  rwClient.RemoteWriteClient
	factory MetricFactory
	cluster string
}

func NewSink(uri *url.URL) (core.EventSink, error) {
	remotewriteUrl := uri.Scheme + "://" + uri.Host + uri.Path
	cluster := uri.Query().Get("cluster")
	log.Println("remote write url: ", remotewriteUrl)
	log.Println("cluster name: ", cluster)
	return &remotewriteSink{
		client:  rwClient.NewClient(remotewriteUrl),
		factory: NewMetricFactory(),
		cluster: cluster,
	}, nil
}

func (sink *remotewriteSink) Name() string {
	return "RemoteWrite"
}

func (sink *remotewriteSink) ExportEvents(batch *core.EventBatch) {
	if len(batch.Events) == 0 {
		return
	}
	sink.write(batch.Events)
}

func (sink *remotewriteSink) Stop() {
}

func (sink *remotewriteSink) write(events []*v1.Event) (err error) {
	var seriesList []*prompb.TimeSeries
	for _, event := range events {
		seriesList = append(seriesList, sink.factory.EventToMetric(event, sink.cluster))
	}
	return sink.client.Write(seriesList)
}
