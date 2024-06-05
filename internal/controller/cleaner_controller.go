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
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
	"gianlucam76/k8s-cleaner/internal/controller/executor"
	"gianlucam76/k8s-cleaner/pkg/scope"

	"github.com/go-logr/logr"
)

// CleanerReconciler reconciles a Cleaner object
type CleanerReconciler struct {
	client.Client
	Scheme               *runtime.Scheme
	ConcurrentReconciles int
}

//+kubebuilder:rbac:groups=apps.projectsveltos.io,resources=cleaners,verbs=get;list;watch
//+kubebuilder:rbac:groups=apps.projectsveltos.io,resources=cleaners/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.projectsveltos.io,resources=cleaners/finalizers,verbs=update
//+kubebuilder:rbac:groups="*",resources="*",verbs="*"

func (r *CleanerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	logger := ctrl.LoggerFrom(ctx)
	logger.Info("Reconciling")

	// Fecth the Cleaner instance
	cleaner := &appsv1alpha1.Cleaner{}
	err := r.Get(ctx, req.NamespacedName, cleaner)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		logger.Error(err, "Failed to fetch Cleaner")
		return reconcile.Result{}, errors.Wrapf(
			err,
			"Failed to fetch Cleaner %s",
			req.NamespacedName,
		)
	}

	cleanerScope, err := scope.NewCleanerScope(scope.CleanerScopeParams{
		Client:         r.Client,
		Logger:         logger,
		Cleaner:        cleaner,
		ControllerName: "cleaner",
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
	// Cleaner changes.
	defer func() {
		if err := cleanerScope.Close(ctx); err != nil {
			reterr = err
		}
	}()

	logger = logger.WithValues("cleaner", cleaner.Name)

	if !cleaner.DeletionTimestamp.IsZero() {
		return reconcile.Result{}, r.reconcileDelete(ctx, cleanerScope, logger)
	}

	return r.reconcileNormal(ctx, cleanerScope, logger)
}

func (r *CleanerReconciler) reconcileDelete(ctx context.Context,
	cleanerScope *scope.CleanerScope, logger logr.Logger) error {

	logger.Info("reconcileDelete")

	removeQueuedJobs(cleanerScope)

	err := r.removeReport(ctx, cleanerScope, logger)
	if err != nil {
		return err
	}

	if controllerutil.ContainsFinalizer(cleanerScope.Cleaner, appsv1alpha1.CleanerFinalizer) {
		controllerutil.RemoveFinalizer(cleanerScope.Cleaner, appsv1alpha1.CleanerFinalizer)
	}

	logger.Info("reconcileDelete succeeded")

	return nil
}

func (r *CleanerReconciler) reconcileNormal(ctx context.Context, cleanerScope *scope.CleanerScope,
	logger logr.Logger) (reconcile.Result, error) {

	logger.Info("reconcileSnapshotNormal")
	if err := r.addFinalizer(ctx, cleanerScope.Cleaner, appsv1alpha1.CleanerFinalizer); err != nil {
		logger.Info(fmt.Sprintf("failed to add finalizer: %s", err))
		return reconcile.Result{}, err
	}

	executorClient := executor.GetClient()
	result := executorClient.GetResult(cleanerScope.Cleaner.Name)
	if result.ResultStatus != executor.Unavailable {
		if result.Err != nil {
			msg := result.Err.Error()
			cleanerScope.SetFailureMessage(&msg)
		} else {
			cleanerScope.SetFailureMessage(nil)
		}
	}

	now := time.Now()
	nextRun, err := schedule(ctx, cleanerScope, logger)
	if err != nil {
		logger.Info("failed to get next run. Err: %v", err)
		msg := err.Error()
		cleanerScope.SetFailureMessage(&msg)
		return ctrl.Result{}, err
	}

	logger.Info("reconcile snapshot succeeded")
	scheduledResult := ctrl.Result{RequeueAfter: nextRun.Sub(now)}
	return scheduledResult, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CleanerReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager,
	numOfWorker int, logger logr.Logger) error {

	executor.InitializeClient(ctx, logger, mgr.GetConfig(), mgr.GetClient(), mgr.GetScheme(), numOfWorker)

	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.Cleaner{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.ConcurrentReconciles,
		}).
		Complete(r)
}

