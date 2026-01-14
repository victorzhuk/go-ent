## Optimal Tooling

| Instead of    | Use               | Reason                          |
|---------------|-------------------|---------------------------------|
| `grep -rn`    | `rg -n`           | 10x faster, respects .gitignore |
| `grep "x"`    | `rg -tgo "x"`     | Go files only                   |
| `find -name`  | `fd`              | 5x faster                       |
| `cat \| grep` | `rg pattern file` | Direct search                   |

---