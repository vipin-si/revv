## Description
Verify that every skill SKILL.md has valid YAML frontmatter with required `name` and `description` fields, and that every skill has an accompanying prompt.md file. Broken skills mean broken QA for all repos using revv.

## Priority
blocking

## Type
automated

## Commands
```bash
FAILED=0
for skill_dir in skills/*/; do
  skill_name=$(basename "$skill_dir")
  
  # Check SKILL.md exists
  if [ ! -f "$skill_dir/SKILL.md" ]; then
    echo "FAIL: $skill_dir missing SKILL.md"
    FAILED=1
    continue
  fi
  
  # Check prompt.md exists
  if [ ! -f "$skill_dir/prompt.md" ]; then
    echo "FAIL: $skill_dir missing prompt.md"
    FAILED=1
  fi
  
  # Check SKILL.md has frontmatter
  head -1 "$skill_dir/SKILL.md" | grep -q "^---" || {
    echo "FAIL: $skill_dir/SKILL.md missing YAML frontmatter"
    FAILED=1
  }
  
  # Check name field exists
  grep -q "^name:" "$skill_dir/SKILL.md" || {
    echo "FAIL: $skill_dir/SKILL.md missing 'name' field"
    FAILED=1
  }
  
  # Check description field exists
  grep -q "^description:" "$skill_dir/SKILL.md" || {
    echo "FAIL: $skill_dir/SKILL.md missing 'description' field"
    FAILED=1
  }
  
  echo "OK: $skill_name"
done

if [ $FAILED -ne 0 ]; then
  exit 1
fi
echo "PASS: all skills have valid SKILL.md + prompt.md"
```

## Expected Output
All 4 skills (init-repo, update-repo, add-tests, run-tests) report OK with valid frontmatter.
