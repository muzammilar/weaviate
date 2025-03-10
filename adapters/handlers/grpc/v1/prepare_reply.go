//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2024 Weaviate B.V. All rights reserved.
//
//  CONTACT: hello@weaviate.io
//

package v1

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/weaviate/weaviate/adapters/handlers/grpc/v1/generative"
	"github.com/weaviate/weaviate/usecases/byteops"

	"github.com/weaviate/weaviate/entities/models"
	"github.com/weaviate/weaviate/entities/schema"
	generate "github.com/weaviate/weaviate/usecases/modulecomponents/additional/generate"
	additionalModels "github.com/weaviate/weaviate/usecases/modulecomponents/additional/models"

	"github.com/go-openapi/strfmt"
	"github.com/pkg/errors"
	"github.com/weaviate/weaviate/entities/additional"
	"github.com/weaviate/weaviate/entities/dto"
	"github.com/weaviate/weaviate/entities/search"
	pb "github.com/weaviate/weaviate/grpc/generated/protocol/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

type mapper interface {
	NewPrimitiveValue(v interface{}, dt schema.DataType) (*pb.Value, error)
	NewNestedValue(v interface{}, dt schema.DataType, parent schema.PropertyInterface, prop search.SelectProperty) (*pb.Value, error)
	NewNilValue() *pb.Value
}

type generativeReplier interface {
	Extract(_additional map[string]any, params any, metadata *pb.MetadataResult) (*pb.GenerativeResult, *pb.GenerativeResult, string, error)
}

type Replier struct {
	generative generativeReplier
	mapper     mapper
	uses123    bool
	logger     logrus.FieldLogger
}

type generativeQueryParamsGetter interface {
	ProviderName() string
	ReturnMetadata() bool
}

func NewReplier(uses123 bool,
	uses125 bool,
	uses127 bool,
	generativeParams generativeQueryParamsGetter,
	logger logrus.FieldLogger,
) *Replier {
	return &Replier{
		generative: generative.NewReplier(logger, generativeParams.ProviderName, generativeParams.ReturnMetadata, uses127),
		mapper:     &Mapper{uses125: uses125},
		uses123:    uses123,
		logger:     logger,
	}
}

func (r *Replier) Search(res []interface{}, start time.Time, searchParams dto.GetParams, scheme schema.Schema) (*pb.SearchReply, error) {
	tookSeconds := float64(time.Since(start)) / float64(time.Second)
	out := &pb.SearchReply{
		Took:                    float32(tookSeconds),
		GenerativeGroupedResult: new(string), // pointer to empty string
	}

	if searchParams.GroupBy != nil {
		out.GroupByResults = make([]*pb.GroupByResult, len(res))
		for i, raw := range res {
			group, generativeGroupResponse, err := r.extractGroup(raw, searchParams, scheme)
			if err != nil {
				return nil, err
			}
			if generativeGroupResponse != "" {
				out.GenerativeGroupedResult = &generativeGroupResponse
			}
			out.GroupByResults[i] = group
		}
	} else {
		objects, generativeGroupedResult, generativeGroupedResults, err := r.extractObjectsToResults(res, searchParams, scheme, false)
		if err != nil {
			return nil, err
		}
		out.GenerativeGroupedResult = &generativeGroupedResult
		out.GenerativeGroupedResults = generativeGroupedResults
		out.Results = objects
	}
	return out, nil
}

func (r *Replier) extractObjectsToResults(res []interface{}, searchParams dto.GetParams, scheme schema.Schema, fromGroup bool) ([]*pb.SearchResult, string, *pb.GenerativeResult, error) {
	results := make([]*pb.SearchResult, len(res))
	generativeGroupResultsReturnDeprecated := ""
	var generativeGroupResults *pb.GenerativeResult
	for i, raw := range res {
		asMap, ok := raw.(map[string]interface{})
		if !ok {
			return nil, "", nil, fmt.Errorf("could not parse returns %v", raw)
		}
		firstObject := i == 0

		var props *pb.PropertiesResult
		var err error

		if r.uses123 {
			props, err = r.extractPropertiesAnswer(scheme, asMap, searchParams.Properties, searchParams.ClassName, searchParams.AdditionalProperties)
		} else {
			props, err = r.extractPropertiesAnswerDeprecated(scheme, asMap, searchParams.Properties, searchParams.ClassName, searchParams.AdditionalProperties)
		}
		if err != nil {
			return nil, "", nil, err
		}

		additionalProps, err := r.extractAdditionalProps(asMap, searchParams.AdditionalProperties, firstObject, fromGroup)
		if err != nil {
			return nil, "", nil, err
		}

		if generativeGroupResultsReturnDeprecated == "" && additionalProps.GenerativeGroupedDeprecated != "" {
			generativeGroupResultsReturnDeprecated = additionalProps.GenerativeGroupedDeprecated
		}
		if generativeGroupResults == nil && additionalProps.GenerativeGrouped != nil {
			generativeGroupResults = additionalProps.GenerativeGrouped
		}

		result := &pb.SearchResult{
			Properties: props,
			Metadata:   additionalProps.Metadata,
			Generative: additionalProps.GenerativeSingle,
		}

		results[i] = result
	}
	return results, generativeGroupResultsReturnDeprecated, generativeGroupResults, nil
}

