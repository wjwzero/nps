# 部署说明
## NPS
1. 打包 `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -extldflags -static -extldflags -static" ./cmd/nps/nps.go`
2. 配置及静态文件
   - `~/GolandProjects/nps/web/views` 
   - `~/GolandProjects/nps/web/static`
   - `~/GolandProjects/nps/conf`
   - `ln -s ~/GolandProjects/nps_file/web /tmp/GoLand/`
3. 开放端口
   - 对外开放 tcp：
     - 8024
   - 对外开放 udp：
     - 6000~6002 
   - 对大厦开放:
     - 8080
4. 启动脚本   生产环境不使用命令直接启动
```
./nps
```
5. nps.conf 配置，生成环境使用prod文件夹下配置 并注意以下文本配置
```
#p2p 本机外网IP地址
p2p_ip=
# 若 p2p_port 设置为6000，请在防火墙开放6000~6002(额外添加2个端口)udp端口
p2p_port=6000
#p2p代理监听的端口 tcp
p2p_listen_port=5212

# 云平台地址 
cloudAddr=
# 最大线程数
max_threads=40000
```
6. systemctl 管理NPS服务
cp 到/usr/lib/systemd/system/
  - 开机启动
    - systemctl enable nps

配置nps.service文件 注意以下文本配置 ExecStart 替换实际路径
```
[Unit]
Description=nps
Documentation=https://docs.cloudreve.org
Wants=network.target

[Service]
WorkingDirectory=/root/nps
ExecStart=/root/nps/nps service
Restart=on-failure
RestartSec=5s
KillMode=mixed

StandardOutput=null
StandardError=syslog

[Install]
WantedBy=multi-user.target
```


## NPC
1. 打包 `CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o npc_arm64 -ldflags "-s -w -extldflags -static -extldflags -static" ./npc.go`
2. 启动脚本
```
   nohup ./npc_arm64_cgl -cloudAddr=http://xx -vkey=chenguolong -type=tcp > client.log 2>&1 &
```
注: 
-server=** //可以直接指定nps节点
-cloudAddr=http://xx //云平台地址，注：Android盒子中 无/etc/resolv.conf 会无法根据域名获取IP 及DNS失效

## NPC SDK
1. 打包 `gomobile bind -target=android/arm,android/arm64 -o npc_arm_221115.aar`
----
ln -s ~/GolandProjects/nps_file/conf /tmp/GoLand/  
ln -s ~/GolandProjects/nps_file/web /tmp/GoLand/
----

# 版本记录
## NPS version

## NPC version

## NPC SDK version
1. npc_arm_221115.aar;添加port参数，由调用者传入 `StartP2PClient("http://xx", "chenguolong76ff", "123123", 52000)`
2. npc_arm_221207.arr;app端连接时password加入签名验证

## 注  传输过程中使用加密数据 ，保存使用源数据
1. cmd/npc/sdk/sdk.go 编译打包后为  app端使用的[SDK](#NPC SDK)
    + 入参 verifyKey 与 password 需签名加密  （加密方式为 加延 32-bit MD5 前4位 参考 crypt文件）
2. cmd/npc/npc.go 编译打包后为 设备端使用的[SDK](#NPC)
    + 入参 verifyKey 不需加密
