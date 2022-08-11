package dag

import (
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/yohamta/dagu/internal/settings"
)

func TestLoadConfig(t *testing.T) {
	l := &Loader{}
	_, err := l.Load(testConfig, "")
	require.NoError(t, err)

	// without .yaml
	s := path.Join(testDir, "config_load")
	_, err = l.Load(s, "")
	require.NoError(t, err)
}

func TestLoadBaseConfig(t *testing.T) {
	l := &Loader{}
	cfg, err := l.loadBaseConfig(
		settings.MustGet(settings.SETTING__BASE_CONFIG),
		&BuildConfigOptions{},
	)
	require.NotNil(t, cfg)
	require.NoError(t, err)
}

func TestLoadBaseConfigError(t *testing.T) {
	for _, f := range []string{
		path.Join(testDir, "config_err_decode.yaml"),
		path.Join(testDir, "config_err_parse.yaml"),
	} {
		l := &Loader{}
		_, err := l.loadBaseConfig(f, &BuildConfigOptions{})
		require.Error(t, err)
	}
}

func TestLoadDeafult(t *testing.T) {
	l := &Loader{}
	cfg, err := l.Load(path.Join(testDir, "config_default.yaml"), "")
	require.NoError(t, err)

	require.Equal(t, time.Second*60, cfg.MaxCleanUpTime)
	require.Equal(t, 30, cfg.HistRetentionDays)
}

func TestLoadData(t *testing.T) {
	dat := `name: test DAG
steps:
  - name: "1"
    command: "true"
`
	l := &Loader{}
	ret, err := l.LoadData([]byte(dat))
	require.NoError(t, err)
	require.Equal(t, ret.Name, "test DAG")

	step := ret.Steps[0]
	require.Equal(t, step.Name, "1")
	require.Equal(t, step.Command, "true")

	// error
	dat = `invalidyaml`
	_, err = l.LoadData([]byte(dat))
	require.Error(t, err)

	// error
	dat = `invalidkey: test DAG`
	_, err = l.LoadData([]byte(dat))
	require.Error(t, err)

	// error
	dat = `name: test DAG`
	_, err = l.LoadData([]byte(dat))
	require.Error(t, err)
}

func TestLoadErrorFileNotExist(t *testing.T) {
	l := &Loader{}
	_, err := l.Load(path.Join(testDir, "not_existing_file.yaml"), "")
	require.Error(t, err)
}

func TestGlobalConfigNotExist(t *testing.T) {
	l := &Loader{}

	file := path.Join(testDir, "config_default.yaml")
	_, err := l.Load(file, "")
	require.NoError(t, err)
}

func TestDecodeError(t *testing.T) {
	l := &Loader{}
	file := path.Join(testDir, "config_err_decode.yaml")
	_, err := l.Load(file, "")
	require.Error(t, err)
}