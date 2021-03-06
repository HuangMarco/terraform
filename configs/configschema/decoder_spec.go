package configschema

import (
	"github.com/hashicorp/hcl2/hcldec"
)

var mapLabelNames = []string{"key"}

// DecoderSpec returns a hcldec.Spec that can be used to decode a HCL Body
// using the facilities in the hcldec package.
//
// The returned specification is guaranteed to return a value of the same type
// returned by method ImpliedType, but it may contain null values if any of the
// block attributes are defined as optional and/or computed respectively.
func (b *Block) DecoderSpec() hcldec.Spec {
	ret := hcldec.ObjectSpec{}
	if b == nil {
		return ret
	}

	for name, attrS := range b.Attributes {
		ret[name] = &hcldec.AttrSpec{
			Name:     name,
			Type:     attrS.Type,
			Required: attrS.Required,
		}
	}

	for name, blockS := range b.BlockTypes {
		if _, exists := ret[name]; exists {
			// This indicates an invalid schema, since it's not valid to
			// define both an attribute and a block type of the same name.
			// However, we don't raise this here since it's checked by
			// InternalValidate.
			continue
		}

		childSpec := blockS.Block.DecoderSpec()

		switch blockS.Nesting {
		case NestingSingle:
			ret[name] = &hcldec.BlockSpec{
				TypeName: name,
				Nested:   childSpec,
				Required: blockS.MinItems == 1 && blockS.MaxItems >= 1,
			}
		case NestingList:
			// We prefer to use a list where possible, since it makes our
			// implied type more complete, but if there are any
			// dynamically-typed attributes inside we must use a tuple
			// instead, at the expense of our type then not being predictable.
			if blockS.Block.ImpliedType().HasDynamicTypes() {
				ret[name] = &hcldec.BlockTupleSpec{
					TypeName: name,
					Nested:   childSpec,
					MinItems: blockS.MinItems,
					MaxItems: blockS.MaxItems,
				}
			} else {
				ret[name] = &hcldec.BlockListSpec{
					TypeName: name,
					Nested:   childSpec,
					MinItems: blockS.MinItems,
					MaxItems: blockS.MaxItems,
				}
			}
		case NestingSet:
			// We forbid dynamically-typed attributes inside NestingSet in
			// InternalValidate, so we don't do anything special to handle
			// that here. (There is no set analog to tuple and object types,
			// because cty's set implementation depends on knowing the static
			// type in order to properly compute its internal hashes.)
			ret[name] = &hcldec.BlockSetSpec{
				TypeName: name,
				Nested:   childSpec,
				MinItems: blockS.MinItems,
				MaxItems: blockS.MaxItems,
			}
		case NestingMap:
			// We prefer to use a list where possible, since it makes our
			// implied type more complete, but if there are any
			// dynamically-typed attributes inside we must use a tuple
			// instead, at the expense of our type then not being predictable.
			if blockS.Block.ImpliedType().HasDynamicTypes() {
				ret[name] = &hcldec.BlockObjectSpec{
					TypeName:   name,
					Nested:     childSpec,
					LabelNames: mapLabelNames,
				}
			} else {
				ret[name] = &hcldec.BlockMapSpec{
					TypeName:   name,
					Nested:     childSpec,
					LabelNames: mapLabelNames,
				}
			}
		default:
			// Invalid nesting type is just ignored. It's checked by
			// InternalValidate.
			continue
		}
	}

	return ret
}
