// Copyright 2017 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spec

import (
	"fmt"
	"path"
	"strings"
)

func modifyItemsRefs(target *Schema, remoteURI string) {
	if target.Items != nil {
		if target.Items.Schema != nil {
			modifyRefs(target.Items.Schema, remoteURI)
		}
		for i := range target.Items.Schemas {
			s := target.Items.Schemas[i]
			modifyRefs(&s, remoteURI)
			target.Items.Schemas[i] = s
		}
	}
}

func modifyRefs(target *Schema, remoteURI string) {
	if target.Ref.String() != "" {
		if target.Ref.HasFullURL {
			return
		}
		if target.Ref.HasFragmentOnly {
			newURL := fmt.Sprintf("%s%s", remoteURI, target.Ref.String())
			target.Ref, _ = NewRef(newURL)
			return
		}
		// relative file path
		if !target.Ref.HasFullFilePath {
			// it is important to remove any trailing slashes
			newURL := path.Join(path.Dir(strings.TrimRight(remoteURI, "/")), target.Ref.String())
			target.Ref, _ = NewRef(newURL)
			return
		}
	}

	modifyItemsRefs(target, remoteURI)
	for i := range target.AllOf {
		modifyRefs(&target.AllOf[i], remoteURI)
	}
	for i := range target.AnyOf {
		modifyRefs(&target.AnyOf[i], remoteURI)
	}
	for i := range target.OneOf {
		modifyRefs(&target.OneOf[i], remoteURI)
	}
	if target.Not != nil {
		modifyRefs(target.Not, remoteURI)
	}
	for k := range target.Properties {
		s := target.Properties[k]
		modifyRefs(&s, remoteURI)
		target.Properties[k] = s
	}
	if target.AdditionalProperties != nil && target.AdditionalProperties.Schema != nil {
		modifyRefs(target.AdditionalProperties.Schema, remoteURI)
	}
	for k := range target.PatternProperties {
		s := target.PatternProperties[k]
		modifyRefs(&s, remoteURI)
		target.PatternProperties[k] = s
	}
	for k := range target.Dependencies {
		if target.Dependencies[k].Schema != nil {
			modifyRefs(target.Dependencies[k].Schema, remoteURI)
		}
	}
	if target.AdditionalItems != nil && target.AdditionalItems.Schema != nil {
		modifyRefs(target.AdditionalItems.Schema, remoteURI)
	}
	for k := range target.Definitions {
		s := target.Definitions[k]
		modifyRefs(&s, remoteURI)
		target.Definitions[k] = s
	}
}