func (r *CleanerReconciler) addFinalizer(ctx context.Context, cleaner *appsv1alpha1.Cleaner, finalizer string) error {
	if controllerutil.ContainsFinalizer(cleaner, finalizer) {
		return nil
	}

	controllerutil.AddFinalizer(cleaner, finalizer)

	if err := r.Client.Update(ctx, cleaner); err != nil {
		return err
	}
	return r.Get(ctx, types.NamespacedName{Name: cleaner.Name}, cleaner)
}

// removeReport deletes (if present) Report generated for this Cleaner
// instance
func (r *CleanerReconciler) removeReport(ctx context.Context,
	cleanerScope *scope.CleanerScope, logger logr.Logger) error {

	report := &appsv1alpha1.Report{}
	err := r.Get(ctx, types.NamespacedName{Name: cleanerScope.Cleaner.GetName()}, report)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}

		logger.Info(fmt.Sprintf("failed to get Report: %v", err))
		return err
	}

	err = r.Delete(ctx, report)
	if err != nil {
		return err
	}

	return fmt.Errorf("report instance still present")
}

func schedule(ctx context.Context, cleanerScope *scope.CleanerScope, logger logr.Logger) (*time.Time, error) {
	newLastRunTime := cleanerScope.Cleaner.Status.LastRunTime

	now := time.Now()
	nextRun, err := getNextScheduleTime(cleanerScope.Cleaner, now)
	if err != nil {
		logger.Info("failed to get next run. Err: %v", err)
		return nil, err
	}

	var newNextScheduleTime *metav1.Time
	if cleanerScope.Cleaner.Status.NextScheduleTime == nil {
		logger.Info("set NextScheduleTime")
		newNextScheduleTime = &metav1.Time{Time: *nextRun}
	} else {
		if shouldSchedule(cleanerScope.Cleaner, logger) {
			logger.Info("queuing job")
			executorClient := executor.GetClient()
			executorClient.Process(ctx, cleanerScope.Cleaner.Name)
			newLastRunTime = &metav1.Time{Time: now}
		}

		newNextScheduleTime = &metav1.Time{Time: *nextRun}
	}

	cleanerScope.SetLastRunTime(newLastRunTime)
	cleanerScope.SetNextScheduleTime(newNextScheduleTime)

	return nextRun, nil
}

// getNextScheduleTime gets the time of next schedule after last scheduled and before now
func getNextScheduleTime(cleaner *appsv1alpha1.Cleaner, now time.Time) (*time.Time, error) {
	sched, err := cron.ParseStandard(cleaner.Spec.Schedule)
	if err != nil {
		return nil, fmt.Errorf("unparseable schedule %q: %w", cleaner.Spec.Schedule, err)
	}

	var earliestTime time.Time
	if cleaner.Status.LastRunTime != nil {
		earliestTime = cleaner.Status.LastRunTime.Time
	} else {
		// If none found, then this is a recently created snapshot
		earliestTime = cleaner.CreationTimestamp.Time
	}
	if cleaner.Spec.StartingDeadlineSeconds != nil {
		// controller is not going to schedule anything below this point
		schedulingDeadline := now.Add(-time.Second * time.Duration(*cleaner.Spec.StartingDeadlineSeconds))

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

func shouldSchedule(cleaner *appsv1alpha1.Cleaner, logger logr.Logger) bool {
	now := time.Now()
	logger.Info(fmt.Sprintf("currently next schedule is %s", cleaner.Status.NextScheduleTime.Time))

	if now.Before(cleaner.Status.NextScheduleTime.Time) {
		logger.Info("do not schedule yet")
		return false
	}

	// if last processed request was within 30 seconds, ignore it.
	// Avoid reprocessing spuriors back-to-back reconciliations
	if cleaner.Status.LastRunTime != nil {
		logger.Info(fmt.Sprintf("last run was requested at %s", cleaner.Status.LastRunTime))
		const ignoreTimeInSecond = 30
		diff := now.Sub(cleaner.Status.LastRunTime.Time)
		logger.Info(fmt.Sprintf("Elapsed time since last run in minutes %f",
			diff.Minutes()))
		return diff.Seconds() >= ignoreTimeInSecond
	}

	return true
}

func removeQueuedJobs(cleanerScope *scope.CleanerScope) {
	executorClient := executor.GetClient()
	executorClient.RemoveEntries(cleanerScope.Cleaner.Name)
}
