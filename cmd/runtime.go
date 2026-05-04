package cmd

import (
	"os"
	"path/filepath"

	"github.com/rokuosan/al/internal/config"
	"github.com/rokuosan/al/internal/model"
	"github.com/rokuosan/al/internal/runner"
)

func staticProvider() config.Provider {
	return config.StaticProvider{
		Configs: []config.LoadedConfig{
			{
				Scope: model.ScopeGlobal,
				Path:  "static",
				Config: config.Config{
					Aliases: map[string]config.AliasConfig{
						"hello": {
							Run:         "printf 'hello\\n'",
							Description: "Temporary hard-coded alias for bring-up",
						},
					},
				},
			},
		},
	}
}

func buildEvalContext() (model.EvalContext, error) {
	wd, err := os.Getwd()
	if err != nil {
		return model.EvalContext{}, err
	}

	evalCtx := model.NewEvalContext(wd, wd, runner.DefaultShellName(), envMap())
	evalCtx.InGitRepo = detectGitRepo(wd)
	return evalCtx, nil
}

func detectGitRepo(startDir string) bool {
	dir := filepath.Clean(startDir)
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return false
		}
		dir = parent
	}
}

func envMap() map[string]string {
	env := make(map[string]string)
	for _, pair := range os.Environ() {
		for i := 0; i < len(pair); i++ {
			if pair[i] != '=' {
				continue
			}
			env[pair[:i]] = pair[i+1:]
			break
		}
	}
	return env
}
