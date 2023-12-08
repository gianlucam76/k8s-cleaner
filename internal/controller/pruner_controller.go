/*
Copyright 2023.

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

package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "gianlucam76/k8s-pruner/api/v1alpha1"
	"gianlucam76/k8s-pruner/internal/controller/executor"
	"gianlucam76/k8s-pruner/pkg/scope"

	"github.com/go-logr/logr"
)

// PrunerReconciler reconciles a Pruner object
type PrunerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.projectsveltos.io,resources=pruners,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.projectsveltos.io,resources=pruners/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.projectsveltos.io,resources=pruners/finalizers,verbs=update
//+kubebuilder:rbac:groups="*",resources="*",verbs="*"

func (r *PrunerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	logger := ctrl.LoggerFrom(ctx)
	logger.Info("Reconciling")

	// Fecth the Pruner instance
	pruner := &appsv1alpha1.Pruner{}
	err := r.Get(ctx, req.NamespacedName, pruner)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		logger.Error(err, "Failed to fetch Pruner")
		return reconcile.Result{}, errors.Wrapf(
			err,
			"Failed to fetch Pruner %s",
			req.NamespacedName,
		)
	}

	prunerScope, err := scope.NewPrunerScope(scope.PrunerScopeParams{
		Client:         r.Client,
		Logger:         logger,
		Pruner:         pruner,
		ControllerName: "pruner",
	})
	if err != nil {
		logger.Error(err, "Failed to create profileScope")
		return reconcile.Result{}, errors.Wrapf(
			err,
			"unable to create profileScope scope for %s",
			req.NamespacedName,
		)
	}

	// Always close the scope when exiting this function so we can persist any
	// Pruner changes.
	defer func() {
		if err := prunerScope.Close(ctx); err != nil {
			reterr = err
		}
	}()

	logger = logger.WithValues("pruner", pruner.Name)

	if !pruner.DeletionTimestamp.IsZero() {
		r.reconcileDelete(prunerScope, logger)
		return ctrl.Result{}, nil
	}

	return r.reconcileNormal(ctx, prunerScope, logger)
}

func (r *PrunerReconciler) reconcileDelete(prunerScope *scope.PrunerScope, logger logr.Logger) {
	logger.Info("reconcileDelete")

	removeQueuedJobs(prunerScope)

	if controllerutil.ContainsFinalizer(prunerScope.Pruner, appsv1alpha1.PrunerFinalizer) {
		controllerutil.RemoveFinalizer(prunerScope.Pruner, appsv1alpha1.PrunerFinalizer)
	}

	logger.Info("reconcileDelete succeeded")
}

func (r *PrunerReconciler) reconcileNormal(ctx context.Context, prunerScope *scope.PrunerScope,
	logger logr.Logger) (reconcile.Result, error) {

	logger.Info("reconcileSnapshotNormal")
	if err := r.addFinalizer(ctx, prunerScope.Pruner, appsv1alpha1.PrunerFinalizer); err != nil {
		logger.Info(fmt.Sprintf("failed to add finalizer: %s", err))
		return reconcile.Result{}, err
	}

	executorClient := executor.GetClient()
	result := executorClient.GetResult(prunerScope.Pruner.Name)
	if result.ResultStatus != executor.Unavailable {
		if result.Err != nil {
			msg := result.Err.Error()
			prunerScope.SetFailureMessage(&msg)
		} else {
			prunerScope.SetFailureMessage(nil)
		}
	}

	now := time.Now()
	nextRun, err := schedule(ctx, prunerScope, logger)
	if err != nil {
		logger.Info("failed to get next run. Err: %v", err)
		msg := err.Error()
		prunerScope.SetFailureMessage(&msg)
		return ctrl.Result{}, err
	}

	logger.Info("reconcile snapshot succeeded")
	scheduledResult := ctrl.Result{RequeueAfter: nextRun.Sub(now)}
	return scheduledResult, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PrunerReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager,
	numOfWorker int, logger logr.Logger) error {

	executor.InitializeClient(ctx, logger, mgr.GetConfig(), mgr.GetClient(), numOfWorker)

	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.Pruner{}).
		Complete(r)
}

func (r *PrunerReconciler) addFinalizer(ctx context.Context, pruner *appsv1alpha1.Pruner, finalizer string) error {
	if controllerutil.ContainsFinalizer(pruner, finalizer) {
		return nil
	}

	controllerutil.AddFinalizer(pruner, finalizer)

	if err := r.Client.Update(ctx, pruner); err != nil {
		return err
	}
	return r.Get(ctx, types.NamespacedName{Name: pruner.Name}, pruner)
}

func schedule(ctx context.Context, prunerScope *scope.PrunerScope, logger logr.Logger) (*time.Time, error) {
	newLastRunTime := prunerScope.Pruner.Status.LastRunTime

	now := time.Now()
	nextRun, err := getNextScheduleTime(prunerScope.Pruner, now)
	if err != nil {
		logger.Info("failed to get next run. Err: %v", err)
		return nil, err
	}

	var newNextScheduleTime *metav1.Time
	if prunerScope.Pruner.Status.NextScheduleTime == nil {
		logger.Info("set NextScheduleTime")
		newNextScheduleTime = &metav1.Time{Time: *nextRun}
	} else {
		if shouldSchedule(prunerScope.Pruner, logger) {
			logger.Info("queuing job")
			executorClient := executor.GetClient()
			executorClient.Process(ctx, prunerScope.Pruner.Name)
			newLastRunTime = &metav1.Time{Time: now}
		}

		newNextScheduleTime = &metav1.Time{Time: *nextRun}
	}

	prunerScope.SetLastRunTime(newLastRunTime)
	prunerScope.SetNextScheduleTime(newNextScheduleTime)

	return nextRun, nil
}

// getNextScheduleTime gets the time of next schedule after last scheduled and before now
func getNextScheduleTime(pruner *appsv1alpha1.Pruner, now time.Time) (*time.Time, error) {
	sched, err := cron.ParseStandard(pruner.Spec.Schedule)
	if err != nil {
		return nil, fmt.Errorf("unparseable schedule %q: %w", pruner.Spec.Schedule, err)
	}

	var earliestTime time.Time
	if pruner.Status.LastRunTime != nil {
		earliestTime = pruner.Status.LastRunTime.Time
	} else {
		// If none found, then this is a recently created snapshot
		earliestTime = pruner.CreationTimestamp.Time
	}
	if pruner.Spec.StartingDeadlineSeconds != nil {
		// controller is not going to schedule anything below this point
		schedulingDeadline := now.Add(-time.Second * time.Duration(*pruner.Spec.StartingDeadlineSeconds))

		if schedulingDeadline.After(earliestTime) {
			earliestTime = schedulingDeadline
		}
	}

	starts := 0
	for t := sched.Next(earliestTime); t.Before(now); t = sched.Next(t) {
		const maxNumberOfFailures = 100
		starts++
		if starts > maxNumberOfFailures {
			return nil,
				fmt.Errorf("too many missed start times (> %d). Set or decrease .spec.startingDeadlineSeconds or check clock skew",
					maxNumberOfFailures)
		}
	}

	next := sched.Next(now)
	return &next, nil
}

func shouldSchedule(pruner *appsv1alpha1.Pruner, logger logr.Logger) bool {
	now := time.Now()
	logger.Info(fmt.Sprintf("currently next schedule is %s", pruner.Status.NextScheduleTime.Time))

	if now.Before(pruner.Status.NextScheduleTime.Time) {
		logger.Info("do not schedule yet")
		return false
	}

	// if last processed request was within 30 seconds, ignore it.
	// Avoid reprocessing spuriors back-to-back reconciliations
	if pruner.Status.LastRunTime != nil {
		logger.Info(fmt.Sprintf("last run was requested at %s", pruner.Status.LastRunTime))
		const ignoreTimeInSecond = 30
		diff := now.Sub(pruner.Status.LastRunTime.Time)
		logger.Info(fmt.Sprintf("Elapsed time since last run in minutes %f",
			diff.Minutes()))
		return diff.Seconds() >= ignoreTimeInSecond
	}

	return true
}

func removeQueuedJobs(prunerScope *scope.PrunerScope) {
	executorClient := executor.GetClient()
	executorClient.RemoveEntries(prunerScope.Pruner.Name)
}
