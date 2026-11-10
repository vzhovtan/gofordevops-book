package v1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CrawlerJobSpec struct {
    Schedule string `json:"schedule"`
    NetworkRanges []string `json:"networkRanges"`
    ConcurrentDevices int32 `json:"concurrentDevices"`
    DatabaseURL string `json:"databaseURL"`
    SSHKeySecret string `json:"sshKeySecret"`
    Image string `json:"image"`
}

type CrawlerJobStatus struct {
    LastRunTime *metav1.Time `json:"lastRunTime,omitempty"`
    LastRunStatus string `json:"lastRunStatus,omitempty"`
    DevicesDiscovered int32 `json:"devicesDiscovered,omitempty"`
    NextScheduledRun *metav1.Time `json:"nextScheduledRun,omitempty"`
}

type CrawlerJob struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec CrawlerJobSpec `json:"spec"`
    Status CrawlerJobStatus `json:"status,omitempty"`
}

type CrawlerJobList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items []CrawlerJob `json:"items"`
}