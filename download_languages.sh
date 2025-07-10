#!/bin/bash

# Create languages directory if it doesn't exist
mkdir -p web/static/hj/languages

# Base URL for highlight.js CDN
BASE_URL="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.11.1/languages"

# List of languages to download
languages=(
    "bash"
    "c"
    "cpp"
    "csharp"
    "css"
    "dart"
    "diff"
    "dockerfile"
    "elixir"
    "go"
    "graphql"
    "haskell"
    "java"
    "javascript"
    "json"
    "kotlin"
    "lua"
    "makefile"
    "markdown"
    "nginx"
    "objectivec"
    "perl"
    "php"
    "python"
    "r"
    "ruby"
    "rust"
    "scss"
    "shell"
    "sql"
    "swift"
    "typescript"
    "vim"
    "xml"
    "yaml"
)

echo "Downloading highlight.js language files..."

for lang in "${languages[@]}"; do
    echo "Downloading ${lang}.min.js..."
    curl -s -o "web/static/hj/languages/${lang}.min.js" "${BASE_URL}/${lang}.min.js"
    if [ $? -eq 0 ]; then
        echo "✓ Downloaded ${lang}.min.js"
    else
        echo "✗ Failed to download ${lang}.min.js"
    fi
done

echo "All language files downloaded!"
