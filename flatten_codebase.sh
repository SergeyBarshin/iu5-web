#!/bin/bash
# Собирает ключевые файлы Go-проекта в единый текстовый файл с путями.

ROOT_DIR="${1:-.}"                                # Корень проекта
OUTPUT_FILE="${2:-go_project_flattened.txt}"       # Имя итогового файла

# Включаем только ключевые типы файлов
INCLUDE_EXTENSIONS=("go" "mod" "sum" "yaml" "yml" "toml" "json" "md")

# Исключаем стандартные неключевые директории
EXCLUDE_DIRS=(.git .idea .vscode bin build dist vendor __pycache__ .cache)

# Очистим старый файл
> "$OUTPUT_FILE"

# Проверка: исключать путь или нет
should_exclude() {
  local path="$1"
  for dir in "${EXCLUDE_DIRS[@]}"; do
    if [[ "$path" == *"/$dir"* ]]; then
      return 0
    fi
  done
  return 1
}

echo "📦 Собираем ключевые файлы из $ROOT_DIR ..."

# Основной цикл по всем файлам
while IFS= read -r -d '' file; do
  should_exclude "$file" && continue

  # Проверяем по расширениям
  match=false
  for ext in "${INCLUDE_EXTENSIONS[@]}"; do
    if [[ "$file" == *.$ext ]]; then
      match=true
      break
    fi
  done
  $match || continue

  echo -e "\n\n### FILE: ${file#$ROOT_DIR/}" >> "$OUTPUT_FILE"
  echo "================================================================================" >> "$OUTPUT_FILE"
  cat "$file" >> "$OUTPUT_FILE" 2>/dev/null
  echo -e "\n================================================================================" >> "$OUTPUT_FILE"

done < <(find "$ROOT_DIR" -type f -print0)

# Итоговая статистика
FILES_COUNT=$(grep -c "^### FILE:" "$OUTPUT_FILE" || echo 0)
echo -e "\n\n📊 ИТОГО:
