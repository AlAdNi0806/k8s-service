#!/bin/bash

output_file="${1:-go_files_output.txt}"

# Массив папок для пропуска (укажите нужные относительные пути без слеша в начале)
skip_dirs=("vendor" "third_party" "node_modules")

if ! git rev-parse --git-dir >/dev/null 2>&1; then
    echo "⚠️ Not in a Git repo. Using 'find' (ignores .gitignore)." >&2
    mapfile -d '' go_files < <(find . -name "*.go" -print0)
else
    # git ls-files uses NEWLINES, so use default mapfile (no -d)
    mapfile -t go_files < <(git ls-files --others --cached --exclude-standard -- '*.go')
fi

# Handle case where no files found
if [ ${#go_files[@]} -eq 0 ]; then
    echo "No non-ignored .go files found." >&2
    exit 1
fi

output=""
for file in "${go_files[@]}"; do
    skip_file=false
    for skip_dir in "${skip_dirs[@]}"; do
        if [[ "$file" == *"$skip_dir/"* ]]; then
            skip_file=true
            break
        fi
    done

    if $skip_file; then
        continue
    fi

    # git ls-files даёт пути вида "main.go", "databases/mysql.go"
    # Мы хотим: "/main.go", "/databases/mysql.go"
    output+="/${file}

\`\`\`go
$(cat "$file")
\`\`\`

"
done

printf '%s' "$output" > "$output_file"

echo "✅ Found ${#go_files[@]} non-ignored Go file(s), skipped folders: ${skip_dirs[*]}"
echo "✅ Output saved to $output_file"