func idToByte(idRaw interface{}) ([]byte, string, error) {
	idStrfmt, ok := idRaw.(strfmt.UUID)
	if !ok {
		return nil, "", errors.New("could not extract format id in additional prop")
	}
	idStrfmtStr := idStrfmt.String()
	hexInteger, success := new(big.Int).SetString(strings.Replace(idStrfmtStr, "-", "", -1), 16)
	if !success {
		return nil, "", fmt.Errorf("failed to parse hex string to integer")
	}
	return hexInteger.Bytes(), idStrfmtStr, nil
}

func (r *Replier) extractAdditionalProps(asMap map[string]any, additionalPropsParams additional.Properties, firstObject, fromGroup bool) (*additionalProps, error) {
	generativeSearchRaw, generativeSearchEnabled := additionalPropsParams.ModuleParams["generate"]
	_, rerankEnabled := additionalPropsParams.ModuleParams["rerank"]

	addProps := &additionalProps{Metadata: &pb.MetadataResult{}}
	if additionalPropsParams.ID && !generativeSearchEnabled && !rerankEnabled && !fromGroup {
		idRaw, ok := asMap["id"]
		if !ok {
			return nil, errors.New("could not extract get id in additional prop")
		}

		idToBytes, idAsString, err := idToByte(idRaw)
		if err != nil {
			return nil, errors.Wrap(err, "could not extract format id in additional prop")
		}
		addProps.Metadata.Id = idAsString
		addProps.Metadata.IdAsBytes = idToBytes
	}
	_, ok := asMap["_additional"]
	if !ok {
		return addProps, nil
	}

	var additionalPropertiesMap map[string]interface{}
	if !fromGroup {
		additionalPropertiesMap = asMap["_additional"].(map[string]interface{})
	} else {
		addPropertiesGroup := asMap["_additional"].(*additional.GroupHitAdditional)
		additionalPropertiesMap = make(map[string]interface{}, 3)
		additionalPropertiesMap["id"] = addPropertiesGroup.ID
		additionalPropertiesMap["vector"] = addPropertiesGroup.Vector
		additionalPropertiesMap["distance"] = addPropertiesGroup.Distance
	}
	// id is part of the _additional map in case of generative search, group, & rerank - don't aks me why
	if additionalPropsParams.ID && (generativeSearchEnabled || fromGroup || rerankEnabled) {
		idRaw, ok := additionalPropertiesMap["id"]
		if !ok {
			return nil, errors.New("could not extract get id generative in additional prop")
		}

		idToBytes, idAsString, err := idToByte(idRaw)
		if err != nil {
			return nil, errors.Wrap(err, "could not extract format id in additional prop")
		}
		addProps.Metadata.Id = idAsString
		addProps.Metadata.IdAsBytes = idToBytes
	}

	if generativeSearchEnabled {
		singleGenerativeResult, groupedGenerativeResult, groupedDeprecated, err := r.generative.Extract(additionalPropertiesMap, generativeSearchRaw, addProps.Metadata)
		if err != nil {
			return nil, err
		}
		addProps.GenerativeSingle = singleGenerativeResult
		addProps.GenerativeGrouped = groupedGenerativeResult
		addProps.GenerativeGroupedDeprecated = groupedDeprecated
	}

	if rerankEnabled {
		rerank, ok := additionalPropertiesMap["rerank"]
		if !ok {
			return nil, errors.New("No results for rerank despite a search request. Is a the rerank module enabled?")
		}

		rerankFmt, ok := rerank.([]*additionalModels.RankResult)
		if !ok {
			return nil, errors.New("could not cast rerank result additional prop")
		}
		addProps.Metadata.RerankScore = *rerankFmt[0].Score
		addProps.Metadata.RerankScorePresent = true
	}

	// additional properties are only present for certain searches/configs => don't return an error if not available
	if additionalPropsParams.Vector {
		vector, ok := additionalPropertiesMap["vector"]
		if ok {
			vectorfmt, ok2 := vector.([]float32)
			if ok2 {
				addProps.Metadata.Vector = vectorfmt // deprecated, remove in a bit
				addProps.Metadata.VectorBytes = byteops.Fp32SliceToBytes(vectorfmt)
			}
		}
	}

	if len(additionalPropsParams.Vectors) > 0 {
		vectors, ok := additionalPropertiesMap["vectors"]
		if ok {
			vectorfmt, ok2 := vectors.(map[string]models.Vector)
			if ok2 {
				addProps.Metadata.Vectors = make([]*pb.Vectors, 0, len(vectorfmt))
				for name, vector := range vectorfmt {
					switch vec := vector.(type) {
					case []float32:
						addProps.Metadata.Vectors = append(addProps.Metadata.Vectors, &pb.Vectors{
							VectorBytes: byteops.Fp32SliceToBytes(vec),
							Name:        name,
							Type:        pb.Vectors_VECTOR_TYPE_SINGLE_FP32,
						})
					case [][]float32:
						addProps.Metadata.Vectors = append(addProps.Metadata.Vectors, &pb.Vectors{
							VectorBytes: byteops.Fp32SliceOfSlicesToBytes(vec),
							Name:        name,
							Type:        pb.Vectors_VECTOR_TYPE_MULTI_FP32,
						})
					default:
						// do nothing
					}
				}
			}
		}
	}

	if additionalPropsParams.Certainty {
		addProps.Metadata.CertaintyPresent = false
		certainty, ok := additionalPropertiesMap["certainty"]
		if ok {
			certaintyfmt, ok2 := certainty.(float64)
			if ok2 {
				addProps.Metadata.Certainty = float32(certaintyfmt)
				addProps.Metadata.CertaintyPresent = true
			}
		}
	}

	if additionalPropsParams.Distance {
		addProps.Metadata.DistancePresent = false
		distance, ok := additionalPropertiesMap["distance"]
		if ok {
			distancefmt, ok2 := distance.(float32)
			if ok2 {
				addProps.Metadata.Distance = distancefmt
				addProps.Metadata.DistancePresent = true
			}
		}
	}

	if additionalPropsParams.CreationTimeUnix {
		addProps.Metadata.CreationTimeUnixPresent = false
		creationtime, ok := additionalPropertiesMap["creationTimeUnix"]
		if ok {
			creationtimefmt, ok2 := creationtime.(int64)
			if ok2 {
				addProps.Metadata.CreationTimeUnix = creationtimefmt
				addProps.Metadata.CreationTimeUnixPresent = true
			}
		}
	}

	if additionalPropsParams.LastUpdateTimeUnix {
		addProps.Metadata.LastUpdateTimeUnixPresent = false
		lastUpdateTime, ok := additionalPropertiesMap["lastUpdateTimeUnix"]
		if ok {
			lastUpdateTimefmt, ok2 := lastUpdateTime.(int64)
			if ok2 {
				addProps.Metadata.LastUpdateTimeUnix = lastUpdateTimefmt
				addProps.Metadata.LastUpdateTimeUnixPresent = true
			}
		}
	}

	if additionalPropsParams.ExplainScore {
		addProps.Metadata.ExplainScorePresent = false
		explainScore, ok := additionalPropertiesMap["explainScore"]
		if ok {
			explainScorefmt, ok2 := explainScore.(string)
			if ok2 {
				addProps.Metadata.ExplainScore = explainScorefmt
				addProps.Metadata.ExplainScorePresent = true
			}
		}
	}

	if additionalPropsParams.Score {
		addProps.Metadata.ScorePresent = false
		score, ok := additionalPropertiesMap["score"]
		if ok {
			scorefmt, ok2 := score.(float32)
			if ok2 {
				addProps.Metadata.Score = scorefmt
				addProps.Metadata.ScorePresent = true
			}
		}
	}

	if additionalPropsParams.IsConsistent {
		isConsistent, ok := additionalPropertiesMap["isConsistent"]
		if ok {
			isConsistentfmt, ok2 := isConsistent.(bool)
			if ok2 {
				addProps.Metadata.IsConsistent = &isConsistentfmt
				addProps.Metadata.IsConsistentPresent = true
			}
		}
	}

	return addProps, nil
}

