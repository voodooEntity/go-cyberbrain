package cerebrum

import (
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/go-cyberbrain/src/system/util"
)

type Demultiplexer struct {
}

func NewDemultiplexer() *Demultiplexer {
	return &Demultiplexer{}
}

func (d *Demultiplexer) Parse(entity transport.TransportEntity) []transport.TransportEntity {
	// prepare return data & some var initis
	var ret []transport.TransportEntity
	typeLookup := make(map[string]int)
	var typePointer [][]*transport.TransportEntity
	if 0 < len(entity.ChildRelations) {
		// collect children pointers grouped by type string
		for key := range entity.ChildRelations {
			if val, ok := typeLookup[entity.ChildRelations[key].Target.Type]; ok {
				typePointer[val] = append(typePointer[val], &(entity.ChildRelations[key].Target))
			} else {
				typePointer = append(typePointer, []*transport.TransportEntity{&(entity.ChildRelations[key].Target)})
				typeLookup[entity.ChildRelations[key].Target.Type] = len(typePointer) - 1
			}
		}

		// now we get the demultiplex each single one of them and build a second pointer list
		demultiplexedTypePointer := make([][]*transport.TransportEntity, len(typePointer))
		for typeId, typePointerList := range typePointer {
			for _, singlePointer := range typePointerList {
				demultiplexedTypePointer[typeId] = append(demultiplexedTypePointer[typeId], d.generateEntityPointerList(d.Parse(*singlePointer))...)
			}
		}

		// now we generate all possible recombinations in which each child entity Type  each type occurs once
		recombinations := d.generateRecombinations(demultiplexedTypePointer)

		for _, recombinationSet := range recombinations {
			var tmpChildren []transport.TransportRelation
			for key := range recombinationSet {
				tmpChildren = append(tmpChildren, transport.TransportRelation{
					Target: *recombinationSet[key],
				})
			}
			ret = append(ret, transport.TransportEntity{
				Type:           entity.Type,
				ID:             entity.ID,
				Value:          entity.Value,
				Context:        entity.Context,
				Properties:     util.CopyStringStringMap(entity.Properties),
				ChildRelations: tmpChildren,
			})
		}
	} else {
		ret = append(ret, transport.TransportEntity{
			Type:       entity.Type,
			ID:         entity.ID,
			Value:      entity.Value,
			Context:    entity.Context,
			Properties: util.CopyStringStringMap(entity.Properties),
		})
	}
	return ret
}

func (d *Demultiplexer) generateEntityPointerList(data []transport.TransportEntity) []*transport.TransportEntity {
	var ret []*transport.TransportEntity
	for k := range data {
		ret = append(ret, &(data[k]))
	}
	return ret
}

func (d *Demultiplexer) generateRecombinations(data [][]*transport.TransportEntity) [][]*transport.TransportEntity {
	if len(data) == 0 {
		return [][]*transport.TransportEntity{}
	}

	var result [][]*transport.TransportEntity

	// Get the first row of values from data
	firstRow := data[0]

	// Recursively generate recombinations for the remaining rows
	remainingRows := d.generateRecombinations(data[1:])

	// If there are no remaining rows, return the first row as the only combination
	if len(remainingRows) == 0 {
		for _, val := range firstRow {
			result = append(result, []*transport.TransportEntity{val})
		}
		return result
	}

	// Combine the values from the first row with each recombination of the remaining rows
	for _, val := range firstRow {
		for _, comb := range remainingRows {
			result = append(result, append([]*transport.TransportEntity{val}, comb...))
		}
	}

	return result
}
