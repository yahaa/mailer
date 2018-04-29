### mailer

![image](http://p1lpgmbe0.bkt.clouddn.com/mailer.png)

> 本来想在自己的服务器上搭一套邮件服务系统，奈何云服务器厂商已经把25端口上行带宽给封杀了，我能怎么办，mail 命令又是非常难用(不会用)
> 于是我脑子发热 用 go 实现的类似 linux/unix 系统下面的 mail 工具，我给它取名为 mailer



### 特性

* 目的是为了能够在 linux 服务器下用定时脚本发送邮件
* 支持 163 邮件。
* 支持 md 文件解析

#### 安装

> go get github.com/yahaa/mailer

#### 使用

* login

```bash
mailer login -e your@163.com -p 123456
```

* send email

> send markdown email
```bash
mailer send -s="这是邮件的主题" -to="to@example.com" -f=ttt.md
```

> send plain email

```bash
mailer send -s="这是普通邮件的主题" -to="to@example.com" -b="这是邮件的 body"

```



