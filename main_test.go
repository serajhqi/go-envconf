package envconf

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type testConfig struct {
	DBPass string `env:"DB_PASS" default:"postgres"`
}

func TestGetBeforeLoadErrors(t *testing.T) {
	resetSingletonForTest()

	_, err := Get[testConfig]()
	if err == nil {
		t.Fatal("Get() error = nil, want error")
	}
	if err != ErrNotLoaded {
		t.Fatalf("Get() error = %v, want ErrNotLoaded", err)
	}
}

func TestMustGetBeforeLoadPanics(t *testing.T) {
	resetSingletonForTest()

	defer func() {
		if recover() == nil {
			t.Fatal("MustGet() did not panic")
		}
	}()

	_ = MustGet[testConfig]()
}

func TestLoadThenGet(t *testing.T) {
	resetSingletonForTest()
	t.Setenv("DB_PASS", "from-env")

	loaded, err := Load[testConfig](WithoutDotEnv())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	got, err := Get[testConfig]()
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if loaded != got {
		t.Fatalf("Get() pointer mismatch")
	}
}

func TestMustLoadThenMustGet(t *testing.T) {
	resetSingletonForTest()
	t.Setenv("DB_PASS", "from-env")

	loaded := MustLoad[testConfig](WithoutDotEnv())
	got := MustGet[testConfig]()
	if loaded != got {
		t.Fatalf("MustGet() pointer mismatch")
	}
}

func TestLoadFromDotEnv(t *testing.T) {
	resetSingletonForTest()

	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("DB_PASS=from-dotenv\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	type cfgType struct {
		DBPass string `env:"DB_PASS"`
	}
	cfg, err := Load[cfgType](WithDotEnvPath(path))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DBPass != "from-dotenv" {
		t.Fatalf("DBPass = %q, want from-dotenv", cfg.DBPass)
	}
}

func TestLoadMissingRequired(t *testing.T) {
	resetSingletonForTest()

	type cfgType struct {
		DBPass string `env:"DB_PASS"`
	}
	_, err := Load[cfgType](WithoutDotEnv())
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "missing required value") {
		t.Fatalf("error = %v, want missing required value", err)
	}
}

func TestEnvOverridesDotEnv(t *testing.T) {
	resetSingletonForTest()
	t.Setenv("DB_PASS", "from-env")

	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("DB_PASS=from-dotenv\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	type cfgType struct {
		DBPass string `env:"DB_PASS" default:"fallback"`
	}
	cfg, err := Load[cfgType](WithDotEnvPath(path))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DBPass != "from-env" {
		t.Fatalf("DBPass = %q, want from-env", cfg.DBPass)
	}
}

func TestRepeatedCompatibleLoadNoOp(t *testing.T) {
	resetSingletonForTest()

	cfg1, err := Load[testConfig](WithoutDotEnv())
	if err != nil {
		t.Fatalf("first Load() error = %v", err)
	}
	cfg2, err := Load[testConfig](WithoutDotEnv())
	if err != nil {
		t.Fatalf("second Load() error = %v", err)
	}
	if cfg1 != cfg2 {
		t.Fatalf("Load() did not return singleton instance")
	}
}

func TestIncompatibleSecondLoadErrors(t *testing.T) {
	resetSingletonForTest()

	if _, err := Load[testConfig](WithoutDotEnv()); err != nil {
		t.Fatalf("first Load() error = %v", err)
	}

	type otherConfig struct {
		DBPass string `env:"DB_PASS" default:"x"`
	}
	if _, err := Load[otherConfig](WithoutDotEnv()); err == nil {
		t.Fatal("second incompatible type Load() error = nil, want error")
	}

	if _, err := Load[testConfig](WithDotEnvPath("custom.env")); err == nil {
		t.Fatal("second incompatible options Load() error = nil, want error")
	}
}

func TestConcurrentInitFirstWins(t *testing.T) {
	resetSingletonForTest()
	t.Setenv("DB_PASS", "value")

	type altConfig struct {
		DBPass string `env:"DB_PASS" default:"x"`
		Extra  string `env:"EXTRA" default:"y"`
	}

	var wg sync.WaitGroup
	wg.Add(2)

	errs := make(chan error, 2)
	go func() {
		defer wg.Done()
		_, err := Load[testConfig](WithoutDotEnv())
		errs <- err
	}()
	go func() {
		defer wg.Done()
		_, err := Load[altConfig](WithoutDotEnv())
		errs <- err
	}()

	wg.Wait()
	close(errs)

	var successCount int
	var errorCount int
	for err := range errs {
		if err == nil {
			successCount++
		} else if strings.Contains(err.Error(), "already initialized") {
			errorCount++
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if successCount != 1 || errorCount != 1 {
		t.Fatalf("success=%d error=%d, want 1/1", successCount, errorCount)
	}
}