func (r *Replier) extractGroup(raw any, searchParams dto.GetParams, scheme schema.Schema) (*pb.GroupByResult, string, error) {
	generativeSearchRaw, generativeSearchEnabled := searchParams.AdditionalProperties.ModuleParams["generate"]
	_, rerankEnabled := searchParams.AdditionalProperties.ModuleParams["rerank"]
	asMap, ok := raw.(map[string]interface{})
	if !ok {
		return nil, "", fmt.Errorf("cannot parse result %v", raw)
	}
	add, ok := asMap["_additional"]
	if !ok {
		return nil, "", fmt.Errorf("_additional is required for groups %v", asMap)
	}
	addProps, ok := add.(models.AdditionalProperties)
	if !ok {
		addProps, ok = add.(map[string]interface{})
	}
	if !ok {
		return nil, "", fmt.Errorf("cannot parse _additional %v", add)
	}
	groupRaw, ok := addProps["group"]
	if !ok {
		return nil, "", fmt.Errorf("group is not present %v", addProps)
	}
	group, ok := groupRaw.(*additional.Group)
	if !ok {
		return nil, "", fmt.Errorf("cannot parse _additional %v", groupRaw)
	}

	ret := &pb.GroupByResult{
		Name:            group.GroupedBy.Value,
		MaxDistance:     group.MaxDistance,
		MinDistance:     group.MinDistance,
		NumberOfObjects: int64(group.Count),
	}

	groupedGenerativeResults := ""
	if generativeSearchEnabled {
		generateFmt, err := extractGenerateResult(addProps)
		if err != nil {
			return nil, "", err
		}

		generativeSearch, ok := generativeSearchRaw.(*generate.Params)
		if !ok {
			return nil, "", errors.New("could not cast generative search params")
		}
		if generativeSearch.Prompt != nil && generateFmt.SingleResult == nil {
			return nil, "", errors.New("No results for generative search despite a search request. Is a generative module enabled?")
		}

		if generateFmt.Error != nil {
			return nil, "", generateFmt.Error
		}

		if generateFmt.SingleResult != nil && *generateFmt.SingleResult != "" {
			ret.Generative = &pb.GenerativeReply{Result: *generateFmt.SingleResult}
		}

		// grouped results are only added to the first object for GQL reasons
		// however, reranking can result in a different order, so we need to check every object
		// recording the result if it's present assuming that it is at least somewhere and will be caught
		if generateFmt.GroupedResult != nil && *generateFmt.GroupedResult != "" {
			groupedGenerativeResults = *generateFmt.GroupedResult
		}
	}

	if rerankEnabled {
		rerankRaw, ok := addProps["rerank"]
		if !ok {
			return nil, "", fmt.Errorf("rerank is not present %v", addProps)
		}

		rerank, ok := rerankRaw.([]*additionalModels.RankResult)
		if !ok {
			return nil, "", fmt.Errorf("cannot parse rerank %v", rerankRaw)
		}
		ret.Rerank = &pb.RerankReply{
			Score: *rerank[0].Score,
		}
	}

	// group results does not support more additional properties
	searchParams.AdditionalProperties = additional.Properties{
		ID:       searchParams.AdditionalProperties.ID,
		Vector:   searchParams.AdditionalProperties.Vector,
		Distance: searchParams.AdditionalProperties.Distance,
	}

	// group objects are returned as a different type than normal results ([]map[string]interface{} vs []interface). As
	// the normal path is used much more often than groupBy, convert the []map[string]interface{} to []interface{}, even
	// though we cast it to map[string]interface{} in the extraction function.
	// This way we only do a copy for groupBy and not for the standard code-path which is used more often
	returnObjectsUntyped := make([]interface{}, len(group.Hits))
	for i := range returnObjectsUntyped {
		returnObjectsUntyped[i] = group.Hits[i]
	}

	objects, _, _, err := r.extractObjectsToResults(returnObjectsUntyped, searchParams, scheme, true)
	if err != nil {
		return nil, "", errors.Wrap(err, "extracting hits from group")
	}

	ret.Objects = objects

	return ret, groupedGenerativeResults, nil
}

