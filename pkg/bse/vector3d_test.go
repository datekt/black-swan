package bse

import (
	"encoding/json"
	"math"
	"testing"
)

func TestVectorAdd(t *testing.T) {
	got := (Vector3D{1, 2, 3}).Add(Vector3D{4, 5, 6})
	want := Vector3D{5, 7, 9}
	if got != want {
		t.Fatalf("Add: got %v want %v", got, want)
	}
}

func TestVectorSub(t *testing.T) {
	got := (Vector3D{4, 5, 6}).Sub(Vector3D{1, 2, 3})
	want := Vector3D{3, 3, 3}
	if got != want {
		t.Fatalf("Sub: got %v want %v", got, want)
	}
}

func TestVectorScale(t *testing.T) {
	got := (Vector3D{1, -2, 3}).Scale(2.5)
	want := Vector3D{2.5, -5, 7.5}
	if got != want {
		t.Fatalf("Scale: got %v want %v", got, want)
	}
}

func TestVectorLength(t *testing.T) {
	if got := (Vector3D{3, 4, 0}).Length(); got != 5 {
		t.Fatalf("Length: got %v want 5", got)
	}
}

func TestVectorNormalize(t *testing.T) {
	got := (Vector3D{3, 4, 0}).Normalize()
	if math.Abs(got.Length()-1) > 1e-12 {
		t.Fatalf("Normalize length: got %v want 1", got.Length())
	}
	if zero := (Vector3D{0, 0, 0}).Normalize(); zero != (Vector3D{}) {
		t.Fatalf("Normalize zero: got %v want zero vector", zero)
	}
}

func TestVectorMarshalJSON(t *testing.T) {
	data, err := json.Marshal(Vector3D{1.5, -2.5, 3})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(data) != "[1.5,-2.5,3]" {
		t.Fatalf("marshal: got %s want [1.5,-2.5,3]", string(data))
	}
}

func TestVectorUnmarshalJSON(t *testing.T) {
	var v Vector3D
	if err := json.Unmarshal([]byte("[1,2,3]"), &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if v != (Vector3D{1, 2, 3}) {
		t.Fatalf("unmarshal: got %v", v)
	}
}

func TestVectorUnmarshalJSONRejectsObject(t *testing.T) {
	var v Vector3D
	if err := json.Unmarshal([]byte(`{"X":1,"Y":2,"Z":3}`), &v); err == nil {
		t.Fatal("unmarshal: expected error for object form")
	}
}

func TestVectorUnmarshalJSONRejectsWrongLength(t *testing.T) {
	var v Vector3D
	if err := json.Unmarshal([]byte("[1,2]"), &v); err == nil {
		t.Fatal("unmarshal: expected error for 2-element array")
	}
}
