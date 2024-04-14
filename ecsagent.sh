#!/usr/bin/env bash

_red() { echo -e "\033[31m\033[01m$@\033[0m"; }
_green() { echo -e "\033[32m\033[01m$@\033[0m"; }
_yellow() { echo -e "\033[33m\033[01m$@\033[0m"; }
_blue() { echo -e "\033[36m\033[01m$@\033[0m"; }
reading() { read -rp "$(_green "$1")" "$2"; }
cd /root >/dev/null 2>&1
if [ ! -d /usr/local/bin ]; then
    mkdir -p /usr/local/bin
fi
reading "主控Token：" token
reading "主控IPV4/域名：" host
reading "主控API端口：" port
rm -rf /usr/local/bin/ecsagent
curl -o /usr/local/bin "https://raw.githubusercontent.com/spiritysdx/ecs_agent/main/ecsagent"
curl -o /etc/systemd/system/ "https://raw.githubusercontent.com/spiritysdx/ecs_agent/main/ecsagent.service"
chmod +x /usr/local/bin/ecsagent
chmod +x /etc/systemd/system/ecsagent.service
if [ -f "/usr/local/bin/ecsagent.service" ]; then
    new_exec_start="ExecStart=/usr/local/bin/ecsagent -token ${token} -host ${host} -port ${port}"
    file_path="/etc/systemd/system/ecsagent.service"
    line_number=6
    sed -i "${line_number}s|.*|${new_exec_start}|" "$file_path"
fi
systemctl daemon-reload
systemctl enable ecsagent.service
systemctl start ecsagent.service