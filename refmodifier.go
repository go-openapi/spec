package spec

import (
	"fmt"
	"log"
)

func modifyItemsRefs(target *Schema, basePath string) {
	if target.Items != nil {
		if target.Items.Schema != nil {
			modifyRefs(target.Items.Schema, basePath)
		}
		for i := range target.Items.Schemas {
			s := target.Items.Schemas[i]
			modifyRefs(&s, basePath)
			target.Items.Schemas[i] = s
		}
	}
}

func modifyRefs(target *Schema, basePath string) {
	log.Printf("BASEPATH is %s", basePath)
	/* This is the base case */
	if target.Ref.String() != "" {
		newURL := fmt.Sprintf("%s%s", basePath, target.Ref.String())
		target.Ref, _ = NewRef(newURL)
	}

	modifyItemsRefs(target, basePath)
	for i := range target.AllOf {
		modifyRefs(&target.AllOf[i], basePath)
	}
	for i := range target.AnyOf {
		modifyRefs(&target.AnyOf[i], basePath)
	}
	for i := range target.OneOf {
		modifyRefs(&target.OneOf[i], basePath)
	}
	if target.Not != nil {
		modifyRefs(target.Not, basePath)
	}
	for k := range target.Properties {
		s := target.Properties[k]
		modifyRefs(&s, basePath)
		target.Properties[k] = s
	}
	if target.AdditionalProperties != nil && target.AdditionalProperties.Schema != nil {
		modifyRefs(target.AdditionalProperties.Schema, basePath)
	}
	for k := range target.PatternProperties {
		s := target.PatternProperties[k]
		modifyRefs(&s, basePath)
		target.PatternProperties[k] = s
	}
	for k := range target.Dependencies {
		if target.Dependencies[k].Schema != nil {
			modifyRefs(target.Dependencies[k].Schema, basePath)
		}
	}
	if target.AdditionalItems != nil && target.AdditionalItems.Schema != nil {
		modifyRefs(target.AdditionalItems.Schema, basePath)
	}
	for k := range target.Definitions {
		s := target.Definitions[k]
		modifyRefs(&s, basePath)
		target.Definitions[k] = s
	}
}
