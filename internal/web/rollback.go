/*
Copyright 2026. projectsveltos.io. All rights reserved.

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

package web

import (
	"net/http"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"gianlucam76/k8s-cleaner/internal/controller/executor"
)

// RollbackHandler reverts the most recent execution of a Cleaner, using the
// pre-action resource state captured on its Report instance.
func RollbackHandler(c client.Client, log logr.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		name := r.PathValue("name")

		results, err := executor.Rollback(ctx, c, name, log)
		if err != nil {
			if apierrors.IsNotFound(err) {
				respondError(w, http.StatusNotFound, "report not found")
				return
			}
			log.Error(err, "failed to roll back", "name", name)
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, results)
	}
}
