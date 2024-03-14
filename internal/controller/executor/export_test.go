/*
Copyright 2023. projectsveltos.io. All rights reserved.

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

var (
	FetchResources          = fetchResources
	GetMatchingResources    = getMatchingResources
	DeleteMatchingResources = deleteMatchingResources
	IsMatch                 = isMatch
	Transform               = transform
	AggregatedSelection     = aggregatedSelection
	GetNamespaces           = getNamespaces
)

var (
	GetWebexInfo = getWebexInfo
	GetSlackInfo = getSlackInfo
)

func (m *Manager) ClearInternalStruct() {
	m.dirty = make([]string, 0)
	m.inProgress = make([]string, 0)
	m.jobQueue = make([]string, 0)
	m.results = make(map[string]error)
}

func (m *Manager) SetInProgress(inProgress []string) {
	m.inProgress = inProgress
}

func (m *Manager) GetInProgress() []string {
	return m.inProgress
}

func (m *Manager) SetDirty(dirty []string) {
	m.dirty = dirty
}

func (m *Manager) GetDirty() []string {
	return m.dirty
}

func (m *Manager) SetJobQueue(cleanerName string) {
	m.jobQueue = []string{cleanerName}
}

func (m *Manager) GetJobQueue() []string {
	return m.jobQueue
}

func (m *Manager) SetResults(results map[string]error) {
	m.results = results
}

func (m *Manager) GetResults() map[string]error {
	return m.results
}

func GetWebexRoom(info *webexInfo) string {
	return info.room
}
func GetWebexToken(info *webexInfo) string {
	return info.token
}

func GetSlackChannelID(info *slackInfo) string {
	return info.channelID
}
func GetSlackToken(info *slackInfo) string {
	return info.token
}
