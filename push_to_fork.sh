#!/usr/bin/env bash
# 一键完成 PentAGI fork 推送收尾。
#
# 前提（两项，都只能由你在浏览器里做）：
#   1) GitHub 上已创建 fork： https://github.com/vxcontrol/pentagi/fork
#      （目标是 limin1975/pentagi）
#   2) 已把本机公钥加进 GitHub 账号： https://github.com/settings/keys
#      公钥路径： ~/.ssh/id_ed25519_github.pub
#
# 本脚本做的事：
#   - 校验 SSH 通道（是否已授权）
#   - 校验 fork 是否存在
#   - 把本地 5 个提交推到 limin1975/pentagi 的 main
#   - 校验推送结果
#
# 用法： bash push_to_fork.sh

set -uo pipefail

REPO_DIR="/f/AI/pentagi"
ORIGIN="ssh://git@ssh.github.com:443/limin1975/pentagi.git"
FORK_API="https://api.github.com/repos/limin1975/pentagi"
PUBKEY_FILE="/c/Users/limin/.ssh/id_ed25519_github.pub"

cd "$REPO_DIR" || { echo "找不到仓库目录 $REPO_DIR"; exit 1; }

echo "=== 1/4 校验 SSH 通道 ==="
if timeout 30 git ls-remote origin >/dev/null 2>&1; then
  echo "  OK: SSH 已授权，origin 可访问"
else
  echo "  ✗ SSH 未授权。请先完成两件事："
  echo "    1) 创建 fork:  https://github.com/vxcontrol/pentagi/fork"
  echo "    2) 添加公钥:  https://github.com/settings/keys"
  echo ""
  echo "    把下面这一整行贴到 GitHub 的 SSH keys 页面："
  echo ""
  cat "$PUBKEY_FILE" 2>/dev/null | sed 's/^/      /'
  echo ""
  echo "    贴完重跑本脚本即可。"
  exit 1
fi

echo "=== 2/4 校验 fork 是否存在 ==="
fork_state=$(curl -s -m 25 --noproxy '*' "$FORK_API" 2>/dev/null \
  | python -c "import sys,json; d=json.load(sys.stdin); print(d.get('full_name',''))" 2>/dev/null)
if [ "$fork_state" = "limin1975/pentagi" ]; then
  echo "  OK: fork 已存在 -> $fork_state"
else
  echo "  ✗ 找不到 limin1975/pentagi。请先在浏览器创建 fork："
  echo "    https://github.com/vxcontrol/pentagi/fork"
  exit 1
fi

echo "=== 3/4 推送本地提交到 fork ==="
echo "  待推送提交："
git log --oneline origin/main..HEAD 2>/dev/null | sed 's/^/    /' \
  || git log --oneline upstream/main..HEAD | sed 's/^/    /'
echo ""
if timeout 180 git push origin main; then
  echo "  OK: 推送成功"
else
  echo "  ✗ 推送失败"
  exit 1
fi

echo "=== 4/4 校验远端状态 ==="
remote_head=$(git ls-remote origin refs/heads/main 2>/dev/null | awk '{print $1}')
local_head=$(git rev-parse HEAD)
echo "  本地 HEAD : $local_head"
echo "  远端 main : $remote_head"
if [ "$remote_head" = "$local_head" ]; then
  echo ""
  echo "✅ 完成。5 个提交已推送到 limin1975/pentagi"
  echo "   本地 remote 结构： origin=你的 fork(SSH/443)，upstream=上游(ghfast.top 镜像)"
else
  echo ""
  echo "⚠️  远端与本地不一致，请检查上面的推送输出"
  exit 1
fi