func (r *Replier) extractPropertiesAnswerDeprecated(scheme schema.Schema, results map[string]interface{}, properties search.SelectProperties, className string, additionalPropsParams additional.Properties) (*pb.PropertiesResult, error) {
	nonRefProps := make(map[string]interface{}, 0)
	refProps := make([]*pb.RefPropertiesResult, 0)
	objProps := make([]*pb.ObjectProperties, 0)
	objArrayProps := make([]*pb.ObjectArrayProperties, 0)
	for _, prop := range properties {
		propRaw, ok := results[prop.Name]
		if !ok {
			continue
		}
		if prop.IsPrimitive {
			nonRefProps[prop.Name] = propRaw
			continue
		}
		if prop.IsObject {
			class := scheme.GetClass(className)
			if class == nil {
				return nil, fmt.Errorf("could not find class %s in schema", className)
			}
			nested, err := schema.GetPropertyByName(class, prop.Name)
			if err != nil {
				return nil, errors.Wrap(err, "getting property")
			}
			singleObj, ok := propRaw.(map[string]interface{})
			if ok {
				extractedNestedProp, err := extractPropertiesNested(scheme.GetClass, singleObj, prop, className, &Property{Property: nested})
				if err != nil {
					return nil, errors.Wrap(err, "extracting nested properties")
				}
				objProps = append(objProps, &pb.ObjectProperties{
					PropName: prop.Name,
					Value:    extractedNestedProp,
				})
				continue
			}
			arrayObjs, ok := propRaw.([]interface{})
			if ok {
				extractedNestedProps := make([]*pb.ObjectPropertiesValue, 0, len(arrayObjs))
				for _, obj := range arrayObjs {
					singleObj, ok := obj.(map[string]interface{})
					if !ok {
						continue
					}
					extractedNestedProp, err := extractPropertiesNested(scheme.GetClass, singleObj, prop, className, &Property{Property: nested})
					if err != nil {
						return nil, err
					}
					extractedNestedProps = append(extractedNestedProps, extractedNestedProp)
				}
				objArrayProps = append(objArrayProps,
					&pb.ObjectArrayProperties{
						PropName: prop.Name,
						Values:   extractedNestedProps,
					},
				)
				continue
			}
		}
		refs, ok := propRaw.([]interface{})
		if !ok {
			continue
		}
		extractedRefProps := make([]*pb.PropertiesResult, 0, len(refs))
		for _, ref := range refs {
			refLocal, ok := ref.(search.LocalRef)
			if !ok {
				continue
			}
			extractedRefProp, err := r.extractPropertiesAnswerDeprecated(scheme, refLocal.Fields, prop.Refs[0].RefProperties, refLocal.Class, additionalPropsParams)
			if err != nil {
				continue
			}
			additionalProps, err := r.extractAdditionalProps(refLocal.Fields, prop.Refs[0].AdditionalProperties, false, false)
			if err != nil {
				return nil, err
			}
			if additionalProps == nil {
				return nil, fmt.Errorf("additional props are nil somehow")
			}
			extractedRefProp.Metadata = additionalProps.Metadata
			extractedRefProps = append(extractedRefProps, extractedRefProp)
		}

		refProp := pb.RefPropertiesResult{PropName: prop.Name, Properties: extractedRefProps}
		refProps = append(refProps, &refProp)
	}
	props := pb.PropertiesResult{}
	if len(nonRefProps) > 0 {
		outProps := pb.ObjectPropertiesValue{}
		if err := extractArrayTypesRoot(scheme.GetClass, className, nonRefProps, &outProps); err != nil {
			return nil, errors.Wrap(err, "extracting non-primitive types")
		}
		newStruct, err := structpb.NewStruct(nonRefProps)
		if err != nil {
			return nil, errors.Wrap(err, "creating non-ref-prop struct")
		}
		props.NonRefProperties = newStruct
		props.IntArrayProperties = outProps.IntArrayProperties
		props.NumberArrayProperties = outProps.NumberArrayProperties
		props.TextArrayProperties = outProps.TextArrayProperties
		props.BooleanArrayProperties = outProps.BooleanArrayProperties
		props.ObjectProperties = outProps.ObjectProperties
		props.ObjectArrayProperties = outProps.ObjectArrayProperties
	}
	if len(refProps) > 0 {
		props.RefProps = refProps
	}
	if len(objProps) > 0 {
		props.ObjectProperties = objProps
	}
	if len(objArrayProps) > 0 {
		props.ObjectArrayProperties = objArrayProps
	}

	props.TargetCollection = className
	return &props, nil
}

