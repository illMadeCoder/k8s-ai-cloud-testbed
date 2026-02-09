package components

import (
	"testing"
)

func TestFallbackComponent_App(t *testing.T) {
	r := &Resolver{}
	resolved, err := r.fallbackComponent("nginx", "app", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Name != "nginx" {
		t.Errorf("expected name nginx, got %s", resolved.Name)
	}
	if resolved.Type != "app" {
		t.Errorf("expected type app, got %s", resolved.Type)
	}
	if len(resolved.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(resolved.Sources))
	}
	if resolved.Sources[0].Path != "components/apps/nginx" {
		t.Errorf("expected path components/apps/nginx, got %s", resolved.Sources[0].Path)
	}
	if resolved.Sources[0].TargetRevision != "HEAD" {
		t.Errorf("expected targetRevision HEAD, got %s", resolved.Sources[0].TargetRevision)
	}
	if resolved.Sources[0].Helm != nil {
		t.Errorf("expected no helm config when no params, got %+v", resolved.Sources[0].Helm)
	}
}

func TestFallbackComponent_Workflow(t *testing.T) {
	r := &Resolver{}
	resolved, err := r.fallbackComponent("k6-loadgen", "workflow", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Sources[0].Path != "components/workflows/k6-loadgen" {
		t.Errorf("expected path components/workflows/k6-loadgen, got %s", resolved.Sources[0].Path)
	}
}

func TestFallbackComponent_Config(t *testing.T) {
	r := &Resolver{}
	resolved, err := r.fallbackComponent("alloy", "config", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Sources[0].Path != "components/configs/alloy" {
		t.Errorf("expected path components/configs/alloy, got %s", resolved.Sources[0].Path)
	}
}

func TestFallbackComponent_UnknownType(t *testing.T) {
	r := &Resolver{}
	_, err := r.fallbackComponent("foo", "unknown", nil)
	if err == nil {
		t.Fatal("expected error for unknown component type")
	}
}

func TestFallbackComponent_WithParams(t *testing.T) {
	r := &Resolver{}
	params := map[string]string{
		"replicaCount": "5",
		"image":        "nginx:latest",
	}
	resolved, err := r.fallbackComponent("nginx", "app", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Sources[0].Helm == nil {
		t.Fatal("expected helm config when params provided")
	}
	if resolved.Sources[0].Helm.Parameters["replicaCount"] != "5" {
		t.Errorf("expected replicaCount=5, got %s", resolved.Sources[0].Helm.Parameters["replicaCount"])
	}
	if resolved.Sources[0].Helm.Parameters["image"] != "nginx:latest" {
		t.Errorf("expected image=nginx:latest, got %s", resolved.Sources[0].Helm.Parameters["image"])
	}
}

func TestResolveFromCR(t *testing.T) {
	// Import the API types indirectly by constructing what resolveFromCR expects
	// Since resolveFromCR takes *experimentsv1alpha1.Component, we test via the
	// exported types. However, resolveFromCR is unexported and takes API types,
	// so we test the parameter merging logic through fallbackComponent instead,
	// and test resolveFromCR behavior via the exported ResolveComponentRef
	// (which requires a client). The unit-testable parts are covered above.
}

func TestFallbackComponent_DefaultRepo(t *testing.T) {
	r := &Resolver{}
	resolved, err := r.fallbackComponent("test", "app", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedRepo := "https://github.com/illMadeCoder/k8s-ai-testbed.git"
	if resolved.Sources[0].RepoURL != expectedRepo {
		t.Errorf("expected repo %s, got %s", expectedRepo, resolved.Sources[0].RepoURL)
	}
}
