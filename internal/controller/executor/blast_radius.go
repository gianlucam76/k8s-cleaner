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

package executor

import (
	"fmt"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
)

// checkBlastRadiusLimit returns a non-nil error if matchedCount exceeds the limits
// configured in limit. A nil limit never trips. totalScanned is the number of
// resources considered by the ResourceSelectors before filtering, used as the
// denominator for MaxPercentage.
func checkBlastRadiusLimit(limit *appsv1alpha1.BlastRadiusLimit, matchedCount, totalScanned int) error {
	if limit == nil {
		return nil
	}

	if limit.MaxCount != nil && matchedCount > *limit.MaxCount {
		return fmt.Errorf("blast radius limit exceeded: %d resources matched (max count %d)",
			matchedCount, *limit.MaxCount)
	}

	if limit.MaxPercentage != nil && totalScanned > 0 {
		percentage := matchedCount * 100 / totalScanned
		if percentage > *limit.MaxPercentage {
			return fmt.Errorf("blast radius limit exceeded: %d%% of resources matched (max percentage %d%%)",
				percentage, *limit.MaxPercentage)
		}
	}

	return nil
}
