package smtp

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/smtp"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cafebazaar/bahram/datasource"
	"github.com/cafebazaar/blacksmith/logging"
	"github.com/garyburd/redigo/redis"
	"github.com/sloonz/go-iconv"
	"github.com/sloonz/go-qprintable"
)

const (
	debugTag = "SMTP"
)

type redisClient struct {
	count int
	conn  redis.Conn
	time  int
}

type Client struct {
	state       int
	helo        string
	mail_from   string
	rcpt_to     string
	read_buffer string
	response    string
	address     string
	data        string
	subject     string
	hash        string
	username    string
	password    string
	time        int64
	tls_on      bool
	auth        bool
	conn        net.Conn
	bufin       *bufio.Reader
	bufout      *bufio.Writer
	kill_time   int64
	errors      int
	clientId    int64
	savedNotify chan int
}

type ClientMessage struct {
	From     string
	To       string
	Data     string
	Subject  string
	Username string
	Auth     bool
}

var gConfig = map[string]string{
	"GSMTP_MAX_SIZE":       "131072",
	"GSMTP_HOST_NAME":      "server.example.com", // This should also be set to reflect your RDNS
	"GSMTP_PUB_KEY":        "/etc/ssl/certs/ssl-cert-snakeoil.pem",
	"GSMTP_PRV_KEY":        "/etc/ssl/private/ssl-cert-snakeoil.key",
	"GM_ALLOWED_HOSTS":     "cafebazaar.ir",
	"GM_PRIMARY_MAIL_HOST": "cafebazaar.ir",
}

func logln(level int, s string) {
	if level == 2 {
		log.Fatalf(s)
	} else if level == 1 {
		logging.Log(debugTag, s)
	} else {
		logging.Debug(debugTag, s)
	}
}

var sem chan int              // currently active clients
var SaveMailChan chan *Client // workers for saving mail
var TLSconfig *tls.Config
var max_size int // max email DATA size
var timeout time.Duration
var allowedHosts = make(map[string]bool, 15)

func initVar() {
	sem = make(chan int, 50)
	SaveMailChan = make(chan *Client, 5)
	timeout = time.Duration(10)
	max_size = 131072
	cert, err := tls.LoadX509KeyPair(gConfig["GSMTP_PUB_KEY"], gConfig["GSMTP_PRV_KEY"])

	if err != nil {
		logln(1, fmt.Sprintf("There was a problem with loading the certificate: %s", err))
	}

	TLSconfig = &tls.Config{Certificates: []tls.Certificate{cert}, ClientAuth: tls.VerifyClientCertIfGiven, ServerName: gConfig["GSMTP_HOST_NAME"]}
	TLSconfig.Rand = rand.Reader
	// map the allow hosts for easy lookup
	if arr := strings.Split(gConfig["GM_ALLOWED_HOSTS"], ","); len(arr) > 0 {
		for i := 0; i < len(arr); i++ {
			allowedHosts[arr[i]] = true
		}
	}
}

func Serve(listenAddr net.TCPAddr, datasource *datasource.DataSource) error {
	initVar()

	addr := listenAddr.String()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logln(2, fmt.Sprintf("Cannot listen on port, %v\n", err))
	} else {
		logln(1, fmt.Sprintf("Listening on tcp %s\n", addr))
	}

	for i := 0; i < 3; i++ {
		go procMail()
	}
	go readFromQueue(datasource)

	var clientId int64
	clientId = 1
	for {
		conn, err := listener.Accept()
		if err != nil {
			logln(1, fmt.Sprintf("Accept error: %s\n", err))
			continue
		}

		logln(1, conn.RemoteAddr().String())

		sem <- 1 // Wait for active queue to drain.
		client := &Client{
			conn:        conn,
			address:     conn.RemoteAddr().String(),
			time:        time.Now().Unix(),
			bufin:       bufio.NewReader(conn),
			bufout:      bufio.NewWriter(conn),
			clientId:    clientId,
			savedNotify: make(chan int),
			auth:        false,
		}
		go handleClient(client, datasource)
		clientId++
	}
}

func (c *redisClient) redisConnection() (err error) {
	if c.count > 100 {
		c.conn.Close()
		c.count = 0
	}
	if c.count == 0 {
		c.conn, err = redis.Dial("tcp", ":6379")
		if err != nil {
			// handle error
			return err
		}
	}
	return nil
}

