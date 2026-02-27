package envconf

import (
	"strings"
	"testing"
	"time"
)

type customPort int

type customRatio float64

type customEnabled bool

type customTimeout int64

func TestLoadSupportedScalarTypes(t *testing.T) {
	resetSingletonForTest()

	t.Setenv("NAME", "app")
	t.Setenv("ENABLED", "true")
	t.Setenv("I8", "-8")
	t.Setenv("I16", "-16")
	t.Setenv("I32", "-32")
	t.Setenv("I64", "-64")
	t.Setenv("U8", "8")
	t.Setenv("U16", "16")
	t.Setenv("U32", "32")
	t.Setenv("U64", "64")
	t.Setenv("UP", "128")
	t.Setenv("F32", "1.25")
	t.Setenv("F64", "2.5")
	t.Setenv("TIMEOUT", "3s")

	type cfgType struct {
		Name    string        `env:"NAME"`
		Enabled bool          `env:"ENABLED"`
		I8      int8          `env:"I8"`
		I16     int16         `env:"I16"`
		I32     int32         `env:"I32"`
		I64     int64         `env:"I64"`
		U8      uint8         `env:"U8"`
		U16     uint16        `env:"U16"`
		U32     uint32        `env:"U32"`
		U64     uint64        `env:"U64"`
		UP      uintptr       `env:"UP"`
		F32     float32       `env:"F32"`
		F64     float64       `env:"F64"`
		Timeout time.Duration `env:"TIMEOUT"`
	}

	cfg, err := Load[cfgType](WithoutDotEnv())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Name != "app" || !cfg.Enabled || cfg.I8 != -8 || cfg.I16 != -16 || cfg.I32 != -32 || cfg.I64 != -64 {
		t.Fatalf("unexpected signed/scalar values: %+v", *cfg)
	}
	if cfg.U8 != 8 || cfg.U16 != 16 || cfg.U32 != 32 || cfg.U64 != 64 || cfg.UP != 128 {
		t.Fatalf("unexpected unsigned values: %+v", *cfg)
	}
	if cfg.F32 != float32(1.25) || cfg.F64 != 2.5 {
		t.Fatalf("unexpected float values: %+v", *cfg)
	}
	if cfg.Timeout != 3*time.Second {
		t.Fatalf("Timeout = %v, want 3s", cfg.Timeout)
	}
}

func TestLoadAliasTypes(t *testing.T) {
	resetSingletonForTest()

	t.Setenv("PORT", "8080")
	t.Setenv("RATIO", "0.75")
	t.Setenv("ENABLED", "1")
	t.Setenv("WAIT", "250")

	type cfgType struct {
		Port    customPort    `env:"PORT"`
		Ratio   customRatio   `env:"RATIO"`
		Enabled customEnabled `env:"ENABLED"`
		Wait    customTimeout `env:"WAIT"`
	}

	cfg, err := Load[cfgType](WithoutDotEnv())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != customPort(8080) || cfg.Ratio != customRatio(0.75) || !bool(cfg.Enabled) || cfg.Wait != customTimeout(250) {
		t.Fatalf("unexpected alias values: %+v", *cfg)
	}
}

func TestInvalidBoolValue(t *testing.T) {
	resetSingletonForTest()
	_, err := Load[struct {
		Enabled bool `env:"ENABLED" default:"maybe"`
	}](WithoutDotEnv())
	if err == nil || !strings.Contains(err.Error(), "failed parsing field Enabled") {
		t.Fatalf("error = %v, want bool parse error", err)
	}
}

func TestIntOverflowValue(t *testing.T) {
	resetSingletonForTest()
	_, err := Load[struct {
		N int8 `env:"N" default:"128"`
	}](WithoutDotEnv())
	if err == nil || !strings.Contains(err.Error(), "failed parsing field N") {
		t.Fatalf("error = %v, want int overflow parse error", err)
	}
}

func TestUnsignedNegativeValue(t *testing.T) {
	resetSingletonForTest()
	_, err := Load[struct {
		U uint `env:"U" default:"-1"`
	}](WithoutDotEnv())
	if err == nil || !strings.Contains(err.Error(), "failed parsing field U") {
		t.Fatalf("error = %v, want uint parse error", err)
	}
}

func TestInvalidFloatValue(t *testing.T) {
	resetSingletonForTest()
	_, err := Load[struct {
		R float64 `env:"R" default:"abc"`
	}](WithoutDotEnv())
	if err == nil || !strings.Contains(err.Error(), "failed parsing field R") {
		t.Fatalf("error = %v, want float parse error", err)
	}
}

func TestInvalidDurationValue(t *testing.T) {
	resetSingletonForTest()
	_, err := Load[struct {
		T time.Duration `env:"T" default:"10"`
	}](WithoutDotEnv())
	if err == nil || !strings.Contains(err.Error(), "failed parsing field T") {
		t.Fatalf("error = %v, want duration parse error", err)
	}
}

func TestPrecedenceForNonStringTypes(t *testing.T) {
	resetSingletonForTest()
	t.Setenv("PORT", "9000")

	cfg, err := Load[struct {
		Port int `env:"PORT" default:"7000"`
	}](WithDotEnvPath("does-not-exist.env"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != 9000 {
		t.Fatalf("Port = %d, want 9000", cfg.Port)
	}
}

func TestUnsupportedKindStillErrors(t *testing.T) {
	resetSingletonForTest()
	_, err := Load[struct {
		Tags []string `env:"TAGS" default:"a,b"`
	}](WithoutDotEnv())
	if err == nil || !strings.Contains(err.Error(), "unsupported kind") {
		t.Fatalf("error = %v, want unsupported kind", err)
	}
}
