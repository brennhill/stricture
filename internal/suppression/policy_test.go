package suppression

import "testing"

func TestDisableNextLineSpecificRule(t *testing.T) {
	src := []byte("// stricture-disable-next-line CONV-file-header\nx\n")
	p := Compile(src)

	if !p.Suppressed("CONV-file-header", 2) {
		t.Fatalf("expected next-line suppression for CONV-file-header")
	}
	if p.Suppressed("CONV-file-naming", 2) {
		t.Fatalf("unexpected suppression of other rule")
	}
}

func TestDisableEnableBlockSpecificRule(t *testing.T) {
	src := []byte(
		"// stricture-disable TQ-no-shallow-assertions\n" +
			"line1\n" +
			"// stricture-enable TQ-no-shallow-assertions\n" +
			"line2\n")
	p := Compile(src)

	if !p.Suppressed("TQ-no-shallow-assertions", 2) {
		t.Fatalf("expected suppression while block is disabled")
	}
	if p.Suppressed("TQ-no-shallow-assertions", 4) {
		t.Fatalf("unexpected suppression after enable")
	}
}

func TestDisableFileAllRules(t *testing.T) {
	src := []byte("// stricture-disable-file\nx\n")
	p := Compile(src)

	if !p.Suppressed("CONV-file-header", 2) {
		t.Fatalf("expected file-level suppression for all rules")
	}
	if !p.Suppressed("ARCH-import-boundary", 10) {
		t.Fatalf("expected file-level suppression for all rules")
	}
}

func TestDisableNextLineAllRules(t *testing.T) {
	src := []byte("// stricture-disable-next-line\nx\n")
	p := Compile(src)

	if !p.Suppressed("CONV-file-header", 2) {
		t.Fatalf("expected next-line all-rule suppression")
	}
	if !p.Suppressed("TQ-no-shallow-assertions", 2) {
		t.Fatalf("expected next-line all-rule suppression")
	}
}

func TestDisableFileSpecificRule(t *testing.T) {
	src := []byte("// stricture-disable-file CTR-manifest-conformance\nx\n")
	p := Compile(src)

	if !p.Suppressed("CTR-manifest-conformance", 2) {
		t.Fatalf("expected file-level suppression for rule")
	}
	if p.Suppressed("CTR-request-shape", 2) {
		t.Fatalf("unexpected suppression for unrelated rule")
	}
}
