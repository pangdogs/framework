package variant

import "testing"

func TestVariantCreatorBuiltins(t *testing.T) {
	for _, tc := range sampleValues(t) {
		t.Run(tc.name, func(t *testing.T) {
			v, err := VariantCreator().New(tc.value.TypeId())
			if err != nil {
				t.Fatalf("VariantCreator.New failed: %v", err)
			}
			if v.TypeId() != tc.value.TypeId() {
				t.Fatalf("unexpected type id: got %d want %d", v.TypeId(), tc.value.TypeId())
			}

			rv, err := VariantCreator().NewReflected(tc.value.TypeId())
			if err != nil {
				t.Fatalf("VariantCreator.NewReflected failed: %v", err)
			}
			if !rv.IsValid() || rv.Interface().(Value).TypeId() != tc.value.TypeId() {
				t.Fatalf("unexpected reflected value: %#v", rv)
			}
		})
	}
}

func TestVariantCreatorDeclareAndTypeId(t *testing.T) {
	creator := _NewVariantCreator()

	if _, err := creator.New(TypeId_Customize + 1); err == nil {
		t.Fatal("expected ErrNotDeclared for empty creator")
	}

	creator.Declare(&customValue{})

	v, err := creator.New(customValueTypeID)
	if err != nil {
		t.Fatalf("custom creator New failed: %v", err)
	}
	if v.TypeId() != customValueTypeID {
		t.Fatalf("unexpected custom type id: %d", v.TypeId())
	}

	if got := GenTypeId(&customValue{}); got != customValueTypeID {
		t.Fatalf("unexpected GenTypeId result: %d", got)
	}
	if got := GenTypeIdT[customValue](); got != customValueTypeID {
		t.Fatalf("unexpected GenTypeIdT result: %d", got)
	}

	typeBuf := encodeReadable(t, customValueTypeID, customValueTypeID.Size())
	var decoded TypeId
	if n, err := decoded.Write(typeBuf); err != nil || n != len(typeBuf) {
		t.Fatalf("TypeId.Write failed: n=%d err=%v", n, err)
	}
	if decoded != customValueTypeID {
		t.Fatalf("unexpected decoded type id: %d", decoded)
	}
}

func TestVariantCreatorPanics(t *testing.T) {
	creator := _NewVariantCreator()

	requirePanic(t, func() {
		creator.Declare(nil)
	})

	creator.Declare(&customValue{})
	requirePanic(t, func() {
		creator.Declare(&customValue{})
	})
}

var customValueTypeID = GenTypeIdT[customValue]()

type customValue struct {
	Data String
}

func (v customValue) Read(p []byte) (int, error) {
	return v.Data.Read(p)
}

func (v *customValue) Write(p []byte) (int, error) {
	return (&v.Data).Write(p)
}

func (v customValue) Size() int {
	return v.Data.Size()
}

func (customValue) TypeId() TypeId {
	return customValueTypeID
}

func (v customValue) Indirect() any {
	return string(v.Data)
}
