package adapter

import "hop.top/kit/toolspec"

// BuiltinWorkflows returns a toolspec.Source that provides
// common workflow patterns for well-known tools. This bridges
// the gap until curated adapters or tldr seeding exist.
func BuiltinWorkflows() toolspec.Source {
	specs := map[string]*toolspec.ToolSpec{
		"git": {
			Name: "git",
			Workflows: []toolspec.Workflow{
				{
					Name:  "feature-branch",
					Steps: []string{"git status", "git add", "git commit", "git push"},
					After: map[string][]string{
						"git status":   {"git add", "git diff", "git commit"},
						"git add":      {"git status", "git commit", "git diff --staged"},
						"git commit":   {"git push", "git log --oneline -5", "git status"},
						"git push":     {"gh pr create", "git log --oneline -5", "git branch -a"},
						"git diff":     {"git add", "git status", "git checkout"},
						"git log":      {"git diff", "git show", "git status"},
						"git checkout":  {"git status", "git pull", "git branch"},
						"git merge":    {"git branch -d", "git push", "git log --oneline -5"},
						"git branch":   {"git checkout", "git push", "git status"},
						"git pull":     {"git status", "git log --oneline -5", "git diff"},
						"git stash":    {"git status", "git stash pop", "git checkout"},
						"git rebase":   {"git status", "git push --force-with-lease", "git log --oneline -5"},
						"git reset":    {"git status", "git diff", "git add"},
						"git restore":  {"git status", "git add", "git diff"},
						"git fetch":    {"git status", "git merge", "git rebase"},
						"git show":     {"git log", "git diff", "git status"},
						"git tag":      {"git push --tags", "git log --oneline -5"},
						"git switch":   {"git status", "git pull"},
						"git init":     {"git add", "git commit -m 'init'"},
						"git clone":    {"cd", "git status"},
					},
				},
			},
		},
		"docker": {
			Name: "docker",
			Workflows: []toolspec.Workflow{
				{
					Name:  "container-lifecycle",
					Steps: []string{"docker build", "docker run", "docker ps", "docker logs", "docker stop"},
					After: map[string][]string{
						"docker build":   {"docker run", "docker images", "docker push"},
						"docker run":     {"docker ps", "docker logs", "docker exec"},
						"docker ps":      {"docker logs", "docker exec", "docker stop"},
						"docker logs":    {"docker exec", "docker stop", "docker restart"},
						"docker stop":    {"docker rm", "docker ps", "docker start"},
						"docker exec":    {"docker logs", "docker ps", "docker stop"},
						"docker images":  {"docker run", "docker rmi", "docker build"},
						"docker pull":    {"docker run", "docker images"},
						"docker push":    {"docker images", "docker tag"},
						"docker rm":      {"docker ps", "docker images"},
						"docker compose": {"docker compose ps", "docker compose logs"},
						"docker network": {"docker network ls", "docker ps"},
						"docker volume":  {"docker volume ls", "docker volume rm"},
					},
				},
			},
		},
		"gh": {
			Name: "gh",
			Workflows: []toolspec.Workflow{
				{
					Name:  "pr-workflow",
					Steps: []string{"gh pr create", "gh pr view", "gh pr merge"},
					After: map[string][]string{
						"gh pr":       {"gh pr view", "gh pr list", "gh pr create"},
						"gh pr create": {"gh pr view", "gh run list"},
						"gh pr view":  {"gh pr merge", "gh pr checkout", "gh pr comment"},
						"gh pr merge": {"git pull", "git branch -d"},
						"gh pr checkout": {"git status", "git diff"},
						"gh issue":    {"gh issue view", "gh issue list", "gh issue create"},
						"gh issue view": {"gh issue edit", "gh issue comment"},
						"gh issue edit": {"git status", "gh issue view"},
						"gh run":      {"gh run view", "gh run list"},
						"gh run list": {"gh run view", "gh run rerun"},
						"gh repo":     {"gh repo view", "gh repo clone"},
					},
				},
			},
		},
		"wrangler": {
			Name: "wrangler",
			Workflows: []toolspec.Workflow{
				{
					Name:  "deploy-cycle",
					Steps: []string{"wrangler dev", "wrangler deploy", "wrangler tail"},
					After: map[string][]string{
						"wrangler dev":    {"wrangler deploy", "wrangler types"},
						"wrangler deploy": {"wrangler tail", "wrangler versions"},
						"wrangler tail":   {"wrangler deploy", "wrangler dev"},
						"wrangler login":  {"wrangler whoami", "wrangler dev"},
						"wrangler whoami": {"wrangler dev", "wrangler deploy"},
						"wrangler d1":     {"wrangler d1 execute", "wrangler d1 list"},
						"wrangler kv":     {"wrangler kv list", "wrangler kv put"},
						"wrangler r2":     {"wrangler r2 object", "wrangler r2 bucket"},
						"wrangler pages":  {"wrangler pages deploy", "wrangler pages project"},
					},
				},
			},
		},
	}

	return toolspec.SourceFunc(func(tool string) (*toolspec.ToolSpec, error) {
		if spec, ok := specs[tool]; ok {
			return spec, nil
		}
		return nil, nil
	})
}
