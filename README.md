# webcrontab

`-_-!` Still not working with `sqlite3`, but working well with `mysql`. Wired.



Web based crontab.

![homepage](scripts/homepage.png)

## Build
```
# Install godep
go get -v github.com/tools/godep

# Build
git clone https://github.com/codeskyblue/webcrontab.git
cd webcrontab
godep go build

# Start
./webcrontab
```

## Contribute
Since the first ok version is released. This really takes me a lot of time, finally get it works. Very happy.

Pull request are welcomed, but make sure code have been tested.

## Thanks
* <https://github.com/robfig/cron>
* Icon from: <http://www.easyicon.net/516578-schedule_april_icon.html>

## LICENSE
[MIT](LICENSE)

## ERROR Still exists
ERROR Details
```
hzsunshx@onlinegame-13-180:~/goproj/src/webcrontab (master)$ ./webcrontab 
[xorm] [info]  2015/08/25 21:50:41.213278 [sql] SELECT name FROM sqlite_master WHERE type='table' and name = ? and ((sql like '%`id`%') or (sql like '%[id]%')) [record]
[xorm] [info]  2015/08/25 21:50:41.214781 [sql] SELECT name FROM sqlite_master WHERE type='table' and name = ? and ((sql like '%`name`%') or (sql like '%[name]%')) [record]
[xorm] [info]  2015/08/25 21:50:41.216150 [sql] SELECT name FROM sqlite_master WHERE type='table' and name = ? and ((sql like '%`index`%') or (sql like '%[index]%')) [record]
[xorm] [info]  2015/08/25 21:50:41.217409 [sql] SELECT name FROM sqlite_master WHERE type='table' and name = ? and ((sql like '%`trigger`%') or (sql like '%[trigger]%')) [record]
[xorm] [info]  2015/08/25 21:50:41.218638 [sql] SELECT name FROM sqlite_master WHERE type='table' and name = ? and ((sql like '%`exit_code`%') or (sql like '%[exit_code]%')) [record]
[xorm] [info]  2015/08/25 21:50:41.219874 [sql] SELECT name FROM sqlite_master WHERE type='table' and name = ? and ((sql like '%`created_at`%') or (sql like '%[created_at]%')) [record]
[xorm] [info]  2015/08/25 21:50:41.220983 [sql] SELECT name FROM sqlite_master WHERE type='table' and name = ? and ((sql like '%`duration`%') or (sql like '%[duration]%')) [record]
[xorm] [info]  2015/08/25 21:50:41.222136 [sql] SELECT name FROM sqlite_master WHERE type='table' and name = ? and ((sql like '%`task`%') or (sql like '%[task]%')) [record]
[xorm] [info]  2015/08/25 21:50:41.223250 [sql] SELECT sql FROM sqlite_master WHERE type='index' and tbl_name = ? [record]
2015/08/25 21:50:41 [INFO][webcrontab] web.go:229: [{hello * * * * * * echo hello world 你好 map[YY:123] true}]
2015/08/25 21:50:41 [INFO][webcrontab] web.go:234: Listening on *:4000
panic: not supported

goroutine 14 [running]:
github.com/go-xorm/xorm.buildConditions(0xc20809c510, 0xc2080580f0, 0x98ae00, 0xc2080fc9c0, 0x1000101, 0xc2080d5770, 0x0, 0x0, 0x0, 0x0, ...)
        /home/hzsunshx/goproj/src/webcrontab/Godeps/_workspace/src/github.com/go-xorm/xorm/statement.go:614 +0x274a
github.com/go-xorm/xorm.(*Statement).genCountSql(0xc2080feb58, 0x98ae00, 0xc2080fc9c0, 0x0, 0x0, 0x0, 0x0, 0x0)
        /home/hzsunshx/goproj/src/webcrontab/Godeps/_workspace/src/github.com/go-xorm/xorm/statement.go:1194 +0x138
github.com/go-xorm/xorm.(*Session).Count(0xc2080feb40, 0x98ae00, 0xc2080fc9c0, 0x0, 0x0, 0x0)
        /home/hzsunshx/goproj/src/webcrontab/Godeps/_workspace/src/github.com/go-xorm/xorm/session.go:1036 +0x123
main.(*Keeper).NewRecord(0xc208039630, 0xc20800bee0, 0x5, 0x0, 0x0, 0x0, 0x0, 0x0)
        /home/hzsunshx/goproj/src/webcrontab/keeper.go:71 +0x436
main.(*Task).Run(0xc2080396d0, 0xa33030, 0x8, 0x0, 0x0)
        /home/hzsunshx/goproj/src/webcrontab/cron.go:41 +0x6b
main.func·004()
        /home/hzsunshx/goproj/src/webcrontab/keeper.go:52 +0x4a
github.com/robfig/cron.FuncJob.Run(0xc20800bfa0)
        /home/hzsunshx/goproj/src/webcrontab/Godeps/_workspace/src/github.com/robfig/cron/cron.go:83 +0x20
created by github.com/robfig/cron.(*Cron).run
        /home/hzsunshx/goproj/src/webcrontab/Godeps/_workspace/src/github.com/robfig/cron/cron.go:159 +0x59f

goroutine 1 [IO wait]:
net.(*pollDesc).Wait(0xc2080aed10, 0x72, 0x0, 0x0)
        /home/hzsunshx/go/src/net/fd_poll_runtime.go:84 +0x47
net.(*pollDesc).WaitRead(0xc2080aed10, 0x0, 0x0)
        /home/hzsunshx/go/src/net/fd_poll_runtime.go:89 +0x43
net.(*netFD).accept(0xc2080aecb0, 0x0, 0x7f69b7df9db0, 0xc208100870)
        /home/hzsunshx/go/src/net/fd_unix.go:419 +0x40b
net.(*TCPListener).AcceptTCP(0xc2080362e8, 0x5f5634, 0x0, 0x0)
        /home/hzsunshx/go/src/net/tcpsock_posix.go:234 +0x4e
net/http.tcpKeepAliveListener.Accept(0xc2080362e8, 0x0, 0x0, 0x0, 0x0)
```
