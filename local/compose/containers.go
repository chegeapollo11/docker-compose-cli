/*
   Copyright 2020 Docker Compose CLI authors

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

package compose

import (
	"context"
	"fmt"
	"sort"

	moby "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	"github.com/docker/compose-cli/utils"
)

// Containers is a set of moby Container
type Containers []moby.Container

type oneOff int

const (
	oneOffInclude = oneOff(iota)
	oneOffExclude
	oneOffOnly
)

func (s *composeService) getContainers(ctx context.Context, project string, oneOff oneOff, stopped bool, selectedServices ...string) (Containers, error) {
	var containers Containers
	f := filters.NewArgs(
		projectFilter(project),
	)
	switch oneOff {
	case oneOffOnly:
		f.Add("label", fmt.Sprintf("%s=%s", oneoffLabel, "True"))
	case oneOffExclude:
		f.Add("label", fmt.Sprintf("%s=%s", oneoffLabel, "False"))
	case oneOffInclude:
	}
	containers, err := s.apiClient.ContainerList(ctx, moby.ContainerListOptions{
		Filters: f,
		All:     stopped,
	})
	if err != nil {
		return nil, err
	}
	if len(selectedServices) > 0 {
		containers = containers.filter(isService(selectedServices...))
	}
	return containers, nil
}

// containerPredicate define a predicate we want container to satisfy for filtering operations
type containerPredicate func(c moby.Container) bool

func isService(services ...string) containerPredicate {
	return func(c moby.Container) bool {
		service := c.Labels[serviceLabel]
		return utils.StringContains(services, service)
	}
}

func isNotService(services ...string) containerPredicate {
	return func(c moby.Container) bool {
		service := c.Labels[serviceLabel]
		return !utils.StringContains(services, service)
	}
}

// filter return Containers with elements to match predicate
func (containers Containers) filter(predicate containerPredicate) Containers {
	var filtered Containers
	for _, c := range containers {
		if predicate(c) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func (containers Containers) names() []string {
	var names []string
	for _, c := range containers {
		names = append(names, getCanonicalContainerName(c))
	}
	return names
}

func (containers Containers) forEach(fn func(moby.Container)) {
	for _, c := range containers {
		fn(c)
	}
}

func (containers Containers) sorted() Containers {
	sort.Slice(containers, func(i, j int) bool {
		return getCanonicalContainerName(containers[i]) < getCanonicalContainerName(containers[j])
	})
	return containers
}
