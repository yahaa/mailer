package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "io/ioutil"
    "net/smtp"
    "os"
    "os/user"
    "strings"
    "github.com/goinbox/color"
    "github.com/pkg/errors"
    md "gopkg.in/russross/blackfriday.v2"
)

var (
    USAGE = `Usage:	mailer COMMAND

Commands:
login               Login to mailer
send                Send email

Run 'mailer COMMAND --help' for more information on a command.
        `

    PATH       = "/.mailer/"
    FILENAME   = "mailer.json"
    LoginInput Login
    SendInput  Send
    HTML       = "Content-Type: text/html; charset=UTF-8"
    PLAIN      = "Content-Type: text/plain; charset=UTF-8"

    SupportHost = map[string]Config{
        "@163.com":   {Protocol: "smtp", Port: 25, Host: "smtp.163.com"},
        //"@qq.com":    {Protocol: "smtp", Port: 25, Host: "smtp.qq.com"},
        //"@gmail.com": {Protocol: "smtp", Port: 587, Host: "smtp.gmail.com"},
    }

    NotSupportError = errors.New("this type email we not support now")
)

type Login struct {
    Email *string
    Pass  *string
}

type Send struct {
    To      *string
    Subject *string
    Type    *string
    Body    *string
    F       *string
}

type Config struct {
    Protocol string
    Port     int
    Host     string
}

type Email struct {
    From       string
    FPass      string
    To         string
    Subject    string
    EmailType  string
    Body       []byte
    HostConfig *Config
}

func (e *Email) Byte() []byte {
    msg := []byte("To: " + e.To +
        "\r\nFrom: <" +
        e.From +
        ">\r\nSubject: " +
        e.Subject +
        "\r\n" +
        e.EmailType + "\r\n\r\n" +
        string(e.Body))
    return msg
}

func (mi Send) Invalid() bool {

    if *mi.To == "" || *mi.Subject == "" {
        return true
    }
    return false
}

func (l Login) Invalid() bool {
    if *l.Email == "" || *l.Pass == "" {
        return true
    }
    return false
}

func getEmailType(f *string) string {
    if *f == "" {
        return PLAIN
    }

    return HTML
}

func getConfig(email string) *Config {
    typs := strings.Split(email, "@")
    if len(typs) == 2 {
        t := "@" + typs[1]
        if v, ok := SupportHost[t]; ok {
            return &v
        }
        return nil
    }
    return nil
}

func md2Html(input []byte) []byte {

    res := make([]byte, 0)
    header := `
	<!doctype html>
	<html lang="en">
	<head>
		<meta charset="utf-8"/>
		<style>
			*{margin:0;padding:0;}
			body {
				 font:13.34px helvetica,arial,freesans,clean,sans-serif;
				 color:black;
				 line-height:1.4em;
				 background-color: #F8F8F8;
				 padding: 0.7em;
			}
			p {
				 margin:1em 0;
				 line-height:1.5em;
			}
			table {
				 font-size:inherit;
				 font:100%;
				 margin:1em;
			}
			table th{border-bottom:1px solid #bbb;padding:.2em 1em;}
			table td{border-bottom:1px solid #ddd;padding:.2em 1em;}
			input[type=text],input[type=password],input[type=image],textarea{font:99% helvetica,arial,freesans,sans-serif;}
			select,option{padding:0 .25em;}
			optgroup{margin-top:.5em;}
			pre,code{font:12px Menlo, Monaco, "DejaVu Sans Mono", "Bitstream Vera Sans Mono",monospace;}
			pre {
				 margin:1em 0;
				 font-size:12px;
				 background-color:#eee;
				 border:1px solid #ddd;
				 padding:5px;
				 line-height:1.5em;
				 color:#444;
				 overflow:auto;
				 -webkit-box-shadow:rgba(0,0,0,0.07) 0 1px 2px inset;
				 -webkit-border-radius:3px;
				 -moz-border-radius:3px;border-radius:3px;
			}
			pre code {
				 padding:0;
				 font-size:12px;
				 background-color:#eee;
				 border:none;
			}
			code {
				 font-size:12px;
				 background-color:#f8f8ff;
				 color:#444;
				 padding:0 .2em;
				 border:1px solid #dedede;
			}
			img{border:0;max-width:100%;}
			abbr{border-bottom:none;}
			a{color:#4183c4;text-decoration:none;}
			a:hover{text-decoration:underline;}
			a code,a:link code,a:visited code{color:#4183c4;}
			h2,h3{margin:1em 0;}
			h1,h2,h3,h4,h5,h6{border:0;}
			h1{font-size:170%;border-top:4px solid #aaa;padding-top:.5em;margin-top:1.5em;}
			h1:first-child{margin-top:0;padding-top:.25em;border-top:none;}
			h2{font-size:150%;margin-top:1.5em;border-top:4px solid #e0e0e0;padding-top:.5em;}
			h3{margin-top:1em;}
			hr{border:1px solid #ddd;}
			ul{margin:1em 0 1em 2em;}
			ol{margin:1em 0 1em 2em;}
			ul li,ol li{margin-top:.5em;margin-bottom:.5em;}
			ul ul,ul ol,ol ol,ol ul{margin-top:0;margin-bottom:0;}
			blockquote{margin:1em 0;border-left:5px solid #ddd;padding-left:.6em;color:#555;}
			dt{font-weight:bold;margin-left:1em;}
			dd{margin-left:2em;margin-bottom:1em;}
			sup {
			   font-size: 0.83em;
			   vertical-align: super;
			   line-height: 0;
			}
			* {
				 -webkit-print-color-adjust: exact;
			}
			@media screen and (min-width: 914px) {
			   body {
			      width: 854px;
			      margin:0 auto;
			   }
			}
			@media print {
				 table, pre {
					  page-break-inside: avoid;
				 }
				 pre {
					  word-wrap: break-word;
				 }
			}
		</style>
	</head>
	`
    footer := "</html>\n"
    res = append(res, []byte(header)...)
    res = append(res, md.Run(input)...)
    res = append(res, []byte(footer)...)
    return res
}