func isAllowedHost(host string) bool {
	if allowed := allowedHosts[host]; !allowed {
		return false
	}
	return true
}

func sendMsg(ns *net.MX, msg ClientMessage) {
	err := smtp.SendMail(ns.Host+":25", nil, msg.From, []string{msg.To}, []byte(msg.Data))
	if err == nil {
		logln(1, "successfully send message")
	} else {
		logln(1, fmt.Sprintf("error in send message: %s", err))
	}
}

func procMsg(msg ClientMessage, datasource *datasource.DataSource) {
	rcpt := make([]string, 0, 100)

	toHost := strings.Split(msg.To, "@")[1]
	if isAllowedHost(toHost) {
		user, err := datasource.UserByEmail(msg.To)
		if err == nil {
			rcpt = append(rcpt, user.InboxAddr)
		} else {
			group, err := datasource.GroupByEmail(msg.To)
			if err == nil {
				for _, member := range group.Members {
					user, err = datasource.UserByEmail(member)
					if err == nil {
						rcpt = append(rcpt, user.InboxAddr)
					}
				}
			} else {
				logln(1, "Can't find such user or group")
				return
			}
		}
	}

	logln(1, fmt.Sprintf("%s", msg.From))
	fromHost := strings.Split(msg.From, "@")[1]
	if isAllowedHost(fromHost) {
		if msg.Auth == false || msg.Username != msg.From {
			logln(1, fmt.Sprintf("%s %s", msg.Username, msg.From))
			logln(1, "Not authenticated")
			return
		}
	}

	for _, email := range rcpt {
		logln(1, email)
		host := strings.Split(email, "@")[1]

		nss, err := net.LookupMX(host)
		if err == nil {
			for _, ns := range nss {
				logln(1, fmt.Sprintf("%s %d", ns.Host, ns.Pref))
			}
			curMsg := msg
			curMsg.To = email
			sendMsg(nss[0], curMsg)
		} else {
			logln(1, "Error in lookup MX")
		}
	}
}

func readFromQueue(datasource *datasource.DataSource) {
	for {
		r := &redisClient{}
		redis_err := r.redisConnection()
		if redis_err == nil {
			values, do_err := redis.Strings(r.conn.Do("LRANGE", "email_queue", 0, 1))
			if do_err == nil {
				if len(values) > 0 {
					var msg ClientMessage
					json.Unmarshal([]byte(values[0]), &msg)
					procMsg(msg, datasource)
				}
				_, do_err = redis.Strings(r.conn.Do("LPOP", "email_queue"))
			} else {
				logln(1, fmt.Sprintf("Redis do error %s", do_err))
			}
		} else {
			logln(1, "redis connection error")
		}
		time.Sleep(3000 * time.Millisecond)
	}
}

func procMail() {
	var to string

	for {
		client := <-SaveMailChan

		if user, _, addr_err := validateEmailData(client); addr_err != nil { // user, host, addr_err
			logln(1, fmt.Sprintln("mail_from didnt validate: %v", addr_err)+" client.mail_from: "+client.mail_from)
			// notify client that a save completed, -1 = error
			client.savedNotify <- -1
			continue
		} else {
			to = user + "@" + gConfig["GM_PRIMARY_MAIL_HOST"]
		}

		logln(1, to)
		logln(1, client.data)

		length := len(client.data)
		client.subject = mimeHeaderDecode(client.subject)
		client.hash = md5hex(to + client.mail_from + client.subject + strconv.FormatInt(time.Now().UnixNano(), 10))

		redis := &redisClient{}
		redis_err := redis.redisConnection()
		if redis_err == nil {
			msg := &ClientMessage{
				From:     client.mail_from,
				To:       client.rcpt_to,
				Auth:     client.auth,
				Data:     client.data,
				Subject:  client.subject,
				Username: client.username,
			}
			tmp, _ := json.Marshal(msg)
			_, do_err := redis.conn.Do("RPUSH", "email_queue", tmp)

			if do_err == nil {
				logln(1, "Email saved "+client.hash+" len: "+strconv.Itoa(length))
				client.savedNotify <- 1
			} else {
				logln(1, "Redis do error")
				client.savedNotify <- -1
			}
		} else {
			logln(1, "redis connection error")
			client.savedNotify <- -1
		}
	}
}

