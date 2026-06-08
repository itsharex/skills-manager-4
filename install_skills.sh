#!/bin/bash
set -e

AGENTS_DIR="$HOME/.agents/skills"
TRAE_DIR="$HOME/.trae-cn/skills"

echo "=== 安装 UI/UX Pro Max 核心技能（3个） ==="
for skill in ui-ux-pro-max ui-styling design-system; do
    src="/tmp/skills-uiux/.claude/skills/$skill"
    dst="$AGENTS_DIR/$skill"
    if [ -d "$src" ]; then
        if [ ! -d "$dst" ]; then
            cp -r "$src" "$dst" && echo "  ✓ $skill"
        else
            echo "  ⚠ $skill 已存在，跳过"
        fi
        if [ ! -L "$TRAE_DIR/$skill" ] && [ ! -d "$TRAE_DIR/$skill" ]; then
            ln -s "$dst" "$TRAE_DIR/$skill"
        fi
    fi
done

echo ""
echo "=== 安装 Patterns.dev 技能（JavaScript + React + Vue） ==="
count=0
for category in javascript react vue; do
    src_base="/tmp/skills-patterns/$category"
    if [ -d "$src_base" ]; then
        for skill_dir in "$src_base"/*/; do
            skill_name=$(basename "$skill_dir")
            [ "$skill_name" = "README.md" ] && continue
            [ ! -d "$skill_dir" ] && continue

            install_name="patterns-${category}-${skill_name}"
            dst="$AGENTS_DIR/$install_name"

            if [ ! -d "$dst" ]; then
                cp -r "$skill_dir" "$dst" && echo "  ✓ $install_name"
                count=$((count + 1))
            else
                echo "  ⚠ $install_name 已存在，跳过"
            fi

            if [ ! -L "$TRAE_DIR/$install_name" ] && [ ! -d "$TRAE_DIR/$install_name" ]; then
                ln -s "$dst" "$TRAE_DIR/$install_name"
            fi
        done
    fi
done

echo ""
echo "本次新安装: $count 个 Patterns.dev 技能"
echo ""
echo "=== 统计 ==="
echo "AGENTS 目录: $(ls -1 "$AGENTS_DIR" | wc -l | tr -d ' ') 个技能"
echo "TRAE 目录:   $(ls -1 "$TRAE_DIR" | wc -l | tr -d ' ') 个条目"