func (r *Replier) extractPropertiesAnswer(scheme schema.Schema, results map[string]interface{}, properties search.SelectProperties, className string, additionalPropsParams additional.Properties) (*pb.PropertiesResult, error) {
	nonRefProps := &pb.Properties{
		Fields: make(map[string]*pb.Value, 0),
	}
	refProps := make([]*pb.RefPropertiesResult, 0)
	for _, prop := range properties {
		propRaw, ok := results[prop.Name]

		if !ok {
			if prop.IsPrimitive || prop.IsObject {
				nonRefProps.Fields[prop.Name] = r.mapper.NewNilValue()
			}
			continue
		}
		if prop.IsPrimitive {
			class := scheme.GetClass(className)
			if class == nil {
				return nil, fmt.Errorf("could not find class %s in schema", className)
			}
			dataType, err := schema.GetPropertyDataType(class, prop.Name)
			if err != nil {
				return nil, errors.Wrap(err, "getting primitive property datatype")
			}
			value, err := r.mapper.NewPrimitiveValue(propRaw, *dataType)
			if err != nil {
				return nil, errors.Wrapf(err, "creating primitive value for %v", prop.Name)
			}
			nonRefProps.Fields[prop.Name] = value
			continue
		}
		if prop.IsObject {
			class := scheme.GetClass(className)
			if class == nil {
				return nil, fmt.Errorf("could not find class %s in schema", className)
			}
			nested, err := schema.GetPropertyByName(class, prop.Name)
			if err != nil {
				return nil, errors.Wrap(err, "getting nested property")
			}
			value, err := r.mapper.NewNestedValue(propRaw, schema.DataType(nested.DataType[0]), &Property{Property: nested}, prop)
			if err != nil {
				return nil, errors.Wrap(err, "creating object value")
			}
			nonRefProps.Fields[prop.Name] = value
			continue
		}
		refs, ok := propRaw.([]interface{})
		if !ok {
			continue
		}
		extractedRefProps := make([]*pb.PropertiesResult, 0, len(refs))
		for _, ref := range refs {
			refLocal, ok := ref.(search.LocalRef)
			if !ok {
				continue
			}
			extractedRefProp, err := r.extractPropertiesAnswer(scheme, refLocal.Fields, prop.Refs[0].RefProperties, refLocal.Class, additionalPropsParams)
			if err != nil {
				continue
			}
			additionalProps, err := r.extractAdditionalProps(refLocal.Fields, prop.Refs[0].AdditionalProperties, false, false)
			if err != nil {
				return nil, err
			}
			if additionalProps == nil {
				return nil, fmt.Errorf("additional props are nil somehow")
			}
			extractedRefProp.Metadata = additionalProps.Metadata
			extractedRefProps = append(extractedRefProps, extractedRefProp)
		}

		refProp := pb.RefPropertiesResult{PropName: prop.Name, Properties: extractedRefProps}
		refProps = append(refProps, &refProp)
	}
	props := pb.PropertiesResult{}
	if len(nonRefProps.Fields) != 0 {
		props.NonRefProps = nonRefProps
	}
	if len(refProps) != 0 {
		props.RefProps = refProps
	}
	props.RefPropsRequested = properties.HasRefs()
	props.TargetCollection = className
	return &props, nil
}

