package mojovdl

import (
	"reflect"

	"v.io/v23/vdl"
)

func Encode(value interface{}) ([]byte, error) {
	enc := &encoder{
		allocator: &allocator{},
	}
	err := vdl.FromReflect(enc, reflect.ValueOf(value))
	return enc.Bytes(), err
}

type encoder struct {
	allocator *allocator
}

func (e *encoder) Bytes() []byte {
	return e.allocator.AllocatedBytes()
}

func (e *encoder) FromBool(src bool, tt *vdl.Type) error {
	panic("cannot encode top level bool")
}
func (e *encoder) FromUint(src uint64, tt *vdl.Type) error {
	panic("cannot encode top level uint")
}
func (e *encoder) FromInt(src int64, tt *vdl.Type) error {
	panic("cannot encode top level int")
}
func (e *encoder) FromFloat(src float64, tt *vdl.Type) error {
	panic("cannot encode top level float")
}
func (e *encoder) FromComplex(src complex128, tt *vdl.Type) error {
	panic("cannot encode top level complex")
}
func (e *encoder) FromBytes(src []byte, tt *vdl.Type) error {
	panic("cannot encode top level bytes")
}
func (e *encoder) FromString(src string, tt *vdl.Type) error {
	panic("cannot encode top level string")
}
func (e *encoder) FromEnumLabel(src string, tt *vdl.Type) error {
	panic("cannot encode top level enum")
}
func (e *encoder) FromTypeObject(src *vdl.Type) error {
	panic("cannot encode top level type object")
}
func (e *encoder) FromNil(tt *vdl.Type) error {
	panic("cannot encode top level nil")
}

func (e *encoder) StartList(tt *vdl.Type, len int) (vdl.ListTarget, error) {
	panic("UNIMPLEMENTED")
	return nil, nil
}
func (e *encoder) FinishList(x vdl.ListTarget) error {
	return nil
}
func (e *encoder) StartSet(tt *vdl.Type, len int) (vdl.SetTarget, error) {
	panic("UNIMPLEMENTED")
}
func (e *encoder) FinishSet(x vdl.SetTarget) error {
	panic("UNIMPLEMENTED")

}
func (e *encoder) StartMap(tt *vdl.Type, len int) (vdl.MapTarget, error) {
	panic("UNIMPLEMENTED")

}
func (e *encoder) FinishMap(x vdl.MapTarget) error {
	panic("UNIMPLEMENTED")

}
func (e *encoder) StartFields(tt *vdl.Type) (vdl.FieldsTarget, error) {
	if tt.Kind() == vdl.Union {
		panic("not yet supported")
	}
	if tt.Kind() == vdl.Optional {
		tt = tt.Elem()
	}
	block := e.allocator.Allocate(neededStructAllocationSize(tt), 0)
	return fieldsTarget{
			vdlType: tt,
			block:   block,
		},
		nil
}
func (e *encoder) FinishFields(x vdl.FieldsTarget) error {
	return nil
}
