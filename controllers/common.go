package controllers

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

func buildResult(requeue bool, requeueAfter time.Time) ctrl.Result {
	return ctrl.Result{
		Requeue:      requeue,
		RequeueAfter: time.Since(requeueAfter),
	}
}
