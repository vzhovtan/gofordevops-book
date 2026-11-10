package controllers

import (
    "context"
    "fmt"
    "time"
    
    batchv1 "k8s.io/api/batch/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    
    infrastructurev1 "example.com/crawler-operator/api/v1"
)

type CrawlerJobReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

func (r *CrawlerJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    crawlerJob := &infrastructurev1.CrawlerJob{}
    if err := r.Get(ctx, req.NamespacedName, crawlerJob); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }
    
    nextRun := r.calculateNextRun(crawlerJob)
    now := time.Now()
    
    if crawlerJob.Status.LastRunTime == nil || now.After(nextRun) {
        job := r.constructJob(crawlerJob)
        if err := r.Create(ctx, job); err != nil {
            return ctrl.Result{}, err
        }
        
        crawlerJob.Status.LastRunTime = &metav1.Time{Time: now}
        crawlerJob.Status.LastRunStatus = "Running"
        if err := r.Status().Update(ctx, crawlerJob); err != nil {
            return ctrl.Result{}, err
        }
    }
    
    return ctrl.Result{RequeueAfter: time.Until(nextRun)}, nil
}

func (r *CrawlerJobReconciler) constructJob(crawlerJob *infrastructurev1.CrawlerJob) *batchv1.Job {
    job := &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name: fmt.Sprintf("%s-%d", crawlerJob.Name, time.Now().Unix()),
            Namespace: crawlerJob.Namespace,
        },
        Spec: batchv1.JobSpec{
            Template: corev1.PodTemplateSpec{
                Spec: corev1.PodSpec{
                    RestartPolicy: corev1.RestartPolicyNever,
                    Containers: []corev1.Container{
                        {
                            Name: "crawler",
                            Image: crawlerJob.Spec.Image,
                            Env: []corev1.EnvVar{
                                {Name: "CRAWLER_DATABASE_URL", Value: crawlerJob.Spec.DatabaseURL},
                                {Name: "CRAWLER_CONCURRENT_DEVICES", Value: fmt.Sprintf("%d", crawlerJob.Spec.ConcurrentDevices)},
                                {Name: "CRAWLER_NETWORK_RANGES", Value: fmt.Sprintf("%v", crawlerJob.Spec.NetworkRanges)},
                            },
                            VolumeMounts: []corev1.VolumeMount{
                                {Name: "ssh-keys", MountPath: "/home/crawler/.ssh", ReadOnly: true},
                            },
                        },
                    },
                    Volumes: []corev1.Volume{
                        {
                            Name: "ssh-keys",
                            VolumeSource: corev1.VolumeSource{
                                Secret: &corev1.SecretVolumeSource{SecretName: crawlerJob.Spec.SSHKeySecret},
                            },
                        },
                    },
                },
            },
        },
    }
    ctrl.SetControllerReference(crawlerJob, job, r.Scheme)
    return job
}

func (r *CrawlerJobReconciler) calculateNextRun(crawlerJob *infrastructurev1.CrawlerJob) time.Time {
    return time.Now().Add(1 * time.Hour)
}

func (r *CrawlerJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&infrastructurev1.CrawlerJob{}).
        Owns(&batchv1.Job{}).
        Complete(r)
}