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

package nearVideo

import (
	"context"

	"github.com/pkg/errors"
	"github.com/weaviate/weaviate/entities/dto"
	"github.com/weaviate/weaviate/entities/modulecapabilities"
	"github.com/weaviate/weaviate/entities/moduletools"
)

type Searcher[T dto.Embedding] struct {
	vectorizer bindVectorizer[T]
}

func NewSearcher[T dto.Embedding](vectorizer bindVectorizer[T]) *Searcher[T] {
	return &Searcher[T]{vectorizer}
}

type bindVectorizer[T dto.Embedding] interface {
	VectorizeVideo(ctx context.Context, video string, cfg moduletools.ClassConfig) (T, error)
}

func (s *Searcher[T]) VectorSearches() map[string]modulecapabilities.VectorForParams[T] {
	vectorSearches := map[string]modulecapabilities.VectorForParams[T]{}
	vectorSearches["nearVideo"] = &vectorForParams[T]{s.vectorizer}
	return vectorSearches
}

type vectorForParams[T dto.Embedding] struct {
	vectorizer bindVectorizer[T]
}

func (v *vectorForParams[T]) VectorForParams(ctx context.Context, params interface{}, className string,
	findVectorFn modulecapabilities.FindVectorFn[T],
	cfg moduletools.ClassConfig,
) (T, error) {
	// find vector for given search query
	vector, err := v.vectorizer.VectorizeVideo(ctx, params.(*NearVideoParams).Video, cfg)
	if err != nil {
		return nil, errors.Errorf("vectorize video: %v", err)
	}

	return vector, nil
}
