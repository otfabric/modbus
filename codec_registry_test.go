package modbus

import (
	"testing"
)

func TestAvailableCodecDescriptors_ReturnsCopy(t *testing.T) {
	all := AvailableCodecDescriptors()
	if all == nil {
		t.Fatal("AvailableCodecDescriptors() must not return nil")
	}
	// Second call returns independent copy (e.g. for deep-copy of Layouts)
	all2 := AvailableCodecDescriptors()
	if len(all) != len(all2) {
		t.Errorf("two calls returned different lengths: %d vs %d", len(all), len(all2))
	}
}

func TestCodecDescriptorByID_NotFound(t *testing.T) {
	_, ok := CodecDescriptorByID("nonexistent/id")
	if ok {
		t.Error("expected false for nonexistent ID")
	}
}

func TestCodecDescriptorsForRegisterCount_ReturnsOnlyMatchingCount(t *testing.T) {
	got := CodecDescriptorsForRegisterCount(2)
	if got == nil {
		got = []CodecDescriptor{}
	}
	for i, d := range got {
		if d.RegisterSpec.Count != 2 {
			t.Errorf("descriptor[%d] %s: RegisterSpec.Count = %d, want 2", i, d.ID, d.RegisterSpec.Count)
		}
	}
}

func TestCodecCandidatesForRegisterCount_ReturnsOnlyMatchingCount(t *testing.T) {
	got := CodecCandidatesForRegisterCount(2)
	if got == nil {
		got = []CodecCandidate{}
	}
	for i, c := range got {
		d, ok := CodecDescriptorByID(c.CodecID)
		if !ok {
			t.Errorf("candidate[%d] CodecID %q not found in registry", i, c.CodecID)
			continue
		}
		if d.RegisterSpec.Count != 2 {
			t.Errorf("candidate[%d] %s: descriptor RegisterSpec.Count = %d, want 2", i, c.CodecID, d.RegisterSpec.Count)
		}
	}
}

func TestFindCodecDescriptors_ZeroQuery_ReturnsAll(t *testing.T) {
	got := FindCodecDescriptors(CodecQuery{})
	all := AvailableCodecDescriptors()
	if len(got) != len(all) {
		t.Errorf("zero query: got %d descriptors, want %d (same as Available)", len(got), len(all))
	}
}

func TestFindCodecDescriptors_WithFilters(t *testing.T) {
	got := FindCodecDescriptors(CodecQuery{
		RegisterCount: 2,
		Family:        CodecFamilyInteger,
	})
	for i, d := range got {
		if d.RegisterSpec.Count != 2 {
			t.Errorf("descriptor[%d] RegisterSpec.Count = %d, want 2", i, d.RegisterSpec.Count)
		}
		if d.Family != CodecFamilyInteger {
			t.Errorf("descriptor[%d] Family = %v, want CodecFamilyInteger", i, d.Family)
		}
	}
}

// TestRegistryWithOneDescriptor verifies the registry returns a descriptor
// once registered (used when we add numeric codecs in Phase 4).
func TestRegistryWithOneDescriptor(t *testing.T) {
	// Save and restore registeredDescriptors so we don't leak state into other tests
	saved := registeredDescriptors
	defer func() { registeredDescriptors = saved }()

	registeredDescriptors = nil
	registerCodecDescriptor(CodecDescriptor{
		ID:           "uint32/layout:4321",
		Name:         "uint32",
		Family:       CodecFamilyInteger,
		ValueKind:    CodecValueUint32,
		RegisterSpec: RegisterSpec{Count: 2},
		ByteSpec:     ByteSpec{Count: 4},
		Layouts: []RegisterLayoutDescriptor{
			{Name: "4321", Common: true, Layout: Layout32_4321},
		},
	})

	all := AvailableCodecDescriptors()
	if len(all) != 1 {
		t.Fatalf("AvailableCodecDescriptors() len = %d, want 1", len(all))
	}
	if all[0].ID != "uint32/layout:4321" {
		t.Errorf("ID = %q, want \"uint32/layout:4321\"", all[0].ID)
	}

	d, ok := CodecDescriptorByID("uint32/layout:4321")
	if !ok {
		t.Fatal("CodecDescriptorByID: expected true")
	}
	if d.Name != "uint32" {
		t.Errorf("Name = %q, want \"uint32\"", d.Name)
	}

	forCount := CodecDescriptorsForRegisterCount(2)
	if len(forCount) != 1 {
		t.Errorf("CodecDescriptorsForRegisterCount(2) len = %d, want 1", len(forCount))
	}

	candidates := CodecCandidatesForRegisterCount(2)
	if len(candidates) != 1 {
		t.Fatalf("CodecCandidatesForRegisterCount(2) len = %d, want 1", len(candidates))
	}
	if candidates[0].CodecID != "uint32/layout:4321" {
		t.Errorf("CodecID = %q", candidates[0].CodecID)
	}
	if candidates[0].LayoutName != "4321" {
		t.Errorf("LayoutName = %q, want \"4321\"", candidates[0].LayoutName)
	}
}
