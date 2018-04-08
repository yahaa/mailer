package main

import (
    "flag"
    "os"
    "fmt"
    "os/user"
    "io/ioutil"
    "encoding/json"
    "strings"
    "net/smtp"
)

var (
    MSG = `Usage:	mailer COMMAND

Commands:
login               Login to mailer
send                Send email

Run 'mailer COMMAND --help' for more information on a command.
        `

    PATH     = "/.mailer/"
    FILENAME = "mailer.json"
    LogInput LoginInput
    ToInput  SendToInput
    TYPE     = map[string]string{
        "html":  "Content-Type: text/html; charset=UTF-8",
        "plain": "Content-Type: text/plain; charset=UTF-8",
    }
)

type LoginInput struct {
    Protocol *string
    Host     *string
    Port     *int
    Email    *string
    Pass     *string
}

type SendToInput struct {
    To      *string
    From    string
    Subject *string
    Type    *string
    Body    *string
}

func (to *SendToInput) Data() []byte {
    msg := []byte("To: " + *to.To +
        "\r\nFrom: " +
        to.From +
        ">\r\nSubject: " +
        *to.Subject +
        "\r\n" +
        *to.Type + "\r\n\r\n" +
        *to.Body)
    return msg

}

func (to *SendToInput) Invalid() bool {
    if *to.To == "" || *to.Subject == "" || *to.Type == "" || *to.Body == "" {
        return true
    }
    return false
}

func (l *LoginInput) Invalid() bool {
    if *l.Protocol == "" || *l.Host == "" || *l.Port == 0 || *l.Email == "" || *l.Pass == "" {
        return true
    }
    return false
}

func SendEmail(lg *LoginInput, to *SendToInput) error {
    auth := smtp.PlainAuth("", *lg.Email, *lg.Pass, *lg.Host)
    sendTo := strings.Split(*to.To, ";")
    err := smtp.SendMail(*lg.Host+":"+string(*lg.Port), auth, to.From, sendTo, to.Data())
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

    // LoginInput
    LogInput.Protocol = login.String("protocol", "", "The protocol which your email use, support smtp")
    LogInput.Host = login.String("host", "", "The protocol host such as smtp.gmail.com or smtp.163.com.")
    LogInput.Port = login.Int("port", 25, "The port or mail server host.")
    LogInput.Email = login.String("email", "", "Email such as example@gmail.com ")
    LogInput.Pass = login.String("pass", "", "The password of your email")

    // SendToInput
    ToInput.To = send.String("to", "", "Which emails want to send more email split by ';' such as a@qq.com;b@qq.com")
    ToInput.Type = send.String("type", "plain", "Which email context type want to send")
    ToInput.Subject = send.String("subject", "", "Email Subject")
    ToInput.Body = send.String("body", "", "Email body")

    if len(os.Args) < 2 {
        fmt.Println(MSG)
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
        fmt.Println("Not such permition")
        os.Exit(1)
    }
    configDir := curUser.HomeDir + PATH
    configName := configDir + FILENAME

    if login.Parsed() {

        if LogInput.Invalid() {
            login.PrintDefaults()
            os.Exit(1)
        }

        curUser, err := user.Current()
        if err != nil {
            fmt.Println("Not such permition")
            os.Exit(1)
        }
        configDir := curUser.HomeDir + PATH

        err = os.MkdirAll(configDir, os.ModePerm)
        if err != nil {
            fmt.Printf("Error %v", err)
            os.Exit(1)
        }

        data, err := json.Marshal(LogInput)
        if err != nil {
            fmt.Println("Process error")
            os.Exit(1)
        }

        err = ioutil.WriteFile(configName, data, os.ModePerm)
        if err != nil {
            fmt.Printf("Error %v", err)
            os.Exit(1)
        }

        fmt.Println("Login success")
    }

    if send.Parsed() {
        if ToInput.Invalid() {
            login.PrintDefaults()
            os.Exit(1)
        }

        var u LoginInput
        data, err := ReadFile(configName)
        if err != nil {
            fmt.Printf("Error %v", err)
            os.Exit(1)
        }
        err = json.Unmarshal(data, &u)
        if err != nil {
            fmt.Printf("Error %v", err)
            os.Exit(1)
        }

        ToInput.From = *u.Email

        SendEmail(&u, &ToInput)
        fmt.Println("Send email success").

    }

}
