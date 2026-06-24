package plugin

import "testing"

func TestInspectIdentityCategoriesWarnings(t *testing.T) {
	manifest := Manifest{
		Identity: Identity{
			Categories: []string{"security", "security", "支付", ""},
		},
	}

	warnings := inspectIdentityCategories(manifest)
	if len(warnings) != 3 {
		t.Fatalf("分类 warning 数量不符合预期: %#v", warnings)
	}
}

func TestInspectIdentityCategoriesAllowsEmpty(t *testing.T) {
	warnings := inspectIdentityCategories(Manifest{})
	if len(warnings) != 0 {
		t.Fatalf("未声明分类不应产生 warning: %#v", warnings)
	}
}