func clientAuth(client *Client, datasource *datasource.DataSource) string {
	succ := "235 Authentication succeeded"
	fail := "535 Authentication failed"

	user, err := datasource.UserByEmail(client.username)
	if err != nil {
		return fail
	}

	logln(1, user.Email)
	logln(1, client.password)
	if user.AcceptsPassword(client.password, datasource.ConfigByteArray("PASSWORD_SALT")) {
		client.auth = true
		return succ
	}
	return fail
}

func handleClient(client *Client, datasource *datasource.DataSource) {
	defer closeClient(client)
	//	defer closeClient(client)
	greeting := "220 " + gConfig["GSMTP_HOST_NAME"] +
		" SMTP Bahram-SMTPd #" + strconv.FormatInt(client.clientId, 10) + " (" + strconv.Itoa(len(sem)) + ") " + time.Now().Format(time.RFC1123Z)
	advertiseTls := "250-STARTTLS\r\n"
	passInput := false
	for i := 0; i < 100; i++ {
		switch client.state {
		case 0:
			responseAdd(client, greeting)
			client.state = 1
		case 1:
			input, err := readSmtp(client)
			if err != nil {
				logln(1, fmt.Sprintf("Read error: %v", err))
				if err == io.EOF {
					// client closed the connection already
					return
				}
				if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
					// too slow, timeout
					return
				}
				break
			}
			input = strings.Trim(input, " \n\r")
			cmd := strings.ToUpper(input)
			switch {
			case passInput:
				dec, err := base64.StdEncoding.DecodeString(input)
				if err != nil {
					return
				}
				client.password = string(dec)
				passInput = false
				resp := clientAuth(client, datasource)
				responseAdd(client, resp)

			case strings.Index(cmd, "HELO") == 0:
				if len(input) > 5 {
					client.helo = input[5:]
				}
				responseAdd(client, "250 "+gConfig["GSMTP_HOST_NAME"]+" Hello ")
			case strings.Index(cmd, "EHLO") == 0:
				if len(input) > 5 {
					client.helo = input[5:]
				}
				responseAdd(client, "250-"+gConfig["GSMTP_HOST_NAME"]+" Hello "+client.helo+"["+client.address+"]"+"\r\n"+"250-SIZE "+gConfig["GSMTP_MAX_SIZE"]+"\r\n"+advertiseTls+"250-AUTH LOGIN\r\n"+"250 HELP")

			case strings.Index(cmd, "MAIL FROM:") == 0:
				if len(input) > 10 {
					client.mail_from = input[10:]
				}
				responseAdd(client, "250 Ok")
			case strings.Index(cmd, "XCLIENT") == 0:
				// Nginx sends this
				// XCLIENT ADDR=212.96.64.216 NAME=[UNAVAILABLE]
				client.address = input[13:]
				client.address = client.address[0:strings.Index(client.address, " ")]
				fmt.Println("client address:[" + client.address + "]")
				responseAdd(client, "250 OK")
			case strings.Index(cmd, "RCPT TO:") == 0:
				if len(input) > 8 {
					client.rcpt_to = input[8:]
				}
				responseAdd(client, "250 Accepted")
			case strings.Index(cmd, "NOOP") == 0:
				responseAdd(client, "250 OK")
			case strings.Index(cmd, "RSET") == 0:
				client.mail_from = ""
				client.rcpt_to = ""
				responseAdd(client, "250 OK")
			case strings.Index(cmd, "DATA") == 0:
				responseAdd(client, "354 Enter message, ending with \".\" on a line by itself")
				client.state = 2
			case (strings.Index(cmd, "STARTTLS") == 0) && !client.tls_on:
				responseAdd(client, "220 Ready to start TLS")
				// go to start TLS state
				client.state = 3
			case strings.Index(cmd, "AUTH LOGIN") == 0:
				tmp := input[11:]
				dec, err := base64.StdEncoding.DecodeString(tmp)
				if err != nil {
					return
				}
				passInput = true
				client.username = string(dec)
				responseAdd(client, "334 UGFzc3dvcmQ6")
			case strings.Index(cmd, "QUIT") == 0:
				responseAdd(client, "221 Bye")
				killClient(client)
			default:
				responseAdd(client, fmt.Sprintf("500 unrecognized command"))
				client.errors++
				if client.errors > 3 {
					responseAdd(client, fmt.Sprintf("500 Too many unrecognized commands"))
					killClient(client)
				}
			}
		case 2:
			var err error
			client.data, err = readSmtp(client)
			if err == nil {
				// to do: timeout when adding to SaveMailChan
				// place on the channel so that one of the save mail workers can pick it up
				SaveMailChan <- client
				// wait for the save to complete
				status := <-client.savedNotify

				if status == 1 {
					responseAdd(client, "250 OK : queued as "+client.hash)
				} else {
					responseAdd(client, "554 Error: transaction failed, blame it on the weather")
				}
			} else {
				logln(1, fmt.Sprintf("DATA read error: %v", err))
			}
			client.state = 1
		case 3:
			// upgrade to TLS
			var tlsConn *tls.Conn
			tlsConn = tls.Server(client.conn, TLSconfig)
			err := tlsConn.Handshake() // not necessary to call here, but might as well
			if err == nil {
				client.conn = net.Conn(tlsConn)
				client.bufin = bufio.NewReader(client.conn)
				client.bufout = bufio.NewWriter(client.conn)
				client.tls_on = true
			} else {
				logln(1, fmt.Sprintf("Could not TLS handshake:%v", err))
			}
			advertiseTls = ""
			client.state = 1
		}
		// Send a response back to the client
		err := responseWrite(client)
		if err != nil {
			if err == io.EOF {
				// client closed the connection already
				return
			}
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				// too slow, timeout
				return
			}
		}
		if client.kill_time > 1 {
			return
		}
	}

}