func extractPropertiesNested[P schema.PropertyInterface](getClass func(string) *models.Class, results map[string]interface{}, property search.SelectProperty, className string, parent P) (*pb.ObjectPropertiesValue, error) {
	primitiveProps := make(map[string]interface{}, 0)
	objProps := make([]*pb.ObjectProperties, 0)
	objArrayProps := make([]*pb.ObjectArrayProperties, 0)
	for _, prop := range property.Props {
		propRaw, ok := results[prop.Name]
		if !ok {
			continue
		}
		if prop.IsPrimitive {
			primitiveProps[prop.Name] = propRaw
			continue
		}
		if prop.IsObject {
			var err error
			objProps, objArrayProps, err = extractObjectProperties(getClass, propRaw, prop, className, parent, objProps, objArrayProps)
			if err != nil {
				return nil, err
			}
		}
	}
	props := pb.ObjectPropertiesValue{}
	if len(primitiveProps) > 0 {
		if err := extractArrayTypesNested(className, primitiveProps, &props, parent); err != nil {
			return nil, errors.Wrap(err, "extracting non-primitive types")
		}
		newStruct, err := structpb.NewStruct(primitiveProps)
		if err != nil {
			return nil, errors.Wrap(err, "creating non-ref-prop struct")
		}
		props.NonRefProperties = newStruct
	}
	if len(objProps) > 0 {
		props.ObjectProperties = objProps
	}
	if len(objArrayProps) > 0 {
		props.ObjectArrayProperties = objArrayProps
	}
	return &props, nil
}

