package main

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	prometheusCommon "github.com/webdevops/go-common/prometheus"

	devopsClient "github.com/webdevops/azure-devops-exporter/azure-devops-client"
)

type MetricsCollectorLatestBuild struct {
	CollectorProcessorProject

	prometheus struct {
		build       *prometheus.GaugeVec
		buildStatus *prometheus.GaugeVec
	}
}

func (m *MetricsCollectorLatestBuild) Setup(collector *CollectorProject) {
	m.CollectorReference = collector

	m.prometheus.build = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_build_latest_info",
			Help: "Azure DevOps build (latest)",
		},
		[]string{
			"projectID",
			"projectName",
			"buildDefinitionID",
			"buildID",
			"agentPoolID",
			"requestedBy",
			"buildNumber",
			"buildName",
			"sourceBranch",
			"sourceVersion",
			"status",
			"reason",
			"result",
			"url",
		},
	)
	prometheus.MustRegister(m.prometheus.build)

	m.prometheus.buildStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azure_devops_build_latest_status",
			Help: "Azure DevOps build (latest)",
		},
		[]string{
			"projectID",
			"projectName",
			"buildID",
			"buildNumber",
			"type",
		},
	)
	prometheus.MustRegister(m.prometheus.buildStatus)
}

func (m *MetricsCollectorLatestBuild) Reset() {
	m.prometheus.build.Reset()
	m.prometheus.buildStatus.Reset()
}

func (m *MetricsCollectorLatestBuild) Collect(ctx context.Context, logger *log.Entry, callback chan<- func(), project devopsClient.Project) {
	list, err := AzureDevopsClient.ListLatestBuilds(project.Id)
	if err != nil {
		logger.Error(err)
		return
	}

	buildMetric := prometheusCommon.NewMetricsList()
	buildStatusMetric := prometheusCommon.NewMetricsList()

	for _, build := range list.List {
		buildMetric.AddInfo(prometheus.Labels{
			"projectID":         project.Id,
			"projectName":       project.Name,
			"buildDefinitionID": int64ToString(build.Definition.Id),
			"buildID":           int64ToString(build.Id),
			"buildNumber":       build.BuildNumber,
			"buildName":         build.Definition.Name,
			"agentPoolID":       int64ToString(build.Queue.Pool.Id),
			"requestedBy":       build.RequestedBy.DisplayName,
			"sourceBranch":      build.SourceBranch,
			"sourceVersion":     build.SourceVersion,
			"status":            build.Status,
			"reason":            build.Reason,
			"result":            build.Result,
			"url":               build.Links.Web.Href,
		})

		buildStatusMetric.AddTime(prometheus.Labels{
			"projectID":   project.Id,
			"projectName": project.Name,
			"buildID":     int64ToString(build.Id),
			"buildNumber": build.BuildNumber,
			"type":        "started",
		}, build.StartTime)

		buildStatusMetric.AddTime(prometheus.Labels{
			"projectID":   project.Id,
			"projectName": project.Name,
			"buildID":     int64ToString(build.Id),
			"buildNumber": build.BuildNumber,
			"type":        "queued",
		}, build.QueueTime)

		buildStatusMetric.AddTime(prometheus.Labels{
			"projectID":   project.Id,
			"projectName": project.Name,
			"buildID":     int64ToString(build.Id),
			"buildNumber": build.BuildNumber,
			"type":        "finished",
		}, build.FinishTime)

		buildStatusMetric.AddDuration(prometheus.Labels{
			"projectID":   project.Id,
			"projectName": project.Name,
			"buildID":     int64ToString(build.Id),
			"buildNumber": build.BuildNumber,
			"type":        "jobDuration",
		}, build.FinishTime.Sub(build.StartTime))
	}

	callback <- func() {
		buildMetric.GaugeSet(m.prometheus.build)
		buildStatusMetric.GaugeSet(m.prometheus.buildStatus)
	}
}
