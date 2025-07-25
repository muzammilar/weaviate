//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2025 Weaviate B.V. All rights reserved.
//
//  CONTACT: hello@weaviate.io
//

package db

import (
	"context"
	"fmt"
	"math"

	"github.com/go-openapi/strfmt"
	"github.com/pkg/errors"
	"github.com/weaviate/weaviate/entities/additional"
	"github.com/weaviate/weaviate/entities/dto"
	"github.com/weaviate/weaviate/entities/filters"
	"github.com/weaviate/weaviate/entities/models"
	"github.com/weaviate/weaviate/entities/schema"
	"github.com/weaviate/weaviate/entities/search"
	"github.com/weaviate/weaviate/usecases/classification"
	"github.com/weaviate/weaviate/usecases/vectorizer"
)

// TODO: why is this logic in the persistence package? This is business-logic,
// move out of here!
func (db *DB) GetUnclassified(ctx context.Context, className string,
	properties []string, propsToReturn []string, filter *filters.LocalFilter,
) ([]search.Result, error) {
	propsToReturnTmp := append(properties, propsToReturn...)
	props := make(search.SelectProperties, len(propsToReturnTmp))
	for i, prop := range propsToReturnTmp {
		props[i] = search.SelectProperty{Name: prop}
	}
	mergedFilter := mergeUserFilterWithRefCountFilter(filter, className, properties,
		filters.OperatorEqual, 0)
	res, err := db.Search(ctx, dto.GetParams{
		ClassName: className,
		Filters:   mergedFilter,
		Pagination: &filters.Pagination{
			Limit: 10000, // TODO: gh-1219 increase
		},
		AdditionalProperties: additional.Properties{
			Classification: true,
			Vector:         true,
			ModuleParams: map[string]interface{}{
				"interpretation": true,
			},
		},
		Properties: props,
	})

	return res, err
}

// TODO: why is this logic in the persistence package? This is business-logic,
// move out of here!
func (db *DB) ZeroShotSearch(ctx context.Context, vector []float32,
	class string, properties []string,
	filter *filters.LocalFilter,
) ([]search.Result, error) {
	props := make(search.SelectProperties, len(properties))
	for i, prop := range properties {
		props[i] = search.SelectProperty{Name: prop}
	}

	res, err := db.VectorSearch(ctx, dto.GetParams{
		ClassName: class,
		Pagination: &filters.Pagination{
			Limit: 1,
		},
		Filters: filter,
		AdditionalProperties: additional.Properties{
			Vector: true,
		},
		Properties: props,
	}, []string{""}, []models.Vector{vector})

	return res, err
}

// TODO: why is this logic in the persistence package? This is business-logic,
// move out of here!
func (db *DB) AggregateNeighbors(ctx context.Context, vector []float32,
	class string, properties []string, k int,
	filter *filters.LocalFilter,
) ([]classification.NeighborRef, error) {
	props := make(search.SelectProperties, len(properties))
	for i, prop := range properties {
		props[i] = search.SelectProperty{Name: prop}
	}
	mergedFilter := mergeUserFilterWithRefCountFilter(filter, class, properties,
		filters.OperatorGreaterThan, 0)
	res, err := db.VectorSearch(ctx, dto.GetParams{
		ClassName: class,
		Pagination: &filters.Pagination{
			Limit: k,
		},
		Filters: mergedFilter,
		AdditionalProperties: additional.Properties{
			Vector: true,
		},
		Properties: props,
	}, []string{""}, []models.Vector{vector})
	if err != nil {
		return nil, errors.Wrap(err, "aggregate neighbors: search neighbors")
	}

	return NewKnnAggregator(res, vector).Aggregate(k, properties)
}

// TODO: this is business logic, move out of here
type KnnAggregator struct {
	input        search.Results
	sourceVector []float32
}

func NewKnnAggregator(input search.Results, sourceVector []float32) *KnnAggregator {
	return &KnnAggregator{input: input, sourceVector: sourceVector}
}

func (a *KnnAggregator) Aggregate(k int, properties []string) ([]classification.NeighborRef, error) {
	neighbors, err := a.extractBeacons(properties)
	if err != nil {
		return nil, errors.Wrap(err, "aggregate: extract beacons from neighbors")
	}

	return a.aggregateBeacons(neighbors)
}