func getBody(f, body *string) ([]byte, error) {
    if *f == "" {
        return []byte(*body), nil
    }
    b, err := ReadFile(*f)
    if err != nil {
        return nil, err
    }

    out := md2Html(b)
    return out, err
}

func SendEmail(e *Email) error {

    host := fmt.Sprintf("%s:%d", e.HostConfig.Host, e.HostConfig.Port)
    auth := smtp.PlainAuth("", e.From, e.FPass, e.HostConfig.Host)
    sendTo := strings.Split(e.To, ";")
    err := smtp.SendMail(host, auth, e.From, sendTo, e.Byte())
    return err

}

// ReadFile 读取配置文件
func ReadFile(filename string) ([]byte, error) {
    data, err := ioutil.ReadFile(filename)
    return data, err
}

func main() {
    // command
    login := flag.NewFlagSet("login", flag.ExitOnError)
    send := flag.NewFlagSet("send", flag.ExitOnError)

    // Login
    LoginInput.Email = login.String("e", "", "Email such as example@gmail.com ")
    LoginInput.Pass = login.String("p", "", "The password of your email")

    // SendInput
    SendInput.To = send.String("to", "", "Which emails want to send more email split by ';' such as a@qq.com;b@qq.com")
    SendInput.Subject = send.String("s", "", "Email Subject")
    SendInput.Body = send.String("b", "", "Email body")
    SendInput.F = send.String("f", "", "markdown file to send")

    if len(os.Args) < 2 {
        fmt.Println(string(color.Green([]byte(USAGE))))
        os.Exit(1)
    }

    switch os.Args[1] {
    case "login":
        login.Parse(os.Args[2:])
    case "send":
        send.Parse(os.Args[2:])
    default:
        flag.PrintDefaults()
        os.Exit(1)
    }

    curUser, err := user.Current()
    if err != nil {
        fmt.Println(string(color.Red([]byte("Not such permition"))))
        os.Exit(1)
    }
    configDir := curUser.HomeDir + PATH
    configName := configDir + FILENAME

    if login.Parsed() {

        if LoginInput.Invalid() {
            login.PrintDefaults()
            os.Exit(1)
        }

        curUser, err := user.Current()
        if err != nil {
            fmt.Println(string(color.Red([]byte("Not such permition"))))
            os.Exit(1)
        }
        configDir := curUser.HomeDir + PATH

        err = os.MkdirAll(configDir, os.ModePerm)
        if err != nil {
            fmt.Println(string(color.Red([]byte(err.Error()))))
            os.Exit(1)
        }

        data, err := json.Marshal(LoginInput)
        if err != nil {
            fmt.Println(string(color.Red([]byte(err.Error()))))
            os.Exit(1)
        }

        err = ioutil.WriteFile(configName, data, os.ModePerm)
        if err != nil {
            fmt.Println(string(color.Red([]byte(err.Error()))))
            os.Exit(1)
        }

        fmt.Println(string(color.Green([]byte("Login success."))))
    }

    if send.Parsed() {
        if SendInput.Invalid() {
            send.PrintDefaults()
            os.Exit(1)
        }

        var u Login
        data, err := ReadFile(configName)
        if err != nil {
            fmt.Println(string(color.Red([]byte("please login"))))
            os.Exit(1)
        }
        err = json.Unmarshal(data, &u)
        if err != nil {
            fmt.Println(string(color.Red([]byte("please login"))))
            os.Exit(1)
        }

        b, err := getBody(SendInput.F, SendInput.Body)
        if err != nil {
            fmt.Println(string(color.Red([]byte(err.Error()))))
            os.Exit(1)
        }

        email := Email{
            From:       *u.Email,
            FPass:      *u.Pass,
            To:         *SendInput.To,
            Subject:    *SendInput.Subject,
            EmailType:  getEmailType(SendInput.F),
            Body:       b,
            HostConfig: getConfig(*u.Email),
        }

        err = SendEmail(&email)
        if err != nil {
            fmt.Println(string(color.Red([]byte(err.Error()))))
            os.Exit(1)
        }
        fmt.Println(string(color.Green([]byte("Send email success"))))

    }

}
