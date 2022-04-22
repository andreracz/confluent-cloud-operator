package controllers

import (
	"errors"
	"fmt"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

func buildResult(requeue bool, requeueAfter time.Time) ctrl.Result {
	return ctrl.Result{
		Requeue:      requeue,
		RequeueAfter: time.Since(requeueAfter),
	}
}

func keyMissingErr(key string) error {
	return errors.New(fmt.Sprintf("failed to retrieve %s: key %s missing from credentials", key, key))
}