func extractObjectProperties[P schema.PropertyInterface](getClass func(string) *models.Class, propRaw interface{}, property search.SelectProperty, className string, parent P, objProps []*pb.ObjectProperties, objArrayProps []*pb.ObjectArrayProperties) ([]*pb.ObjectProperties, []*pb.ObjectArrayProperties, error) {
	prop, ok := propRaw.(map[string]interface{})
	if ok {
		objProp, err := extractObjectSingleProperties(getClass, prop, property, className, parent)
		if err != nil {
			return objProps, objArrayProps, err
		}
		objProps = append(objProps, objProp)
	}
	propArray, ok := propRaw.([]interface{})
	if ok {
		objArrayProp, err := extractObjectArrayProperties(getClass, propArray, property, className, parent)
		if err != nil {
			return objProps, objArrayProps, err
		}
		objArrayProps = append(objArrayProps, objArrayProp)
	}
	return objProps, objArrayProps, nil
}

func extractObjectSingleProperties[P schema.PropertyInterface](getClass func(string) *models.Class, prop map[string]interface{}, property search.SelectProperty, className string, parent P) (*pb.ObjectProperties, error) {
	nested, err := schema.GetNestedPropertyByName(parent, property.Name)
	if err != nil {
		return nil, errors.Wrap(err, "getting property")
	}
	extractedNestedProp, err := extractPropertiesNested(getClass, prop, property, className, &NestedProperty{NestedProperty: nested})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("extracting nested properties from %v", nested))
	}
	return &pb.ObjectProperties{
		PropName: property.Name,
		Value:    extractedNestedProp,
	}, nil
}

func extractObjectArrayProperties[P schema.PropertyInterface](getClass func(string) *models.Class, propObjs []interface{}, property search.SelectProperty, className string, parent P) (*pb.ObjectArrayProperties, error) {
	extractedNestedProps := make([]*pb.ObjectPropertiesValue, 0, len(propObjs))
	for _, objRaw := range propObjs {
		nested, err := schema.GetNestedPropertyByName(parent, property.Name)
		if err != nil {
			return nil, errors.Wrap(err, "getting property")
		}
		obj, ok := objRaw.(map[string]interface{})
		if !ok {
			continue
		}
		extractedNestedProp, err := extractPropertiesNested(getClass, obj, property, className, &NestedProperty{NestedProperty: nested})
		if err != nil {
			return nil, errors.Wrap(err, "extracting nested properties")
		}
		extractedNestedProps = append(extractedNestedProps, extractedNestedProp)
	}
	return &pb.ObjectArrayProperties{
		PropName: property.Name,
		Values:   extractedNestedProps,
	}, nil
}

func extractArrayTypesRoot(getClass func(string) *models.Class, className string, rawProps map[string]interface{}, props *pb.ObjectPropertiesValue) error {
	dataTypes := make(map[string]*schema.DataType, 0)
	class := getClass(className)
	if class == nil {
		return fmt.Errorf("could not find class %s in schema", className)
	}
	for propName := range rawProps {
		dataType, err := schema.GetPropertyDataType(class, propName)
		if err != nil {
			return err
		}
		dataTypes[propName] = dataType
	}
	return extractArrayTypes(rawProps, props, dataTypes)
}

func extractArrayTypesNested[P schema.PropertyInterface](className string, rawProps map[string]interface{}, props *pb.ObjectPropertiesValue, parent P) error {
	dataTypes := make(map[string]*schema.DataType, 0)
	for propName := range rawProps {
		dataType, err := schema.GetNestedPropertyDataType(parent, propName)
		if err != nil {
			return err
		}
		dataTypes[propName] = dataType
	}
	return extractArrayTypes(rawProps, props, dataTypes)
}

