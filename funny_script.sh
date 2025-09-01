set -e

# 1) Make sure we’re on pkg/auth and save your local edits there
git checkout pkg/auth || git switch pkg/auth

# If you want a quick backup branch before touching anything:
git branch "backup/pkg-auth-$(date +%F-%H%M%S)"

# Commit your working changes on pkg/auth (preferred), or stash if you must.
git add -A
git commit -m "WIP: local changes before merging ratelimit into pkg/auth" || true

# 2) Wire up tracking for pkg/auth (creates remote branch if missing)
if git ls-remote --exit-code --heads origin pkg/auth >/dev/null 2>&1; then
  git branch --set-upstream-to=origin/pkg/auth
  git pull --ff-only
else
  git push -u origin pkg/auth
fi

# 3) Update pkg/ratelimit branch locally
git fetch origin
if git rev-parse --verify pkg/ratelimit >/dev/null 2>&1; then
  git checkout pkg/ratelimit || git switch pkg/ratelimit
else
  # If you don’t have it locally yet but it exists remotely:
  git checkout -b pkg/ratelimit origin/pkg/ratelimit || git switch -c pkg/ratelimit origin/pkg/ratelimit
fi

# Optional: backup this branch too
git branch "backup/pkg-ratelimit-$(date +%F-%H%M%S)"

# Ensure pkg/ratelimit tracks remote (or create it if missing)
if git ls-remote --exit-code --heads origin pkg/ratelimit >/dev/null 2>&1; then
  git branch --set-upstream-to=origin/pkg/ratelimit
  git pull --ff-only
else
  git push -u origin pkg/ratelimit
fi

# 4) Merge ratelimit → auth
git checkout pkg/auth || git switch pkg/auth
# Regular merge (keeps both histories)
git merge --no-ff pkg/ratelimit || true

echo
echo ">>> If you see conflicts, resolve them now. Typical hotspots: go.mod, Makefile, shared files."
echo "   Open each conflicted file, combine the changes, delete conflict markers,"
echo "   then: git add <files> && git commit"
echo