func responseAdd(client *Client, line string) {
	client.response = line + "\r\n"
}
func closeClient(client *Client) {
	client.conn.Close()
	<-sem // Done; enable next client to run.
}
func killClient(client *Client) {
	client.kill_time = time.Now().Unix()
}

func readSmtp(client *Client) (input string, err error) {
	var reply string
	// Command state terminator by default
	suffix := "\r\n"
	if client.state == 2 {
		// DATA state
		suffix = "\r\n.\r\n"
	}
	for err == nil {
		client.conn.SetDeadline(time.Now().Add(timeout * time.Second))
		reply, err = client.bufin.ReadString('\n')
		if reply != "" {
			input = input + reply
			if len(input) > max_size {
				err = errors.New("Maximum DATA size exceeded (" + strconv.Itoa(max_size) + ")")
				return input, err
			}
			if client.state == 2 {
				// Extract the subject while we are at it.
				scanSubject(client, reply)
			}
		}
		if err != nil {
			break
		}
		if strings.HasSuffix(input, suffix) {
			break
		}
	}
	return input, err
}

// Scan the data part for a Subject line. Can be a multi-line
func scanSubject(client *Client, reply string) {
	if client.subject == "" && (len(reply) > 8) {
		test := strings.ToUpper(reply[0:9])
		if i := strings.Index(test, "SUBJECT: "); i == 0 {
			// first line with \r\n
			client.subject = reply[9:]
		}
	} else if strings.HasSuffix(client.subject, "\r\n") {
		// chop off the \r\n
		client.subject = client.subject[0 : len(client.subject)-2]
		if (strings.HasPrefix(reply, " ")) || (strings.HasPrefix(reply, "\t")) {
			// subject is multi-line
			client.subject = client.subject + reply[1:]
		}
	}
}

func responseWrite(client *Client) (err error) {
	var size int
	client.conn.SetDeadline(time.Now().Add(timeout * time.Second))
	size, err = client.bufout.WriteString(client.response)
	client.bufout.Flush()
	client.response = client.response[size:]
	return err
}

func md5hex(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	sum := h.Sum([]byte{})
	return hex.EncodeToString(sum)
}

