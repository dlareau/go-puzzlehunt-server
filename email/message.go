package email

import "bytes"
import "errors"
import "fmt"
import "io/ioutil"
import "mime/multipart"
import "net/mail"
// import "net/smtp"
import "strings"
import "time"

type Message struct {
  From    mail.Address
  To      []mail.Address
  Cc      []mail.Address
  Bcc     []mail.Address
  Headers map[string]string
  Subject string
  Content string
}

const crlf = "\r\n"

func addresses(buf *bytes.Buffer, name string, values []mail.Address) {
  if len(values) == 0 {
    return
  }
  buf.WriteString(name + ": ")
  for i, address := range values {
    if i > 0 {
      buf.WriteString(", ")
    }
    buf.WriteString(address.String())
  }
  buf.WriteString(crlf)
}

func (m *Message) Bytes() []byte {
  var buf bytes.Buffer

  addresses(&buf, "From", []mail.Address{m.From})
  addresses(&buf, "To", m.To)
  addresses(&buf, "Cc", m.Cc)
  addresses(&buf, "Bcc", m.Bcc)
  m.Headers["Content-Type"] = "text/plain; charset=utf-8"
  m.Headers["Date"] = time.Now().Format("Mon, 2 Sep 2006 15:04:05 -0700 (MST)")

  for k, v := range(m.Headers) {
    fmt.Fprintf(&buf, "%s: %s%s", k, v, crlf)
  }

  fmt.Fprintf(&buf, "Subject: %s%s%s", m.Subject, crlf, crlf)
  buf.WriteString(m.Content)
  return buf.Bytes()
}

func (m *Message) Send(server string) error {
  to := make([]string, 0)
  for _, addr := range m.To {
    to = append(to, addr.Address)
  }
  // return smtp.SendMail(server, nil, m.From.Address, to, m.Bytes())
  return nil
}

func Plaintext(m *mail.Message) ([]byte, error) {
  typ := m.Header.Get("Content-Type")
  if strings.HasPrefix(typ, "text/plain") {
    return ioutil.ReadAll(m.Body)
  }
  if !strings.HasPrefix(typ, "multipart/alternative") {
    return []byte{}, errors.New("Unknown content-type")
  }
  parts := strings.SplitN(typ, "=", 2)
  if len(parts) != 2 {
    return []byte{}, errors.New("Unknown boundary for multipart")
  }

  r := multipart.NewReader(m.Body, parts[1])

  for {
    part, err := r.NextPart()
    if err != nil {
      return []byte{}, err
    }
    typ := part.Header.Get("Content-Type")
    if strings.HasPrefix(typ, "text/plain") {
      return ioutil.ReadAll(part)
    }
  }

  return []byte{}, errors.New("No plaintext part to the message")
}
