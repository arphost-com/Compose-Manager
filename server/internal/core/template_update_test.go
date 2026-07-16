package core

import "testing"

func TestMergeEnvKeepingValues_Emulatorjs(t *testing.T) {
	// A project deployed from the OLD emulatorjs template being updated to the
	// NEW one: the frontend key is added, the user's management port is kept,
	// and the now-unused key is preserved (not lost).
	existing := "EMULATORJS_PORT=3000\nEMULATORJS_MGMT_PORT=3001\nPUID=1000\nPGID=1000\nTZ=Etc/UTC\n"
	template := "EMULATORJS_FRONTEND_PORT=8082\nEMULATORJS_MGMT_PORT=3000\nPUID=1000\nPGID=1000\nTZ=Etc/UTC\n"

	merged := mergeEnvKeepingValues(existing, template)
	m := parseEnvMap(merged)

	if m["EMULATORJS_FRONTEND_PORT"] != "8082" {
		t.Errorf("new template key not added: got %q", m["EMULATORJS_FRONTEND_PORT"])
	}
	if m["EMULATORJS_MGMT_PORT"] != "3001" {
		t.Errorf("existing value not preserved for EMULATORJS_MGMT_PORT: got %q, want 3001", m["EMULATORJS_MGMT_PORT"])
	}
	if m["EMULATORJS_PORT"] != "3000" {
		t.Errorf("dropped key not preserved: EMULATORJS_PORT got %q, want 3000", m["EMULATORJS_PORT"])
	}
	if m["PUID"] != "1000" || m["TZ"] != "Etc/UTC" {
		t.Errorf("common keys mangled: PUID=%q TZ=%q", m["PUID"], m["TZ"])
	}
}

func TestMergeEnvKeepingValues_PreservesSecrets(t *testing.T) {
	// A generated secret in the project's .env must survive a template update
	// even though the template ships a change-me placeholder.
	existing := "JWT_SECRET=abc123realsecret\nDB_PASSWORD=hunter2\n"
	template := "JWT_SECRET=change-me\nDB_PASSWORD=change-me\nNEW_FLAG=on\n"
	m := parseEnvMap(mergeEnvKeepingValues(existing, template))
	if m["JWT_SECRET"] != "abc123realsecret" || m["DB_PASSWORD"] != "hunter2" {
		t.Errorf("secrets clobbered by template placeholders: %v", m)
	}
	if m["NEW_FLAG"] != "on" {
		t.Errorf("new template key not added: %q", m["NEW_FLAG"])
	}
}

func TestNewEnvKeys(t *testing.T) {
	got := newEnvKeys("A=1\nB=2\n", "A=x\nB=y\nC=z\n")
	if len(got) != 1 || got[0] != "C" {
		t.Errorf("newEnvKeys = %v, want [C]", got)
	}
}