// Decode strings in Mime header format
// eg. =?ISO-2022-JP?B?GyRCIVo9dztSOWJAOCVBJWMbKEI=?=
func mimeHeaderDecode(str string) string {
	reg, _ := regexp.Compile(`=\?(.+?)\?([QBqp])\?(.+?)\?=`)
	matched := reg.FindAllStringSubmatch(str, -1)
	var charset, encoding, payload string
	if matched != nil {
		for i := 0; i < len(matched); i++ {
			if len(matched[i]) > 2 {
				charset = matched[i][1]
				encoding = strings.ToUpper(matched[i][2])
				payload = matched[i][3]
				switch encoding {
				case "B":
					str = strings.Replace(str, matched[i][0], mailTransportDecode(payload, "base64", charset), 1)
				case "Q":
					str = strings.Replace(str, matched[i][0], mailTransportDecode(payload, "quoted-printable", charset), 1)
				}
			}
		}
	}
	return str
}

func validateEmailData(client *Client) (user string, host string, addr_err error) {
	if user, host, addr_err = extractEmail(client.mail_from); addr_err != nil {
		return user, host, addr_err
	}
	client.mail_from = user + "@" + host
	if user, host, addr_err = extractEmail(client.rcpt_to); addr_err != nil {
		return user, host, addr_err
	}
	client.rcpt_to = user + "@" + host
	// check if on allowed hosts
	/*if allowed := allowedHosts[host]; !allowed {
		return user, host, errors.New("invalid host:" + host)
	}*/
	return user, host, addr_err
}

func extractEmail(str string) (name string, host string, err error) {
	re, _ := regexp.Compile(`<(.+?)@(.+?)>`) // go home regex, you're drunk!
	if matched := re.FindStringSubmatch(str); len(matched) > 2 {
		host = validHost(matched[2])
		name = matched[1]
	} else {
		if res := strings.Split(str, "@"); len(res) > 1 {
			name = res[0]
			host = validHost(res[1])
		}
	}
	if host == "" || name == "" {
		err = errors.New("Invalid address, [" + name + "@" + host + "] address:" + str)
	}
	return name, host, err
}

func mailTransportDecode(str string, encoding_type string, charset string) string {
	if charset == "" {
		charset = "UTF-8"
	} else {
		charset = strings.ToUpper(charset)
	}
	if encoding_type == "base64" {
		str = fromBase64(str)
	} else if encoding_type == "quoted-printable" {
		str = fromQuotedP(str)
	}
	if charset != "UTF-8" {
		charset = fixCharset(charset)
		// eg. charset can be "ISO-2022-JP"
		convstr, err := iconv.Conv(str, "UTF-8", charset)
		if err == nil {
			return convstr
		}
	}
	return str
}

func fixCharset(charset string) string {
	reg, _ := regexp.Compile(`[_:.\/\\]`)
	fixed_charset := reg.ReplaceAllString(charset, "-")
	// Fix charset
	// borrowed from http://squirrelmail.svn.sourceforge.net/viewvc/squirrelmail/trunk/squirrelmail/include/languages.php?revision=13765&view=markup
	// OE ks_c_5601_1987 > cp949
	fixed_charset = strings.Replace(fixed_charset, "ks-c-5601-1987", "cp949", -1)
	// Moz x-euc-tw > euc-tw
	fixed_charset = strings.Replace(fixed_charset, "x-euc", "euc", -1)
	// Moz x-windows-949 > cp949
	fixed_charset = strings.Replace(fixed_charset, "x-windows_", "cp", -1)
	// windows-125x and cp125x charsets
	fixed_charset = strings.Replace(fixed_charset, "windows-", "cp", -1)
	// ibm > cp
	fixed_charset = strings.Replace(fixed_charset, "ibm", "cp", -1)
	// iso-8859-8-i -> iso-8859-8
	fixed_charset = strings.Replace(fixed_charset, "iso-8859-8-i", "iso-8859-8", -1)
	if charset != fixed_charset {
		return fixed_charset
	}
	return charset
}

func validHost(host string) string {
	host = strings.Trim(host, " ")
	re, _ := regexp.Compile(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)
	if re.MatchString(host) {
		return host
	}
	return ""
}

func fromBase64(data string) string {
	buf := bytes.NewBufferString(data)
	decoder := base64.NewDecoder(base64.StdEncoding, buf)
	res, _ := ioutil.ReadAll(decoder)
	return string(res)
}

func fromQuotedP(data string) string {
	buf := bytes.NewBufferString(data)
	decoder := qprintable.NewDecoder(qprintable.BinaryEncoding, buf)
	res, _ := ioutil.ReadAll(decoder)
	return string(res)
}