func (a *KnnAggregator) extractBeacons(properties []string) (neighborProps, error) {
	neighbors := neighborProps{}
	for i, elem := range a.input {
		schemaMap, ok := elem.Schema.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("expecteded element[%d].Schema to be map, got: %T", i, elem.Schema)
		}

		for _, prop := range properties {
			refProp, ok := schemaMap[prop]
			if !ok {
				return nil, fmt.Errorf("expecteded element[%d].Schema to have property %q, but didn't", i, prop)
			}

			refTyped, ok := refProp.(models.MultipleRef)
			if !ok {
				return nil, fmt.Errorf("expecteded element[%d].Schema.%s to be models.MultipleRef, got: %T", i, prop, refProp)
			}

			if len(refTyped) != 1 {
				return nil, fmt.Errorf("a knn training data object needs to have exactly one label: "+
					"expecteded element[%d].Schema.%s to have exactly one reference, got: %d",
					i, prop, len(refTyped))
			}

			distance, err := vectorizer.NormalizedDistance(a.sourceVector, elem.Vector)
			if err != nil {
				return nil, errors.Wrap(err, "calculate distance between source and candidate")
			}

			beacon := refTyped[0].Beacon.String()
			neighborProp := neighbors[prop]
			if neighborProp.beacons == nil {
				neighborProp.beacons = neighborBeacons{}
			}
			neighborProp.beacons[beacon] = append(neighborProp.beacons[beacon], distance)
			neighbors[prop] = neighborProp
		}
	}

	return neighbors, nil
}

func (a *KnnAggregator) aggregateBeacons(props neighborProps) ([]classification.NeighborRef, error) {
	var out []classification.NeighborRef
	for propName, prop := range props {
		var winningBeacon string
		var winningCount int
		var totalCount int

		for beacon, distances := range prop.beacons {
			totalCount += len(distances)
			if len(distances) > winningCount {
				winningBeacon = beacon
				winningCount = len(distances)
			}
		}

		distances := a.distances(prop.beacons, winningBeacon)
		out = append(out, classification.NeighborRef{
			Beacon:       strfmt.URI(winningBeacon),
			WinningCount: winningCount,
			OverallCount: totalCount,
			LosingCount:  totalCount - winningCount,
			Property:     propName,
			Distances:    distances,
		})
	}

	return out, nil
}

func (a *KnnAggregator) distances(beacons neighborBeacons,
	winner string,
) classification.NeighborRefDistances {
	out := classification.NeighborRefDistances{}

	var winningDistances []float32
	var losingDistances []float32

	for beacon, distances := range beacons {
		if beacon == winner {
			winningDistances = distances
		} else {
			losingDistances = append(losingDistances, distances...)
		}
	}

	if len(losingDistances) > 0 {
		mean := mean(losingDistances)
		out.MeanLosingDistance = &mean

		closest := min_custom(losingDistances)
		out.ClosestLosingDistance = &closest
	}

	out.ClosestOverallDistance = min_custom(append(winningDistances, losingDistances...))
	out.ClosestWinningDistance = min_custom(winningDistances)
	out.MeanWinningDistance = mean(winningDistances)

	return out
}

type neighborProps map[string]neighborProp

type neighborProp struct {
	beacons neighborBeacons
}

type neighborBeacons map[string][]float32

func mergeUserFilterWithRefCountFilter(userFilter *filters.LocalFilter, className string,
	properties []string, op filters.Operator, refCount int,
) *filters.LocalFilter {
	countFilters := make([]filters.Clause, len(properties))
	for i, prop := range properties {
		countFilters[i] = filters.Clause{
			Operator: op,
			Value: &filters.Value{
				Type:  schema.DataTypeInt,
				Value: refCount,
			},
			On: &filters.Path{
				Class:    schema.ClassName(className),
				Property: schema.PropertyName(prop),
			},
		}
	}

	var countRootClause filters.Clause
	if len(countFilters) == 1 {
		countRootClause = countFilters[0]
	} else {
		countRootClause = filters.Clause{
			Operands: countFilters,
			Operator: filters.OperatorAnd,
		}
	}

	rootFilter := &filters.LocalFilter{}
	if userFilter == nil {
		rootFilter.Root = &countRootClause
	} else {
		rootFilter.Root = &filters.Clause{
			Operator: filters.OperatorAnd, // so we can AND the refcount requirements and whatever custom filters, the user has
			Operands: []filters.Clause{*userFilter.Root, countRootClause},
		}
	}

	return rootFilter
}

func mean(in []float32) float32 {
	sum := float32(0)
	for _, v := range in {
		sum += v
	}

	return sum / float32(len(in))
}

func min_custom(in []float32) float32 {
	min := float32(math.MaxFloat32)
	for _, dist := range in {
		if dist < min {
			min = dist
		}
	}

	return min
}