// slices cannot be part of a grpc struct, so we need to handle each of them separately
func extractArrayTypes(rawProps map[string]interface{}, props *pb.ObjectPropertiesValue, dataTypes map[string]*schema.DataType) error {
	for propName, prop := range rawProps {
		dataType := dataTypes[propName]
		switch *dataType {
		case schema.DataTypeIntArray:
			propIntAsFloat, ok := prop.([]float64)
			if !ok {
				emptyArr, ok := prop.([]interface{})
				if ok && len(emptyArr) == 0 {
					continue
				}
				return fmt.Errorf("property %v with datatype %v needs to be []float64, got %T", propName, dataType, prop)
			}
			propInt := make([]int64, len(propIntAsFloat))
			for i := range propIntAsFloat {
				propInt[i] = int64(propIntAsFloat[i])
			}
			if props.IntArrayProperties == nil {
				props.IntArrayProperties = make([]*pb.IntArrayProperties, 0)
			}
			props.IntArrayProperties = append(props.IntArrayProperties, &pb.IntArrayProperties{PropName: propName, Values: propInt})
			delete(rawProps, propName)
		case schema.DataTypeNumberArray:
			propFloat, ok := prop.([]float64)
			if !ok {
				emptyArr, ok := prop.([]interface{})
				if ok && len(emptyArr) == 0 {
					continue
				}
				return fmt.Errorf("property %v with datatype %v needs to be []float64, got %T", propName, dataType, prop)
			}

			if props.NumberArrayProperties == nil {
				props.NumberArrayProperties = make([]*pb.NumberArrayProperties, 0)
			}
			props.NumberArrayProperties = append(
				props.NumberArrayProperties,
				&pb.NumberArrayProperties{PropName: propName, ValuesBytes: byteops.Fp64SliceToBytes(propFloat), Values: propFloat},
			)
			delete(rawProps, propName)
		case schema.DataTypeStringArray, schema.DataTypeTextArray, schema.DataTypeDateArray, schema.DataTypeUUIDArray:
			propString, ok := prop.([]string)
			if !ok {
				emptyArr, ok := prop.([]interface{})
				if ok && len(emptyArr) == 0 {
					continue
				}
				return fmt.Errorf("property %v with datatype %v needs to be []string, got %T", propName, dataType, prop)
			}
			if props.TextArrayProperties == nil {
				props.TextArrayProperties = make([]*pb.TextArrayProperties, 0)
			}
			props.TextArrayProperties = append(props.TextArrayProperties, &pb.TextArrayProperties{PropName: propName, Values: propString})
			delete(rawProps, propName)
		case schema.DataTypeBooleanArray:
			propBool, ok := prop.([]bool)
			if !ok {
				emptyArr, ok := prop.([]interface{})
				if ok && len(emptyArr) == 0 {
					continue
				}
				return fmt.Errorf("property %v with datatype %v needs to be []bool, got %T", propName, dataType, prop)
			}
			if props.BooleanArrayProperties == nil {
				props.BooleanArrayProperties = make([]*pb.BooleanArrayProperties, 0)
			}
			props.BooleanArrayProperties = append(props.BooleanArrayProperties, &pb.BooleanArrayProperties{PropName: propName, Values: propBool})
			delete(rawProps, propName)
		default:
			_, isArray := schema.IsArrayType(*dataType)
			if isArray {
				return fmt.Errorf("property %v with array type not handled %v", propName, dataType)
			}
		}
	}
	return nil
}

func extractGenerateResult(additionalPropertiesMap map[string]interface{}) (*additionalModels.GenerateResult, error) {
	generateFmt := &additionalModels.GenerateResult{}
	if generate, ok := additionalPropertiesMap["generate"]; ok {
		generateParams, ok := generate.(map[string]interface{})
		if !ok {
			return nil, errors.New("could not cast generative result additional prop")
		}
		if generateParams["singleResult"] != nil {
			if singleResult, ok := generateParams["singleResult"].(*string); ok {
				generateFmt.SingleResult = singleResult
			}
		}
		if generateParams["groupedResult"] != nil {
			if groupedResult, ok := generateParams["groupedResult"].(*string); ok {
				generateFmt.GroupedResult = groupedResult
			}
		}
		if generateParams["error"] != nil {
			if err, ok := generateParams["error"].(error); ok {
				generateFmt.Error = err
			}
		}
	}
	return generateFmt, nil
}

type additionalProps struct {
	Metadata                    *pb.MetadataResult
	GenerativeSingle            *pb.GenerativeResult
	GenerativeGrouped           *pb.GenerativeResult
	GenerativeGroupedDeprecated string
}
